package solver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type NodeJSON struct {
	// TODO: Is there no way to enforce that these fields are required upon
	// unmarshalling?
	Children []string `json:"children"`
	Reward   int      `json:"reward"`
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HttpClient
)

func init() {
	Client = &http.Client{}
}

func GetUriForNode(nodeId byte) string {
	return fmt.Sprintf("https://graph.modulitos.com/node/%s", string(nodeId))
}

func crawlNode(nodeUri string, ch chan int, errs chan error) {
	// TODO: leverage a cache that checks whether a node has already been
	// visited.
	defer close(ch)

	request, err := http.NewRequest(http.MethodGet, nodeUri, nil)
	resp, err := Client.Do(request)
	if err != nil {
		errs <- err
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var data NodeJSON
	if err := json.Unmarshal(body, &data); err != nil {
		errs <- err
		return
	}

	chans := make([]chan int, len(data.Children))
	fmt.Println("children: ", data.Children)
	for i, nodeUri := range data.Children {
		chans[i] = make(chan int)
		go crawlNode(nodeUri, chans[i], errs)
	}

	totalReward := data.Reward

	for i := range chans {

		// TODO: don't do this in band!
		select {
		case err := <-errs:
			errs <- err
			return
		case reward := <-chans[i]:
			totalReward += reward
		}
	}

	ch <- totalReward
}

func CalculateReward(nodeId byte) (int, error) {
	ch := make(chan int, 1)
	errs := make(chan error, 1)

	go crawlNode(GetUriForNode(nodeId), ch, errs)


	select {
	case err := <-errs:
		return -1, err
	case reward := <-ch:
		return reward, nil
	}
}
