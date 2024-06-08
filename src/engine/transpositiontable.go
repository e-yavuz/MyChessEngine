package chessengine

const (
	hashfEXACT byte = iota
	hashfALPHA
	hashfBETA
)

type tagHASHE struct { // Size: 16 bytes
	key   uint64 // 8 bytes
	depth byte   // 1 byte
	flags byte   // 1 byte
	value int    // 4 bytes
	best  Move   // 2 bytes
}

const TableCapacity = 65536 * 64 // 64 MB table with 16 byte size entries
var TableSize = 0

var hash_table [TableCapacity]tagHASHE

func ProbeHash(depth byte, alpha, beta int, key uint64) int {
	phashe := &hash_table[key%TableCapacity]

	if phashe.key == key {
		if phashe.depth >= depth {
			if phashe.flags == hashfEXACT {
				return phashe.value
			}
			if phashe.flags == hashfALPHA && phashe.value <= alpha {
				return alpha
			}
			if phashe.flags == hashfBETA && phashe.value >= beta {
				return beta
			}
		}
	}
	return MIN_VALUE
}

func RecordHash(depth byte, val int, hashf byte, bestMove Move, key uint64) {
	phashe := &hash_table[key%TableCapacity]
	if phashe.key == 0 {
		TableSize++
	}

	phashe.key = key
	phashe.best = bestMove
	phashe.value = val
	phashe.flags = hashf
	phashe.depth = depth
}

func GetEntry(key uint64) tagHASHE {
	return hash_table[key%TableCapacity]
}
