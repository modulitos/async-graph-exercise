package solver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
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

	childChans := make([]chan int, len(data.Children))
	for i, nodeId := range data.Children {
		childChans[i] = make(chan int)
		// unsafe cast from string to byte:
		go crawlNode(GetUriForNode(nodeId[0]), childChans[i], errs)
	}

	totalReward := data.Reward

	wg := new(sync.WaitGroup)
	wg.Add(len(data.Children))
	for i := range childChans {
		go func(childChan chan int) {
			defer wg.Done()
			select {
			case err := <-errs:
				errs <- err
				return
			case reward := <-childChan:
				totalReward += reward
			}
		}(childChans[i])

	}
	wg.Wait()

	ch <- totalReward
}

func CalculateReward(nodeId byte) (int, error) {
	ch := make(chan int, 1)
	errs := make(chan error, 1)

	go crawlNode(GetUriForNode(nodeId), ch, errs)


	select {
	// What if there are both errors and results?
	case err := <-errs:
		return -1, err
	case reward := <-ch:
		return reward, nil
	}
}
