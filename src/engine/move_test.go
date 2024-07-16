package chessengine

import (
	"fmt"
	"testing"
)

func (board *Board) equalNoStateCompare(other *Board) bool {
	bitBoardsCompare := board.B.Pawn == other.B.Pawn &&
		board.B.Knight == other.B.Knight &&
		board.B.Rook == other.B.Rook &&
		board.B.Bishop == other.B.Bishop &&
		board.B.Queen == other.B.Queen &&
		board.B.King == other.B.King &&
		board.W.Pawn == other.W.Pawn &&
		board.W.Knight == other.W.Knight &&
		board.W.Rook == other.W.Rook &&
		board.W.Bishop == other.W.Bishop &&
		board.W.Queen == other.W.Queen &&
		board.W.King == other.W.King

	PieceInfoArrCompare := true

	for i := 0; i < 64; i++ {
		if board.PieceInfoArr[i] != nil && other.PieceInfoArr[i] != nil {
			PieceInfoArrCompare = board.PieceInfoArr[i].Equal(other.PieceInfoArr[i]) &&
				PieceInfoArrCompare
		} else {
			PieceInfoArrCompare = board.PieceInfoArr[i] == nil &&
				other.PieceInfoArr[i] == nil && PieceInfoArrCompare
		}

		if !PieceInfoArrCompare {
			return false
		}

	}

	return bitBoardsCompare
}

func (board *Board) equalNoCapture(other *Board) bool {
	bitBoardsCompare := board.B.Pawn == other.B.Pawn &&
		board.B.Knight == other.B.Knight &&
		board.B.Rook == other.B.Rook &&
		board.B.Bishop == other.B.Bishop &&
		board.B.Queen == other.B.Queen &&
		board.B.King == other.B.King &&
		board.W.Pawn == other.W.Pawn &&
		board.W.Knight == other.W.Knight &&
		board.W.Rook == other.W.Rook &&
		board.W.Bishop == other.W.Bishop &&
		board.W.Queen == other.W.Queen &&
		board.W.King == other.W.King

	stCompare := board.GetTopState().equalNoCapture(other.GetTopState())

	PieceInfoArrCompare := true

	for i := 0; i < 64; i++ {
		if board.PieceInfoArr[i] != nil && other.PieceInfoArr[i] != nil {
			PieceInfoArrCompare = board.PieceInfoArr[i].Equal(other.PieceInfoArr[i]) &&
				PieceInfoArrCompare
		} else {
			PieceInfoArrCompare = board.PieceInfoArr[i] == nil &&
				other.PieceInfoArr[i] == nil && PieceInfoArrCompare
		}

		if !PieceInfoArrCompare {
			return false
		}

	}

	return bitBoardsCompare && stCompare
}

func (si *StateInfo) equalNoCapture(other *StateInfo) bool {
	var promotionCompare, valueCompare bool

	if si.PrePromotionBitBoard != nil && other.PrePromotionBitBoard != nil {
		promotionCompare = *si.PrePromotionBitBoard == *other.PrePromotionBitBoard
	} else {
		promotionCompare = si.PrePromotionBitBoard == other.PrePromotionBitBoard
	}

	valueCompare = si.EnPassantPosition == other.EnPassantPosition &&
		si.IsWhiteTurn == other.IsWhiteTurn &&
		si.getCastleWKing() == other.getCastleWKing() &&
		si.getCastleBKing() == other.getCastleBKing() &&
		si.getCastleWQueen() == other.getCastleWQueen() &&
		si.getCastleBQueen() == other.getCastleBQueen() &&
		si.HalfMoveClock == other.HalfMoveClock &&
		si.TurnCounter == other.TurnCounter

	return promotionCompare && valueCompare
}

func Test_Basic(t *testing.T) {
	InitZobristTable()
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	var truth *Board
	var test *Board
	var from, to Position
	var flag uint16

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
}

