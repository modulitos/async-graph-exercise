package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type NodeJSON struct {
	Children []string `json:"children"`
	Reward   int      `json:"reward"`
}

func main() {

	// https://blog.golang.org/json
	// https://medium.com/@fsufitch/deserializing-json-in-go-a-tutorial-d042412958ea
	resp, err := http.Get("https://graph.modulitos.com/node/a")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Printf("body: %s\n", body)
	var data NodeJSON
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("failed to unmarshal:", err)
	} else {
		fmt.Println(data)
	}
}
