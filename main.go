package main

import (
	"go_learn/ticker"
)

func main() {
	go ticker.Ticker()
	for{}
}
