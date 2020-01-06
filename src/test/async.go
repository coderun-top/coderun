package main

import (
	"fmt"
//	"time"
)

func sum(s []int) {
	//time.Sleep(time.Duration(2)*time.Second)
	sum := 0
	for _, v := range s {
		sum += v
	}
	fmt.Println("1111111")
	//c <- sum // send sum to c
}

func main() {
	s := []int{7, 2, 8, -9, 4, 0}

	//c := make(chan int)
	go sum(s[:len(s)/2])
	go sum(s[len(s)/2:])
	//x, y := <-c, <-c // receive from c

	//fmt.Println(x, y, x+y)
	fmt.Println("hello world")

}
