package raid6

import (
	"fmt"
	"strings"
)

type raid6 struct {
	dataShards     int
	parityShards   int
	totalShards    int
	encodingMatrix matrix
	diskArray      matrix
	dataDisk       matrix
	parityDisk     matrix
}

func buildMatrix(rows, cols int) matrix {
	m := matrix(make([][]byte, rows))
	for i := range m {
		m[i] = make([]byte, cols)
	}
	return m
}

func fixedVandermond(rows, cols int) matrix {
	// Generate a fixed Vandermonde matrix based on
	// https://web.eecs.utk.edu/~jplank/plank/papers/CS-03-504.html
	result := buildMatrix(rows, cols)

	for r, row := range result {
		for c := range row {
			result[r][c] = galExp(byte(r), c)
		}
	}

	// Perform Gausian elimination to transform data part into an identity matrix
	// Skip r == 0 because that is already identity row.
	for r := 1; r < cols; r++ {
		factor := result[r][r]
		// first make result[r][r] == 1 by dividing the column by result[r][r]
		if factor != 1 {
			for r1 := r; r1 < rows; r1++ {
				result[r1][r] = galDivide(result[r1][r], factor)
			}
		}

		// then subtract this row from all other rows to make all other entries in this column 0
		for c1 := 0; c1 < cols; c1++ {
			if c1 == r {
				continue
			} else {
				multiplier := result[r][c1]
				for r1 := r; r1 < rows; r1++ {
					// Subtract rows by applying bitwise XOR
					result[r1][c1] ^= galMultiply(multiplier, result[r1][r])
				}
			}
		}
	}
	fmt.Printf("Encoding matrix: %d \n", result)
	return result
}

func BuildRaidSystem(dataShards, parityShards int) *raid6 {
	r := raid6{
		dataShards:   dataShards,
		parityShards: parityShards,
		totalShards:  dataShards + parityShards,
	}

	r.encodingMatrix = fixedVandermond(r.totalShards, r.dataShards)
	r.diskArray = buildMatrix(r.totalShards, 5000)
	r.dataDisk = r.diskArray[:r.dataShards]
	r.parityDisk = r.diskArray[r.dataShards:]

	fmt.Printf("Build Disk Array: %d, %d \n", len(r.diskArray), len(r.diskArray[0]))
	fmt.Printf("Build Data Disk: %d, %d \n", len(r.dataDisk), len(r.dataDisk[0]))
	fmt.Printf("Build Parity Disk: %d, %d \n", len(r.parityDisk), len(r.parityDisk[0]))

	return &r
}

func (r *raid6) Encode(shards [][]byte) {
	// We perform encoding when save data to disk
	// encoding_matrix(n+m, n) * data_matrix(n,n) = data_shard(n+m,n)
	// [   identity matrix  ]       [      ]           [  data  ]
	// [--------------------]  *    [ data ]      =    [--------]
	// [ vandermonde matrix ]       [      ]           [ parity ]

	r.dataDisk, _ = r.encodingMatrix.Multiply(shards)
	fmt.Printf("Data Disk:\n")
	for _, row := range r.dataDisk {
		for _, val := range row {
			fmt.Print(string(val), "\t")
		}
		fmt.Println()
	}
}

// Split splits the input string into shards of equal length,
// and binarize each character in the string.
// For simplicity, this implementation only takes string as input.
func (r *raid6) Split(dataString string) ([][]byte, int) {
	length := len(dataString)
	paddingLength := 0
	if length%r.dataShards != 0 {
		paddingLength = r.dataShards - length%r.dataShards
		dataString += strings.Repeat("0", paddingLength)
	}
	fmt.Printf("Padded string: %s \n", dataString)
	totalStringLength := length + paddingLength
	splitStringLength := totalStringLength / r.dataShards

	shards := make([][]byte, r.dataShards)
	for i := 0; i < totalStringLength; i += splitStringLength {
		end := i + splitStringLength
		shards[i/splitStringLength] = []byte(dataString[i:end])
	}

	fmt.Printf("Split shards: %d, %d \n", len(shards), len(shards[0]))

	return shards, length
}

func (r *raid6) Join(shards [][]byte, length int) string {
	var outputSlice []string

	for _, v := range shards {
		outputSlice = append(outputSlice, string(v))
	}
	outputString := strings.Join(outputSlice, "")
	return outputString[:length]
}
