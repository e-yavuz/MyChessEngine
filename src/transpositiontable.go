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

const TableSize = 65536 // 1 MB table with 16 byte size entries

var hash_table [TableSize]tagHASHE


func ProbeHash(depth byte, alpha, beta int, key uint64) (Move, int) {
	phashe := &hash_table[key%TableSize]

	if phashe.key == key {
		if phashe.depth >= depth {
			if phashe.flags == hashfEXACT {
				return phashe.best, phashe.value
			}
			if phashe.flags == hashfALPHA && phashe.value <= alpha {
				return phashe.best, alpha
			}
			if phashe.flags == hashfBETA && phashe.value >= beta {
				return phashe.best, beta
			}
		}
	}
	return NULL_MOVE, MIN_VALUE
}

func RecordHash(depth byte, val int, hashf byte, bestMove Move, key uint64) {
	phashe := &hash_table[key%TableSize]

	phashe.key = key
	phashe.best = bestMove
	phashe.value = val
	phashe.flags = hashf
	phashe.depth = depth
}
