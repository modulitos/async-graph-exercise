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

func crawlNode(nodeUri string, ch chan int) (int, error) {
	// TODO: leverage a cache that checks whether a node has already been
	// visited.
	defer close(ch)

	// https://blog.golang.org/json
	// https://medium.com/@fsufitch/deserializing-json-in-go-a-tutorial-d042412958ea
	request, err := http.NewRequest(http.MethodGet, nodeUri, nil)
	resp, err := Client.Do(request)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var data NodeJSON
	if err := json.Unmarshal(body, &data); err != nil {
		return -1, err
	}

	return data.Reward, nil
}

func CalculateReward(nodeId byte) (int, error) {
	ch := make(chan int)

	reward, err := crawlNode(GetUriForNode(nodeId), ch)

	<-ch
	return reward, err
}