func Test_3Moves(t *testing.T) {
	InitZobristTable()
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	var truth *Board
	var test *Board
	var from, to Position
	var flag uint16

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

	fmt.Println("Test Case MakeMove Passed")

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

	fmt.Println("Test Case UnMakeMove Passed")
}

func Test_EnPassant(t *testing.T) {
	InitZobristTable()
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	var truth *Board
	var test *Board
	var from, to Position
	var flag uint16

	// Test for en-passant capture

	test = InitFENBoard("rnbqkbnr/1pp1pppp/p7/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3")

	from = E5
	to = D6
	flag = epCaptureFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("rnbqkbnr/1pp1pppp/p2P4/8/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 3")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}
}
func Test_GetIntermediaryRay(t *testing.T) {
	InitZobristTable()
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	from := E4
	to := H7
	expected := BitBoard(70506183131136)
	result := getIntermediaryRay(from, to)
	if result != expected {
		t.Fatalf("\tFrom:\n%s\n\tTo:\n%s\n\tExpected:\n%s\n\tGot:\n%s", BitBoardToString(1<<from), BitBoardToString(1<<to), BitBoardToString(expected), BitBoardToString(result))
	}

	from = A1
	to = H8
	expected = BitBoard(18049651735527936)
	result = getIntermediaryRay(from, to)
	if result != expected {
		t.Fatalf("\tFrom:\n%s\n\tTo:\n%s\n\tExpected:\n%s\n\tGot:\n%s", BitBoardToString(1<<from), BitBoardToString(1<<to), BitBoardToString(expected), BitBoardToString(result))
	}

	from = D4
	to = H4
	expected = BitBoard(1879048192)
	result = getIntermediaryRay(from, to)
	if result != expected {
		t.Fatalf("\tFrom:\n%s\n\tTo:\n%s\n\tExpected:\n%s\n\tGot:\n%s", BitBoardToString(1<<from), BitBoardToString(1<<to), BitBoardToString(expected), BitBoardToString(result))
	}

	from = E5
	to = E1
	expected = BitBoard(269488128)
	result = getIntermediaryRay(from, to)
	if result != expected {
		t.Fatalf("\tFrom:\n%s\n\tTo:\n%s\n\tExpected:\n%s\n\tGot:\n%s", BitBoardToString(1<<from), BitBoardToString(1<<to), BitBoardToString(269488128), BitBoardToString(result))
	}

	from = B6
	to = G1
	expected = BitBoard(17315143680)
	result = getIntermediaryRay(from, to)
	if result != expected {
		t.Fatalf("\tFrom:\n%s\n\tTo:\n%s\n\tExpected:\n%s\n\tGot:\n%s", BitBoardToString(1<<from), BitBoardToString(1<<to), BitBoardToString(expected), BitBoardToString(result))
	}

	from = B5
	to = G1
	expected = BitBoard(0)
	result = getIntermediaryRay(from, to)
	if result != expected {
		t.Fatalf("\tFrom:\n%s\n\tTo:\n%s\n\tExpected:\n%s\n\tGot:\n%s", BitBoardToString(1<<from), BitBoardToString(1<<to), BitBoardToString(expected), BitBoardToString(result))
	}
}

