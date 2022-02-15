package ticker

import (
	"fmt"
	"time"
)

func Ticker() {
	t := time.NewTicker(time.Second)

	isOK := true
	for isOK {
		select {
		case <-t.C:
			fmt.Println(time.Now())
			isOK = false
		}
	}

}
