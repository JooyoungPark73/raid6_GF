package main

import (
	"fmt"

	"raid6/raid6"
)

func main() {
	r := raid6.BuildRaidSystem(6, 2)
	fmt.Println()

	data_string := "Singapore is a fine city!"
	shards, length := r.Split(data_string)
	fmt.Printf("Shards:\n")
	for _, row := range shards {
		for _, val := range row {
			fmt.Print(string(val), "\t")
		}
		fmt.Println()
	}

	output := r.Join(shards, length)
	fmt.Printf("Output: %s \n", output)
	r.Encode(shards)

}
