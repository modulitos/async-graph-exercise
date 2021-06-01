package main

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

func GetUrlForNode(nodeId byte) string {
	return fmt.Sprintf("https://graph.modulitos.com/node/%b", nodeId)
}

func CalculateReward(url string) (int, error) {
	// https://blog.golang.org/json
	// https://medium.com/@fsufitch/deserializing-json-in-go-a-tutorial-d042412958ea
	request, err := http.NewRequest(http.MethodGet, url, nil)
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

func main() {
	CalculateReward(GetUrlForNode('a'))
}
