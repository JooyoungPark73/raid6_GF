package raid6

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

type raid6 struct {
	dataShards     int
	parityShards   int
	totalShards    int
	encodingMatrix matrix
	DiskArray      matrix
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
	Print2DArray("Fixed Vandermonde Matrix", result)
	return result
}

func BuildRaidSystem(dataShards, parityShards int) *raid6 {
	r := raid6{
		dataShards:   dataShards,
		parityShards: parityShards,
		totalShards:  dataShards + parityShards,
	}

	r.encodingMatrix = fixedVandermond(r.totalShards, r.dataShards)
	r.DiskArray = buildMatrix(r.totalShards, 5000)

	fmt.Printf("Build Disk Array: %d, %d \n", len(r.DiskArray), len(r.DiskArray[0]))

	return &r
}

func (r *raid6) Encode(shards [][]byte) {
	// We perform encoding when save data to disk
	// encoding_matrix(n+m, n) * data_matrix(n,n) = data_shard(n+m,n)
	// [   identity matrix  ]       [      ]           [  data  ]
	// [--------------------]  *    [ data ]      =    [--------]
	// [ vandermonde matrix ]       [      ]           [ parity ]

	r.DiskArray, _ = r.encodingMatrix.Multiply(shards)
	Print2DArray("Encoded Data Disk", r.DiskArray)
}

// Verify assumes error detected is the result of a bit flip
// This function cannot detect erasure.
// To detect erasure, we need to be notified which disk is corrupted.
func (r *raid6) Verify() []bool {
	// Compare each shard with encoded matrix of matrix to verify

	calculated_shard, _ := r.encodingMatrix.Multiply(r.DiskArray[:r.dataShards])
	calculated_shard = calculated_shard[r.dataShards:]
	output := make([]bool, r.parityShards)

	Print2DArray("Calculated Shard", calculated_shard)
	Print2DArray("Disk Array", r.DiskArray)

	for i, calculated := range calculated_shard {
		if !bytes.Equal(calculated, r.DiskArray[i+r.dataShards]) {
			output[i] = false
		} else {
			output[i] = true
		}
	}
	return output
}

func (r *raid6) DetectBrokenDisk() []bool {
	validDiskList := make([]bool, r.totalShards)
	for i := 0; i < r.totalShards; i++ {
		if r.DiskArray[i] != nil {
			validDiskList[i] = true
		} else {
			validDiskList[i] = false
		}
	}
	return validDiskList
}

func (r *raid6) Reconstruct(validDisks []bool) {
	// Reconstruct data from parity shards
	// First, we need to find the inverse of the encoding matrix
	// We can do this by performing Gaussian elimination
	// on the encoding matrix augmented with the identity matrix
	// [ encoding_matrix(n+m, n) | identity_matrix(n+m, n) ] -> [ identity_matrix(n+m, n) | inverse_encoding_matrix(n+m, n) ]

}

func (r *raid6) CreateBitFlip(shards [][]byte, nShard int, nBit int) [][]byte {
	// Create error in a specific shard
	if shards[nShard] != nil {
		shards[nShard][nBit] ^= 1
		Print2DArray("CreateBitFlip", shards)
	} else {
		fmt.Printf("Shard %d is nil \n", nShard)
	}
	return shards
}

func (r *raid6) DropShard(shards [][]byte, nShard int) [][]byte {
	// Create error in a specific shard
	shards[nShard] = nil
	Print2DArray("DropShard", shards)
	return shards
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
		if v == nil {
			outputSlice = append(outputSlice, strings.Repeat("/", int(math.Ceil(float64(length)/float64(r.dataShards)))))
		} else {
			outputSlice = append(outputSlice, string(v))
		}
	}
	outputString := strings.Join(outputSlice, "")
	return outputString[:length]
}
