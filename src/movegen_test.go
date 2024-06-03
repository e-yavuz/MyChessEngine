package chessengine

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

var testBoard Board

func allMoves(board *Board) []Move {
	return append(*board.generateMoves(CAPTURE), *board.generateMoves(QUIET)...)
}

func convertStringToMap(input string) map[string]uint64 {
	result := make(map[string]uint64)
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			continue
		}
		result[key] = value
	}
	return result
}

func compareMoveLists(map1, map2 map[string]uint64) (missingInMap1, missingInMap2 map[string]uint64, mismatched map[string][2]uint64) {
	missingInMap1 = make(map[string]uint64)
	missingInMap2 = make(map[string]uint64)
	mismatched = make(map[string][2]uint64)

	for key, val1 := range map1 {
		if val2, ok := map2[key]; !ok {
			missingInMap2[key] = val1
		} else if val1 != val2 {
			mismatched[key] = [2]uint64{val1, val2}
		}
	}

	for key, value := range map2 {
		if _, ok := map1[key]; !ok {
			missingInMap1[key] = value
		}
	}

	return
}

func perft(board *Board, ply int, rootLevel bool) (retval uint64, rootNodes map[string]uint64) {
	if ply == 0 {
		return 1, nil
	}
	if rootLevel {
		rootNodes = make(map[string]uint64)
	}
	moveList := allMoves(board)
	if ply == 1 && !rootLevel {
		return uint64(len(moveList)), nil
	}

	for _, move := range moveList {
		board.MakeMove(move)
		leafCount, _ := perft(board, ply-1, false)
		retval += leafCount
		if rootLevel {
			rootNodes[MoveToString(move)] += leafCount
		}
		board.UnMakeMove()
	}
	return retval, rootNodes
}

func Test_StartPosition(t *testing.T) {
	InitMagicBitBoardTable("../magic_rook", "../magic_bishop")
	InitZobristTable()
	test := InitStartBoard()
	var perftOut uint64
	var rootNodes map[string]uint64

	perftOut, rootNodes = perft(test, 1, true)
	if perftOut != 20 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 1)", 20, perftOut, rootNodes)
	}

	perftOut, rootNodes = perft(test, 2, true)
	if perftOut != 400 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 2)", 400, perftOut, rootNodes)
	}

	perftOut, rootNodes = perft(test, 3, true)
	if perftOut != 8902 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%s\n%v\n", "perft(test, 3)", 8902, perftOut, test.DisplayBoard(), rootNodes)
	}

	perftOut, rootNodes = perft(test, 4, true)
	if perftOut != 197281 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 4)", 197281, perftOut, rootNodes)
	}

	perftOut, rootNodes = perft(test, 5, true)
	if perftOut != 4865609 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 5)", 4865609, perftOut, rootNodes)
	}

	startTime := time.Now().UnixMilli()
	perftOut, _ = perft(test, 6, true)
	if perftOut != 119060324 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 6)", 119060324, perftOut, rootNodes)
	}
	fmt.Printf("Speed: %d Nodes/sec", 1000*uint64(float64(perftOut)/float64(time.Now().UnixMilli()-startTime)))
}

func Test_Position5(t *testing.T) {
	InitMagicBitBoardTable("../magic_rook", "../magic_bishop")
	InitZobristTable()
	test := InitFENBoard("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8")
	var perftOut, expected uint64
	var rootNodes map[string]uint64

	perftOut, rootNodes = perft(test, 1, true)
	expected = 44
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 1)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = perft(test, 2, true)
	expected = 1486
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 2)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = perft(test, 3, true)
	expected = 62379
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 3)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = perft(test, 4, true)
	expected = 2103487
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 4)", expected, perftOut, rootNodes)
	}

	startTime := time.Now().UnixMilli()
	perftOut, rootNodes = perft(test, 5, true)
	expected = 89941194
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "perft(test, 5)", expected, perftOut, rootNodes)
	}
	fmt.Printf("Speed: %d Nodes/sec", 1000*uint64(float64(perftOut)/float64(time.Now().UnixMilli()-startTime)))
}