func Test_Castling(t *testing.T) {
	InitZobristTable()
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	var truth *Board
	var test *Board
	var from, to Position
	var flag uint16

	// Test for king-side castling

	test = InitFENBoard("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQK2R w KQkq - 0 1")
	from = E1
	to = G1
	flag = kingCastleFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQ1RK1 b kq - 1 1")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for queen-side castling
	test = InitFENBoard("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1")
	from = E1
	to = C1
	flag = queenCastleFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/2KR3R b kq - 1 1")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for king-side castling negated after rook move

	test = InitFENBoard("8/8/8/8/8/8/8/RNBQK2R w KQ - 0 1")
	from = H1
	to = G1
	flag = quietFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("8/8/8/8/8/8/8/RNBQK1R1 b Q - 1 1")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for queen-side castling negated after rook move
	test = InitFENBoard("8/8/8/8/8/8/8/R3K2R w KQ - 0 1")
	from = A1
	to = B1
	flag = quietFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("8/8/8/8/8/8/8/1R2K2R b K - 1 1")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for both sides castling negated after white king move
	test = InitFENBoard("8/8/8/8/8/8/8/R3K2R w KQ - 0 1")
	from = E1
	to = E2
	flag = quietFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("8/8/8/8/8/8/4K3/R6R b - - 1 1")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for both sides castling negated after black king move
	test = InitFENBoard("r3k2r/8/8/8/8/8/8/8 b kq - 0 1")
	from = E8
	to = E7
	flag = quietFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("r6r/4k3/8/8/8/8/8/8 w - - 1 2")

	if !test.Equal(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for king castling negated after king rook captured
	test = InitFENBoard("8/8/8/8/8/8/7q/R3K2R b KQ - 0 1")
	from = H2
	to = H1
	flag = captureFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("8/8/8/8/8/8/8/R3K2q w Q - 0 2")

	if !test.equalNoCapture(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for queen castling negated after queen rook captured
	test = InitFENBoard("8/8/8/8/8/8/q7/R3K2R b KQ - 0 1")
	from = A2
	to = A1
	flag = captureFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("8/8/8/8/8/8/8/q3K2R w K - 0 2")

	if !test.equalNoCapture(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}
}

func Test_Promotion(t *testing.T) {
	InitZobristTable()
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	var truth *Board
	var test *Board
	var from, to Position
	var flag uint16

	// Test for queen promotion

	test = InitFENBoard("8/P7/8/8/8/8/8/8 w - - 0 1")
	from = A7
	to = A8
	flag = queenPromotionFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("Q7/8/8/8/8/8/8/8 b - - 1 2")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for knight promotion

	test = InitFENBoard("8/P7/8/8/8/8/8/8 w - - 0 1")
	from = A7
	to = A8
	flag = knightPromotionFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("N7/8/8/8/8/8/8/8 b - - 1 2")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for bishop promotion

	test = InitFENBoard("8/P7/8/8/8/8/8/8 w - - 0 1")
	from = A7
	to = A8
	flag = bishopPromotionFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("B7/8/8/8/8/8/8/8 b - - 1 2")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for rook promotion

	test = InitFENBoard("8/P7/8/8/8/8/8/8 w - - 0 1")
	from = A7
	to = A8
	flag = rookPromotionFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("R7/8/8/8/8/8/8/8 b - - 1 2")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for queen promotion capture

	test = InitFENBoard("1p6/P7/8/8/8/8/8/8 w - - 0 1")
	from = A7
	to = B8
	flag = queenPromoCaptureFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("1Q6/8/8/8/8/8/8/8 b - - 1 2")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for knight promotion capture

	test = InitFENBoard("1p6/P7/8/8/8/8/8/8 w - - 0 1")
	from = A7
	to = B8
	flag = knightPromoCaptureFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("1N6/8/8/8/8/8/8/8 b - - 1 2")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for bishop promotion capture

	test = InitFENBoard("1p6/P7/8/8/8/8/8/8 w - - 0 1")
	from = A7
	to = B8
	flag = bishopPromoCaptureFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("1B6/8/8/8/8/8/8/8 b - - 1 2")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}

	// Test for rook promotion capture

	test = InitFENBoard("1p6/P7/8/8/8/8/8/8 w - - 0 1")
	from = A7
	to = B8
	flag = rookPromoCaptureFlag
	test.MakeMove(NewMove(from, to, flag))

	truth = InitFENBoard("1R6/8/8/8/8/8/8/8 b - - 1 2")

	if !test.equalNoStateCompare(truth) {
		t.Fatalf("Move %d->%d with flag: %d\nWanted:\n%s\nGot:\n%s", from, to, flag, truth.DisplayBoard(), test.DisplayBoard())
	}
}
