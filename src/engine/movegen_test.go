package chessengine

import (
	"fmt"
	"testing"
	"time"
)

func Test_StartPosition(t *testing.T) {
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	InitZobristTable()
	test := InitStartBoard()
	var perftOut uint64
	var rootNodes map[string]uint64

	perftOut, rootNodes = Perft(test, 1, true)
	if perftOut != 20 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 1)", 20, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 2, true)
	if perftOut != 400 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 2)", 400, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 3, true)
	if perftOut != 8902 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%s\n%v\n", "Perft(test, 3)", 8902, perftOut, test.DisplayBoard(), rootNodes)
	}

	perftOut, rootNodes = Perft(test, 4, true)
	if perftOut != 197281 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 4)", 197281, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 5, true)
	if perftOut != 4865609 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 5)", 4865609, perftOut, rootNodes)
	}

	startTime := time.Now().UnixMilli()
	perftOut, _ = Perft(test, 6, true)
	if perftOut != 119060324 {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 6)", 119060324, perftOut, rootNodes)
	}
	fmt.Printf("Speed: %d Nodes/sec\n", 1000*uint64(float64(perftOut)/float64(time.Now().UnixMilli()-startTime)))
}

func Test_Position5(t *testing.T) {
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	InitZobristTable()
	test := InitFENBoard("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8")
	var perftOut, expected uint64
	var rootNodes map[string]uint64

	perftOut, rootNodes = Perft(test, 1, true)
	expected = 44
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 1)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 2, true)
	expected = 1486
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 2)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 3, true)
	expected = 62379
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 3)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 4, true)
	expected = 2103487
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 4)", expected, perftOut, rootNodes)
	}

	startTime := time.Now().UnixMilli()
	perftOut, rootNodes = Perft(test, 5, true)
	expected = 89941194
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 5)", expected, perftOut, rootNodes)
	}
	fmt.Printf("Speed: %d Nodes/sec", 1000*uint64(float64(perftOut)/float64(time.Now().UnixMilli()-startTime)))
}

func Test_Position_bk02(t *testing.T) {
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	InitZobristTable()
	test := InitFENBoard("3r1k2/4npp1/1ppr3p/p6P/P2PPPP1/1NR5/5K2/2R5 w - - 0 1")
	var perftOut, expected uint64
	var rootNodes map[string]uint64

	perftOut, rootNodes = Perft(test, 1, true)
	expected = 33
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 1)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 2, true)
	expected = 793
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 2)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 3, true)
	expected = 26013
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 3)", expected, perftOut, rootNodes)
	}

	perftOut, rootNodes = Perft(test, 4, true)
	expected = 622922
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 4)", expected, perftOut, rootNodes)
	}

	startTime := time.Now().UnixMilli()
	perftOut, rootNodes = Perft(test, 5, true)
	expected = 20077998
	if perftOut != expected {
		t.Fatalf("%s failed\n\texpected: %d\n\tgot: %d\n%v\n", "Perft(test, 5)", expected, perftOut, rootNodes)
	}
	fmt.Printf("Speed: %d Nodes/sec", 1000*uint64(float64(perftOut)/float64(time.Now().UnixMilli()-startTime)))
}
