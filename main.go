package main

import (
	"fmt"

	"raid6/raid6"
)

func main() {
	fmt.Println("Hello, World!")
	r := raid6.NewRaidSystem(4, 2)
	fmt.Println(r)
}
