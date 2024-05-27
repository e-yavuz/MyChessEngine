package chessengine

import (
	"fmt"
	"testing"
)

func Test_Basic(t *testing.T) {
	var truth *Board
	var test *Board
	var from, to, flag uint16

	test = InitStartBoard()

	from = E2
	to = E3
	flag = quietFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("rnbqkbnr/pppppppp/8/8/8/4P3/PPPP1PPP/RNBQKBNR b KQkq - 0 1")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	test.UnMakeMove()
	truth = InitStartBoard()

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	fmt.Println("Test_Basic Passed")
}

func Test_3Moves(t *testing.T) {
	var truth *Board
	var test *Board
	var from, to, flag uint16

	// Starting state, Test case Derived from https:// www.chessprogramming.org/Forsyth-Edwards_Notation#Samples
	test = InitStartBoard()

	from = E2
	to = E4
	flag = doublePawnPushFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	from = C7
	to = C5
	flag = doublePawnPushFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	from = G1
	to = F3
	flag = quietFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	fmt.Print("Test Case 1 Passed")

	test.UnMakeMove()
	truth = InitFENBoard("rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	test.UnMakeMove()
	truth = InitFENBoard("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	test.UnMakeMove()
	truth = InitStartBoard()

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	fmt.Println("Test_3Moves Passed")
}
