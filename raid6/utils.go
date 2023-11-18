package raid6

import (
	"fmt"
	"os"
)

func StrToBin(s string) (binString string) {
	for _, c := range s {
		binString = fmt.Sprintf("%s%b", binString, c)
	}
	return
}

func Print2DArray(name string, array [][]byte) {
	fmt.Printf("%s:\n", name)
	for _, row := range array {
		for _, val := range row {
			fmt.Printf("%x \t", val)
		}
		fmt.Println()
	}
	fmt.Println()
}

func PrintDiskHex(name string, array [][]byte) {
	fmt.Printf("%s:\n", name)
	for i, row := range array {
		fmt.Printf("Disk %d \t", i)
		for _, val := range row {
			fmt.Printf("%x \t", val)
		}
		fmt.Println()
	}
	fmt.Println()
}

func (r *raid6) PrintDiskString(name string, array [][]byte) {
	fmt.Printf("%s:\n", name)
	for i, row := range array {
		fmt.Printf("Disk %d: \t", i)
		for _, val := range row {
			if i < r.dataShards {
				fmt.Printf("%s  ", string(val))
			} else {
				fmt.Printf("%x ", val)
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(2)
	}
}
