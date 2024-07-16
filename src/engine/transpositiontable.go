package chessengine

import (
	"fmt"
)

const (
	PVnode byte = iota
	ALLnode
	CUTnode
	NULLnode
)

// TODO work on TT optimization, seems to function identical to single entry max-depth TT
// july 2: switch to new version on v7f seems to break it???

type ttSubEntry struct { // 16 bytes
	zobristKey uint64 // 8 bytes
	score      int    // 4 bytes
	move       Move   // 2 bytes
	turn       byte   // 1 byte
	ttInfo     byte   // 1 byte
}

var NULLttSubEntry = ttSubEntry{}

func getDepth(ttInfo byte) int8 {
	return int8(ttInfo >> 2)
}

func getNodeType(ttInfo byte) byte {
	return ttInfo & 0x3
}

func makeTTInfo(flag byte, depth int8) byte {
	return flag + byte(depth<<2)
}

type ttEntry struct { // Size: 48 bytes
	subEntries [ttEntry_ARcount]ttSubEntry // 16x(ttEntry_ARcount + 1 maxDepthNode) bytes
}

const ttEntry_ARcount = 3 // 2 always replace entries and one max depth
const sizeTagHASHE = ttEntry_ARcount * 16
const DefaultTTMBSize = 16

var TableCapacity uint64 = (1024 * 1024 / sizeTagHASHE) * DefaultTTMBSize // 16 MB table with 72 byte entries
var DebugTableSize = 0
var DebugKeyCollisions = 0
var DebugIndexCollisions = 0
var DebugNewEntries = 0
var DebugTableMinimumMoveCount uint16 = 0 // Marks the point at which an entry will never be usable, thus updating it increases Table size!
var DebugTableHits = 0
var DebugTableProbes = 0

var hash_table []ttEntry

func init() {
	TTReset(nil, DefaultTTMBSize)
}

func getBestSubEntry(depth int8, turn byte, zobristKey uint64) (ttSubEntry, byte, Move) {
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

func getReplaceEntry(depth int8, turn byte, zobristKey uint64) *ttSubEntry {
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

func probeHash(depth int8, turn byte, alpha, beta int, zobristKey uint64) (int, byte, Move) {
	subEntry, entryNodeType, entryMove := getBestSubEntry(depth, turn, zobristKey)
	DebugTableProbes++

	if subEntry != NULLttSubEntry {
		DebugTableHits++
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

func recordHash(depth int8, nodeType, turn byte, score int, bestMove Move, zobristKey uint64) {
	replacedSubEntry := getReplaceEntry(depth, turn, zobristKey)

	replacedSubEntry.zobristKey = zobristKey
	replacedSubEntry.move = bestMove
	replacedSubEntry.score = score
	replacedSubEntry.ttInfo = makeTTInfo(nodeType, depth)
	replacedSubEntry.turn = turn
}

func TTReset(board *Board, sizeMB uint64) {
	TableCapacity = (1024 * 1024 / sizeTagHASHE) * sizeMB
	hash_table = make([]ttEntry, TableCapacity)
	DebugTableSize = 0
	TTDebugReset(board)
}

func TTDebugReset(board *Board) {
	DebugKeyCollisions = 0
	DebugIndexCollisions = 0
	DebugNewEntries = 0
	DebugTableHits = 0
	DebugTableProbes = 0
	if board == nil {
		DebugTableMinimumMoveCount = 0
	} else {
		DebugTableMinimumMoveCount = board.moveCount() - 1
	}
}

func TTDebugInfo() string {
	return fmt.Sprintf("TT occupancy: %0.2f%%\n\tNew Entries: %d(%0.2f%%)\n\tKey Collisions: %d(%0.2f%%)\n\tIndex Collisions: %d(%0.2f%%)\n\tHit Rate: %0.2f%%\n", 100*float32(DebugTableSize)/float32(TableCapacity), DebugNewEntries, 100*float32(DebugNewEntries)/float32(ttEntry_ARcount*DebugTableSize), DebugKeyCollisions, 100*float32(DebugKeyCollisions)/float32(DebugNewEntries), DebugIndexCollisions, 100*float32(DebugIndexCollisions)/float32(DebugNewEntries), 100*float32(DebugTableHits)/float32(DebugTableProbes))
}
