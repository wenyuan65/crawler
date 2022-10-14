package main

import (
	"fmt"
	"sync"
)

func main() {
	pool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("creating new instance")
			return struct{}{}
		},
	}

	pool.Get()
	a := pool.Get()
	pool.Put(a)
	fmt.Println("-------1----------")
	pool.Get()
	fmt.Println("-------2------------")
	pool.Get()
	fmt.Println("-------3------------")
	pool.Get()

	// var a = []int{1, 2, 3, 4, 5}
	// var b = []int{1, 2, 3, 4, 6}
	// var c = []int{11, 21, 32, 4, 7}

	// i := utils.Distinct(a, b, c)
	// fmt.Println(i)
}
