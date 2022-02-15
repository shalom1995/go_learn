package main

import (
	"go_learn/ticker定时器"
)

func main() {
	go ticker.Ticker()
	for{}
}
