package chessengine

import "fmt"

const (
	PVnode byte = iota
	ALLnode
	CUTnode
	NULLnode
)

// TODO work on TT optimization, seems to function identical to single entry max-depth TT

type ttSubEntry struct { // 15 bytes
	zobristKey uint64 // 8 bytes
	score      int    // 4 bytes
	move       Move   // 2 bytes
	ttInfo     byte   // 1 bytes
	turn       byte   // 1 bytes
}

var NULLttSubEntry = ttSubEntry{}

func getDepth(ttInfo byte) byte {
	return ttInfo >> 2
}

// func getCounter(ttInfo uint16) byte {
// 	return byte(ttInfo & 0xFF00)
// }

func getNodeType(ttInfo byte) byte {
	return ttInfo & 0x3
}

// func makeTTInfo(flag, depth, counter byte) uint16 {
// 	return uint16(flag) + (uint16(depth) << 2) + (uint16(counter) << 8)
// }

func makeTTInfo(flag, depth byte) byte {
	return flag + (depth << 2)
}

type ttEntry struct { // Size: 48 bytes
	subEntries [ttEntry_ARcount]ttSubEntry // 16x(ttEntry_ARcount + 1 maxDepthNode) bytes
}

func getBestSubEntry(depth, turn byte, zobristKey uint64) (ttSubEntry, byte, Move) {
	entry := &hash_table[zobristKey%TableCapacity]
	retvalSubEntry := NULLttSubEntry
	retvalNodeType := NULLnode
	retvalMove := NULL_MOVE

	for i := 0; i < ttEntry_ARcount; i++ {
		if entry.subEntries[i].zobristKey == zobristKey &&
			getDepth(entry.subEntries[i].ttInfo) >= depth &&
			entry.subEntries[i].turn >= turn {

			retvalSubEntry = entry.subEntries[i] // found a matching zobristKey with higher depth
			break

		}
	}

	if entry.subEntries[0] != NULLttSubEntry {
		retvalNodeType = getNodeType(entry.subEntries[0].ttInfo)
		retvalMove = entry.subEntries[0].move
	}

	return retvalSubEntry, retvalNodeType, retvalMove // nothing found that matches this zobristKey
}

func getReplaceEntry(depth byte, turn byte, zobristKey uint64) *ttSubEntry {
	entry := &hash_table[zobristKey%TableCapacity]
	var retval *ttSubEntry

	if turn > entry.subEntries[0].turn { // Current entry is stale/empty, no need to shift
		retval = &entry.subEntries[0]
	} else if depth >= getDepth(entry.subEntries[0].ttInfo) { // Could replace max depth entry
		for i := 1; i < ttEntry_ARcount; i++ {
			entry.subEntries[i] = entry.subEntries[i-1]
		}
		retval = &entry.subEntries[0]
	} else { // If cannot replace max depth entry, shift all always-replaces entries right one, and push in new entry
		for i := 2; i < ttEntry_ARcount; i++ {
			entry.subEntries[i] = entry.subEntries[i-1]
		}
		retval = &entry.subEntries[1]
	}

	if turn > entry.subEntries[0].turn { // Updated a stale/empty entry, so TT now has a new valid entry
		DebugTableSize++
	} else if retval.zobristKey != zobristKey {
		DebugKeyCollisions++
	} else if retval.zobristKey == zobristKey {
		DebugIndexCollisions++
	}
	DebugNewEntries++

	return retval
}

const ttEntry_ARcount = 3 // 2 always replace entries and one max depth
const sizeTagHASHE = ttEntry_ARcount * 16
const TableCapacity = (1024 * 1024 / sizeTagHASHE) * 64 // 64 MB table with 45 byte entries
var DebugTableSize = 0
var DebugKeyCollisions = 0
var DebugIndexCollisions = 0
var DebugNewEntries = 0
var DebugDroppedEntries = 0

var hash_table [TableCapacity]ttEntry

func probeHash(depth, turn byte, alpha, beta int, zobristKey uint64) (int, byte, Move) {
	subEntry, entryNodeType, entryMove := getBestSubEntry(depth, turn, zobristKey)

	if subEntry != NULLttSubEntry {
		nodeType := getNodeType(subEntry.ttInfo)
		if nodeType == PVnode {
			return subEntry.score, PVnode, subEntry.move
		}
		if nodeType == ALLnode && subEntry.score <= alpha {
			return alpha, ALLnode, NULL_MOVE
		}
		if nodeType == CUTnode && subEntry.score >= beta {
			return beta, CUTnode, NULL_MOVE
		}
	}
	return MIN_VALUE, entryNodeType, entryMove
}

func recordHash(depth, nodeType, turn byte, score int, bestMove Move, zobristKey uint64) {
	replacedSubEntry := getReplaceEntry(depth, turn, zobristKey)

	replacedSubEntry.zobristKey = zobristKey
	replacedSubEntry.move = bestMove
	replacedSubEntry.score = score
	replacedSubEntry.ttInfo = makeTTInfo(nodeType, depth)
	replacedSubEntry.turn = turn
}

func GetEntry(zobristKey uint64) ttEntry {
	return hash_table[zobristKey%TableCapacity]
}

func TTClear() {
	hash_table = [TableCapacity]ttEntry{}
	DebugTableSize = 0
	DebugKeyCollisions = 0
	DebugIndexCollisions = 0
	DebugDroppedEntries = 0
	DebugNewEntries = 0
}

func TTDebugInfo() string {
	return fmt.Sprintf("TT occupancy: %0.2f%%\n\tNewEntries: %d(%f%%)\n\tKey Collisions: %d(%f%%)\n\tIndex Collisions: %d(%f%%)\n", 100*float32(DebugTableSize)/float32(TableCapacity), DebugNewEntries, float32(DebugNewEntries)/float32(DebugTableSize), DebugKeyCollisions, float32(DebugKeyCollisions)/float32(DebugNewEntries), DebugIndexCollisions, float32(DebugIndexCollisions)/float32(DebugNewEntries))
}
