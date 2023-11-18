# raid6_GF

Used library for Arithmetics and Matrix Algebra over an 8-bit Galois Field.
[galois.go](./raid6/galois.go), [matrix.go](./raid6/matrix.go) are referenced from [galois.go](galois.go),  [matrix.go](https://github.com/klauspost/reedsolomon/blob/master/matrix.go)

Vandermond matrix generation is implemented by me, referring to [Technical Report CS-03-504](https://web.eecs.utk.edu/~jplank/plank/papers/CS-96-332.html)


To run this program, use the command below:

```bash
go run main.go
```

## Function Explanation
- Create new RAID-6 system (number of data, number of parity): `r, err := raid6.BuildRaidSystem(5, 5)`
- Split input into equal size across different disks, and add padding if not divisible: `shards, length := r.Split(data_string)`
- Join collects data from multiple disks, concatenates them into a string, and removes any padding: `output := r.Join(r.DiskArray, length)`
- DropShard drops a shard to trigger an erasure: `err = r.DropShard(8)`
- ReconstructDisk recovers from the erasure of one or more disks: `err = r.ReconstructDisk()`
- CreateBitFlip flips a bit to generate corruption: `err = r.CreateBitFlip(6, 1)`
- ReconstructCorruption recovers from corruption caused by a Parity bit flip: `err = r.ReconstructCorruption()`

## Example output
### Erasure Recovery

```log
Saved Output:
Lorem ipsum dolor sit amet, consectetur adipiscing elit.

Clean Disk Array:
Disk 0: 	L  o  r  e  m     i  p  s  u  m
Disk 1: 	d  o  l  o  r     s  i  t     a  m
Disk 2: 	e  t  ,     c  o  n  s  e  c  t  e
Disk 3: 	t  u  r     a  d  i  p  i  s  c  i
Disk 4: 	n  g     e  l  i  t  .  0  0  0  0
Disk 5: 	d0 61 a3 53 3d 53 20 6b d e6 66 e6
Disk 6: 	92 74 44 70 8a 7e 9b fd 7e 34 f1 47
Disk 7: 	15 73 87 4c c6 4f d2 a2 48 a7 bc d0
Disk 8: 	56 33 1b 9a b4 b3 90 b3 5 3b 42 92
Disk 9: 	f8 2f 56 48 fa 77 db af c6 51 57 90

Erasure Disk Array:
Disk 0: 	L  o  r  e  m     i  p  s  u  m
Disk 1: 	d  o  l  o  r     s  i  t     a  m
Disk 2:
Disk 3:
Disk 4:
Disk 5: 	d0 61 a3 53 3d 53 20 6b d e6 66 e6
Disk 6: 	92 74 44 70 8a 7e 9b fd 7e 34 f1 47
Disk 7:
Disk 8:
Disk 9: 	f8 2f 56 48 fa 77 db af c6 51 57 90

Corrupted Output:
Lorem ipsum dolor sit am////////////////////////////////

Reconstructed Disk Array:
Disk 0: 	L  o  r  e  m     i  p  s  u  m
Disk 1: 	d  o  l  o  r     s  i  t     a  m
Disk 2: 	e  t  ,     c  o  n  s  e  c  t  e
Disk 3: 	t  u  r     a  d  i  p  i  s  c  i
Disk 4: 	n  g     e  l  i  t  .  0  0  0  0
Disk 5: 	d0 61 a3 53 3d 53 20 6b d e6 66 e6
Disk 6: 	92 74 44 70 8a 7e 9b fd 7e 34 f1 47
Disk 7: 	15 73 87 4c c6 4f d2 a2 48 a7 bc d0
Disk 8: 	56 33 1b 9a b4 b3 90 b3 5 3b 42 92
Disk 9: 	f8 2f 56 48 fa 77 db af c6 51 57 90

Recovered Output:
Lorem ipsum dolor sit amet, consectetur adipiscing elit.
```

### Bit Corruption Detection
```log
// Bit flip at Disk [6][1], which is parity disk so recoverable
Corrupt Disk Array:
Disk 0: 	L  o  r  e  m     i  p  s  u  m
Disk 1: 	d  o  l  o  r     s  i  t     a  m
Disk 2: 	e  t  ,     c  o  n  s  e  c  t  e
Disk 3: 	t  u  r     a  d  i  p  i  s  c  i
Disk 4: 	n  g     e  l  i  t  .  0  0  0  0
Disk 5: 	d0 61 a3 53 3d 53 20 6b d e6 66 e6
Disk 6: 	92 75 44 70 8a 7e 9b fd 7e 34 f1 47
Disk 7: 	15 73 87 4c c6 4f d2 a2 48 a7 bc d0
Disk 8: 	56 33 1b 9a b4 b3 90 b3 5 3b 42 92
Disk 9: 	f8 2f 56 48 fa 77 db af c6 51 57 90

Reconstructed Disk Array:
Disk 0: 	L  o  r  e  m     i  p  s  u  m
Disk 1: 	d  o  l  o  r     s  i  t     a  m
Disk 2: 	e  t  ,     c  o  n  s  e  c  t  e
Disk 3: 	t  u  r     a  d  i  p  i  s  c  i
Disk 4: 	n  g     e  l  i  t  .  0  0  0  0
Disk 5: 	d0 61 a3 53 3d 53 20 6b d e6 66 e6
Disk 6: 	92 74 44 70 8a 7e 9b fd 7e 34 f1 47
Disk 7: 	15 73 87 4c c6 4f d2 a2 48 a7 bc d0
Disk 8: 	56 33 1b 9a b4 b3 90 b3 5 3b 42 92
Disk 9: 	f8 2f 56 48 fa 77 db af c6 51 57 90

// Bit flip at Disk [2][1], which is data disk so not recoverable
Corrupt Disk Array:
Disk 0: 	L  o  r  e  m     i  p  s  u  m
Disk 1: 	d  o  l  o  r     s  i  t     a  m
Disk 2: 	e  u  ,     c  o  n  s  e  c  t  e
Disk 3: 	t  u  r     a  d  i  p  i  s  c  i
Disk 4: 	n  g     e  l  i  t  .  0  0  0  0
Disk 5: 	d0 61 a3 53 3d 53 20 6b d e6 66 e6
Disk 6: 	92 74 44 70 8a 7e 9b fd 7e 34 f1 47
Disk 7: 	15 73 87 4c c6 4f d2 a2 48 a7 bc d0
Disk 8: 	56 33 1b 9a b4 b3 90 b3 5 3b 42 92
Disk 9: 	f8 2f 56 48 fa 77 db af c6 51 57 90

Error: possible data disk corruption. cannot recover from data corruption
```
