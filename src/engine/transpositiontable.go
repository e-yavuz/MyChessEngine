package chessengine

const (
	hashfEXACT byte = iota
	hashfALPHA
	hashfBETA
)

type tagHASHE struct {
	key   uint64 // 8 bytes (zobrist key)
	value int16  // 2 bytes (score)
	best  Move   // 2 bytes (best move)
	depth byte   // 1 byte (depth searched when recorded)
	flags byte   // 1 byte (hashfEXACT | hashfALPHA | hashfBETA)
}

const sizeTagHASHE = 12                                 // Size of tagHASHE struct in bytes
const TableCapacity = (1024 * 1024 / sizeTagHASHE) * 16 // 16 MB table with 12 byte size entries
var DebugTableSize = 0
var DebugCollisions = 0
var DebugNewEntries = 0

var hash_table [TableCapacity]tagHASHE

func probeHash(depth byte, alpha, beta int16, key uint64) int16 {
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

func recordHash(depth byte, val int16, hashf byte, bestMove Move, key uint64) {
	phashe := &hash_table[key%TableCapacity]
	if phashe.key == 0 {
		DebugTableSize++
	} else if phashe.key != key {
		DebugCollisions++
	}
	DebugNewEntries++

	phashe.key = key
	phashe.best = bestMove
	phashe.value = val
	phashe.flags = hashf
	phashe.depth = depth
}

func GetEntry(key uint64) tagHASHE {
	return hash_table[key%TableCapacity]
}
