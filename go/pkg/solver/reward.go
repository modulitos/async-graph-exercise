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

type NodeFormatter struct {
	baseUrl string
}

func NewNodeFormatter() NodeFormatter {
	return NodeFormatter{
		baseUrl: "https://graph.modulitos.com/node/",
	}
}

func (nf *NodeFormatter) GetUriForNode(nodeId byte) string {
	return fmt.Sprintf(nf.baseUrl+"%s", string(nodeId))
}

type NodeService struct {
	client    HttpClient
	formatter NodeFormatter
	cache     map[byte]NodeJSON
	mu        sync.Mutex
}

func (ns *NodeService) GetNode(nodeId byte) (NodeJSON, error) {
	// Use node from cache, if possible:
	ns.mu.Lock()
	value, cacheContainsKey := ns.cache[nodeId]
	ns.mu.Unlock()
	if cacheContainsKey {
		return value, nil
	}

	nodeUri := ns.formatter.GetUriForNode(nodeId)
	request, err := http.NewRequest(http.MethodGet, nodeUri, nil)
	resp, err := Client.Do(request)
	if err != nil {
		return NodeJSON{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var data NodeJSON
	if err := json.Unmarshal(body, &data); err != nil {
		return NodeJSON{}, err
	}

	// Update the cache:
	ns.mu.Lock()
	ns.cache[nodeId] = data
	ns.mu.Unlock()

	return data, nil
}

func NewNodeService() NodeService {
	return NodeService{
		client:    Client,
		formatter: NewNodeFormatter(),
		cache:     map[byte]NodeJSON{},
	}
}

func crawlNode(nodeId byte, ch chan int, errs chan error, nodeService *NodeService) {
	defer close(ch)

	data, err := nodeService.GetNode(nodeId)

	if err != nil {
		errs <- err
		return
	}

	childChans := make([]chan int, len(data.Children))
	for i, nodeId := range data.Children {
		childChans[i] = make(chan int)
		go crawlNode(nodeId[0], childChans[i], errs, nodeService)
	}

	totalReward := data.Reward

	// Instead of wait group, maybe we instead use a channel:
	// https://talks.golang.org/2012/10things.slide#8
	// although WaitGroup is probably cleaner.
	// done := make(chan struct{})
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

	nodeService := NewNodeService()

	go crawlNode(nodeId, ch, errs, &nodeService)

	select {
	// What if there are both errors and results?
	case err := <-errs:
		return -1, err
	case reward := <-ch:
		return reward, nil
	}
}
