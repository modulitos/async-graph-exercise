package main

import (
	"fmt"
	"solution/pkg/solver"
)

func main() {
	println("starting...")
	var nodeId byte = 'a'
	reward, err := solver.CalculateReward(nodeId)

	if err != nil {
		panic(fmt.Sprintf("error! %s", err))
	} else {
		println("your reward is:", reward)
	}
}
