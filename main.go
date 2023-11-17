package main

import (
	"fmt"

	"raid6/raid6"
)

func main() {
	r := raid6.BuildRaidSystem(3, 3)
	fmt.Println()

	data_string := "abcdefghijklmnopqrstuvwxyz"
	shards, length := r.Split(data_string)
	raid6.Print2DArray("Shards", shards)

	r.Encode(shards)
	r.DropShard(r.DiskArray, 3)
	// r.CreateBitFlip(r.DiskArray, 4, 1)
	// checkResult := r.Verify()
	checkResult := r.DetectBrokenDisk()
	fmt.Printf("Check Result: %v \n", checkResult)

	output := r.Join(r.DiskArray, length)
	fmt.Printf("Output: %s \n", output)
}
