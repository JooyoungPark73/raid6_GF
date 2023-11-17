package raid6

import "fmt"

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
}
