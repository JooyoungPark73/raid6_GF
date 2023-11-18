package raid6

import (
	"bytes"
	"errors"
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

func fixedVandermond(rows, cols int) matrix {
	// Generate a fixed Vandermonde matrix based on
	// https://web.eecs.utk.edu/~jplank/plank/papers/CS-03-504.html
	result, _ := newMatrix(rows, cols)

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
	return result
}

func BuildRaidSystem(dataShards, parityShards int) (*raid6, error) {
	if dataShards <= 0 || parityShards <= 0 {
		return nil, errors.New("invalid data or parity shards")
	}

	r := raid6{
		dataShards:   dataShards,
		parityShards: parityShards,
		totalShards:  dataShards + parityShards,
	}

	r.encodingMatrix = fixedVandermond(r.totalShards, r.dataShards)
	r.DiskArray, _ = newMatrix(r.totalShards, 5000)

	fmt.Printf("Build Disk Array: %d, %d \n", len(r.DiskArray), len(r.DiskArray[0]))
	fmt.Printf("Data Shards: %d \n", r.dataShards)
	fmt.Printf("Parity Shards: %d \n", r.parityShards)
	fmt.Println()

	return &r, nil
}

func (r *raid6) Encode(shards [][]byte) {
	// We perform encoding when save data to disk
	// encoding_matrix(n+m, n) * data_matrix(n,n) = data_shard(n+m,n)
	// [   identity matrix  ]       [      ]           [  data  ]
	// [--------------------]  *    [ data ]      =    [--------]
	// [ vandermonde matrix ]       [      ]           [ parity ]

	r.DiskArray, _ = r.encodingMatrix.Multiply(shards)
}

func (r *raid6) Verify() ([]bool, matrix) {
	// Verify assumes error detected is the result of a bit flip
	// This function cannot detect erasure.
	// To detect erasure, we need to be notified which disk is corrupted.
	// Compare each shard with encoded matrix of matrix to verify

	calculated_shard, _ := r.encodingMatrix.Multiply(r.DiskArray[:r.dataShards])
	calculated_shard = calculated_shard[r.dataShards:]
	output := make([]bool, r.parityShards)

	for i, calculated := range calculated_shard {
		if r.DiskArray[i+r.dataShards] == nil {
			output[i] = false
		} else if !bytes.Equal(calculated, r.DiskArray[i+r.dataShards]) {
			output[i] = false
		} else {
			output[i] = true
		}
	}
	return output, calculated_shard
}

func (r *raid6) ReconstructCorruption() error {
	// Reconstruct parity from data shards
	validParityList, calculated_parity := r.Verify()

	nValidParity := 0
	for _, v := range validParityList {
		if v {
			nValidParity++
		}
	}

	if nValidParity == 0 {
		return errors.New("possible data disk corruption. cannot recover from data corruption")
	}

	for i, calculated := range calculated_parity {
		if !validParityList[i] {
			r.DiskArray[i+r.dataShards] = calculated
		}
	}

	return nil
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

func (r *raid6) ReconstructDisk() error {
	// Reconstruct data from parity shards

	validDisks := r.DetectBrokenDisk()
	if len(validDisks) < r.dataShards {
		return errors.New("invalid valid disk list")
	}

	nValidDataDisks := 0
	nValidParityDisks := 0
	for i, v := range validDisks {
		if v && i < r.dataShards {
			nValidDataDisks++
		} else if v && i >= r.dataShards {
			nValidParityDisks++
		}
	}

	if nValidDataDisks+nValidParityDisks < r.dataShards {
		// Not enough valid disks to reconstruct data
		return errors.New("not enough valid disks to reconstruct data")
	} else if nValidDataDisks < r.dataShards {
		// Data disk erasure detected
		// Will reconstruct Parity disk too

		err := r.ReconstructDataDisk(validDisks)
		if err != nil {
			return err
		}
		err = r.ReconstructCorruption()
		return err
	} else if nValidDataDisks == r.dataShards && nValidParityDisks < r.parityShards {
		// only Parity disk erasure detected
		err := r.ReconstructCorruption()
		return err
	} else if nValidDataDisks == r.dataShards && nValidParityDisks == r.parityShards {
		return nil
		// No disk erasure detected
	} else {
		return errors.New("invalid disk reconstruction condition")
	}
}

func (r *raid6) ReconstructDataDisk(validDisks []bool) error {
	// inverted_broken_encoding_matrix(n, n+m-b) * broken_data_shard(n+m,n) = data_matrix(n,n)
	//     [   inverted encoding matrix  ]       *    [  broken_data  ]     =    [ data ]

	// Pull intact disks from disk array, up to number of data shards
	// We need to generate a subshard that contains only the intact disks
	// Also build a square subEncodingMatrix that contains only the row with intact disks

	subShards := make([][]byte, r.dataShards)
	subEncodingMatrix, _ := newMatrix(r.dataShards, r.dataShards)
	subMatrixRow := 0
	for matrixRow := 0; matrixRow < r.totalShards && subMatrixRow < r.dataShards; matrixRow++ {
		if validDisks[matrixRow] {
			subShards[subMatrixRow] = r.DiskArray[matrixRow]
			subEncodingMatrix[subMatrixRow] = r.encodingMatrix[matrixRow]
			subMatrixRow++
		}
	}

	// Next, we need to find the inverse of the encoding matrix
	// We can do this by performing Gaussian elimination
	// on the encoding matrix augmented with the identity matrix
	// [ encoding_matrix | identity_matrix ] -> [ identity_matrix | inverse_encoding_matrix ]

	dataDecodeMatrix, err := subEncodingMatrix.Invert()
	if err != nil {
		return err
	}
	dataGenerated, err := dataDecodeMatrix.Multiply(subShards)
	if err != nil {
		return err
	}

	for i, v := range dataGenerated {
		r.DiskArray[i] = v
	}

	return nil
}

func (r *raid6) CreateBitFlip(nShard int, nBit int) error {
	// Create error in a specific shard
	if nShard >= len(r.DiskArray) || nBit >= len(r.DiskArray[nShard]) {
		err := errors.New("invalid shard number or bit number")
		return err
	}
	if r.DiskArray[nShard] != nil {
		r.DiskArray[nShard][nBit] ^= 1
		return nil
	} else {
		err := errors.New("cannot create bit flip in a nil shard")
		return err
	}
}

func (r *raid6) DropShard(nShard int) error {
	// Create error in a specific shard
	if nShard >= len(r.DiskArray) {
		err := errors.New("invalid shard number")
		return err
	}
	r.DiskArray[nShard] = nil
	return nil
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
	totalStringLength := length + paddingLength
	splitStringLength := totalStringLength / r.dataShards

	shards := make([][]byte, r.dataShards)
	for i := 0; i < totalStringLength; i += splitStringLength {
		end := i + splitStringLength
		shards[i/splitStringLength] = []byte(dataString[i:end])
	}
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
