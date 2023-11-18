package main

import (
	"fmt"

	"raid6/raid6"
)

func main() {
	r, err := raid6.BuildRaidSystem(5, 5)
	raid6.CheckErr(err)

	data_string := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	fmt.Printf("Saved Output: \n%s \n\n", data_string)
	shards, length := r.Split(data_string)

	r.Encode(shards)
	r.PrintDiskString("Clean Disk Array", r.DiskArray)

	err = r.DropShard(2)
	raid6.CheckErr(err)
	err = r.DropShard(3)
	raid6.CheckErr(err)
	err = r.DropShard(4)
	raid6.CheckErr(err)
	err = r.DropShard(7)
	raid6.CheckErr(err)
	err = r.DropShard(8)
	raid6.CheckErr(err)
	r.PrintDiskString("Erasure Disk Array", r.DiskArray)
	corrupt_output := r.Join(r.DiskArray, length)
	fmt.Printf("Corrupted Output: \n%s \n\n", corrupt_output)
	err = r.ReconstructDisk()
	raid6.CheckErr(err)

	r.PrintDiskString("Reconstructed Disk Array", r.DiskArray)
	output := r.Join(r.DiskArray, length)
	fmt.Printf("Recovered Output: \n%s \n\n", output)

	err = r.CreateBitFlip(6, 1)
	raid6.CheckErr(err)
	r.PrintDiskString("Corrupt Disk Array", r.DiskArray)
	err = r.ReconstructCorruption()
	r.PrintDiskString("Reconstructed Disk Array", r.DiskArray)
	raid6.CheckErr(err)

	err = r.CreateBitFlip(2, 1)
	r.PrintDiskString("Corrupt Disk Array", r.DiskArray)
	raid6.CheckErr(err)
	err = r.ReconstructCorruption()
	raid6.CheckErr(err)

}
