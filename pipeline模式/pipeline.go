package channel

import "fmt"

func echo(nums []int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}


func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}


func odd(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			if n%2 != 0 {
				out <- n
			}
		}
		close(out)
	}()
	return out
}


func sum(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		var sum = 0
		for n := range in {
			sum += n
		}
		out <- sum
		close(out)
	}()
	return out
}

func Run() {
	var nums = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for n := range sum(sq(odd(echo(nums)))) {
		fmt.Println(n)
	}
}

type echoFunc func(nums []int) <-chan int
type proxyFunc func(in <-chan int) <-chan int
func pipeline(nums []int,echo echoFunc,proxies ...proxyFunc) <-chan int{
	ch:=echo(nums)
	for i := range proxies {
		ch = proxies[i](ch)
	}
	return ch
}
func pipeline_(echoCh <-chan int,proxies ...proxyFunc) <-chan int{
	for i := range proxies {
		echoCh = proxies[i](echoCh)
	}
	return echoCh
}
func RunProxy() {
	var nums = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for n := range pipeline(nums,echo,sq,odd,sum) {
		fmt.Println(n)
	}
}