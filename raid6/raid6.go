package raid6

import "fmt"

const gfMaxSize = 256
const generator = 0x2

// https://github.com/klauspost/reedsolomon

type raid6 struct {
	dataShards      int
	parityShards    int
	totalShards     int
	encoding_matrix matrix
	disk_matrix     matrix
	disk_parity     matrix
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
	fmt.Println("original matrix: ", result)

	// Perform Gausian elimination to transform data part into an identity matrix
	// Skip r == 0 because that is already identity row.
	for r := 1; r < cols; r++ {
		factor := result[r][r]
		fmt.Println("divisor: ", factor)

		// first make result[r][r] == 1 by dividing the column by result[r][r]
		if factor != 1 {
			for r1 := r; r1 < rows; r1++ {
				result[r1][r] = galDivide(result[r1][r], factor)
			}
		}
		fmt.Println("divided with factor: ", result)

		// then subtract this row from all other rows to make all other entries in this column 0
		for c1 := 0; c1 < cols; c1++ {
			if c1 == r {
				continue
			} else {
				multiplier := result[r][c1]
				fmt.Println("multiplier: ", multiplier)
				for r1 := r; r1 < rows; r1++ {
					// Subtract rows by applying bitwise XOR
					result[r1][c1] ^= galMultiply(multiplier, result[r1][r])
				}
			}
		}

		fmt.Printf("subtracted matrix: %d \n", result)
	}

	return result
}

func NewRaidSystem(dataShards, parityShards int) *raid6 {
	// We perform encoding when save data to disk
	// encoding_matrix(n+m, n) * data_matrix(n,n) = data_shard(n+m,n)

	// [   identity matrix  ]   [      ]   [  data  ]
	// [--------------------] * [ data ] = [--------]
	// [ vandermonde matrix ]   [      ]   [ parity ]

	r := raid6{
		dataShards:   dataShards,
		parityShards: parityShards,
		totalShards:  dataShards + parityShards,
	}

	r.encoding_matrix = fixedVandermond(r.totalShards, r.dataShards)

	return &r
}
