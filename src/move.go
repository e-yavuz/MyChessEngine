package chessengine

import (
	"fmt"
	"log"
)

/*
	Contains
		info for move + all possible flags
		Constant indicies for each possible square
*/

const (
	A1 Position = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
)

const NULL_MOVE Move = 0

type Position = byte
type Move = uint16
type Flag = uint16

func PositionToSquare(position Position) string {
	col := position & 0b111
	row := (position >> 3) & 0b111

	return string(rune('a'+col)) + string(rune('1'+row))
}

// Returns a bitboard of 1's between the given two positions, (excluding the two positions) if they are along a diagonal, row, or column
func GetIntermediaryRay(from, to Position) BitBoard {
	if from == to {
		return 0
	}

	fromCol := from & 0b111
	toCol := to & 0b111
	fromRow := (from >> 3) & 0b111
	toRow := (to >> 3) & 0b111
	// If on same row
	if fromRow == toRow {
		correctRow := Shift(Row1Full, N*int8(fromRow))
		return Shift(correctRow, E*int8(min(fromCol, toCol)+1)) & Shift(correctRow, W*int8(1+7-max(fromCol, toCol)))
	}
	// If on same column
	if fromCol == toCol {
		correctCol := Shift(Col1Full, E*int8(fromCol))
		return Shift(correctCol, N*int8(min(fromRow, toRow)+1)) & Shift(correctCol, S*int8(1+7-max(fromRow, toRow)))
	}
	// If on same diagonal
	if abs(int(fromCol)-int(toCol)) == abs(int(fromRow)-int(toRow)) {
		excessDiagonal := bishopCalculatedMoves[from] & bishopCalculatedMoves[to]
		// Idea is to move North and South to strip the outer bits
		northRemoved := Shift(Shift(excessDiagonal, N*int8(1+7-max(fromRow, toRow))), -N*int8(1+7-max(fromRow, toRow)))
		southRemoved := Shift(Shift(excessDiagonal, S*int8(1+min(fromRow, toRow))), -S*int8(1+min(fromRow, toRow)))
		return northRemoved & southRemoved
	}

	return 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func NewMove(from, to Position, flag uint16) Move {
	return Move(from) + (Move(to) << 6) + (flag << 12)
}

func GetStartingPosition(move Move) Position {
	return Position(move & 0x3F)
}

func GetTargetPosition(move Move) Position {
	return Position((move & 0xFC0) >> 6)
}

/*
FLAGS
code	promotion	capture	special 1	special 0	kind of move
0	0	0	0	0	quiet moves
1	0	0	0	1	double pawn push
2	0	0	1	0	king castle
3	0	0	1	1	queen castle
4	0	1	0	0	captures
5	0	1	0	1	ep-capture
8	1	0	0	0	knight-promotion
9	1	0	0	1	queen-promotion
11	1	1	0	0	knight-promo capture
13	1	1	0	1	queen-promo capture
*/

const (
	quietFlag              Flag = 0b0000
	doublePawnPushFlag     Flag = 0b0001
	kingCastleFlag         Flag = 0b0010
	queenCastleFlag        Flag = 0b0011
	captureFlag            Flag = 0b0100
	epCaptureFlag          Flag = 0b0101
	knightPromotionFlag    Flag = 0b1000 // 0
	rookPromotionFlag      Flag = 0b1001 // 1
	bishopPromotionFlag    Flag = 0b1010 // 2
	queenPromotionFlag     Flag = 0b1011 // 3
	knightPromoCaptureFlag Flag = 0b1100
	rookPromoCaptureFlag   Flag = 0b1101
	bishopPromoCaptureFlag Flag = 0b1110
	queenPromoCaptureFlag  Flag = 0b1111
)

func FlagToString(flag Flag) string {
	switch flag {
	case quietFlag:
		return "quietFlag"
	case doublePawnPushFlag:
		return "doublePawnPushFlag"
	case kingCastleFlag:
		return "kingCastleFlag"
	case queenCastleFlag:
		return "queenCastleFlag"
	case captureFlag:
		return "captureFlag"
	case epCaptureFlag:
		return "epCaptureFlag"
	case knightPromotionFlag:
		return "knightPromotionFlag"
	case queenPromotionFlag:
		return "queenPromotionFlag"
	case knightPromoCaptureFlag:
		return "knightPromoCaptureFlag"
	case queenPromoCaptureFlag:
		return "queenPromoCaptureFlag"
	default:
		return "InvalidFlag"
	}
}

func GetFlag(move Move) Flag {
	return move >> 12
}

// Invariant: Assumes move is legal
func (board *Board) MakeMove(move Move) {
	if move == NULL_MOVE {
		return
	}

	from := GetStartingPosition(move)
	to := GetTargetPosition(move)

	piece := board.PieceInfoArr[from]
	currentState := board.GetTopState()

	// Copy over from currentState to a new state
	st := &StateInfo{
		IsWhiteTurn:       !currentState.IsWhiteTurn,
		HalfMoveClock:     currentState.HalfMoveClock + 1,
		TurnCounter:       currentState.TurnCounter,
		CastleWKing:       currentState.CastleWKing,
		CastleBKing:       currentState.CastleBKing,
		CastleWQueen:      currentState.CastleWQueen,
		CastleBQueen:      currentState.CastleBQueen,
		EnPassantPosition: INVALID_POSITION,
		PrecedentMove:     move,
	}

	if !currentState.IsWhiteTurn {
		st.TurnCounter += 1
		if piece.thisBitBoard == &board.B.Pawn {
			st.HalfMoveClock = 0
		}
	} else if piece.thisBitBoard == &board.W.Pawn {
		st.HalfMoveClock = 0
	}

	// Start by removing this piece from the bitBoard it was originally, it is essentially in limbo now
	RemoveFromBitBoard(piece.thisBitBoard, GetStartingPosition(move))

	// Now need to check flags to see what is happening!
	switch GetFlag(move) {
	case doublePawnPushFlag:
		st.EnPassantPosition = (from + to) / 2
	case kingCastleFlag:
		if piece.isWhite {
			st.CastleWKing = false
			st.CastleWQueen = false
			board.swapPiecePositions(&board.W.Rook, to+1, to-1)
		} else {
			st.CastleBKing = false
			st.CastleBQueen = false
			board.swapPiecePositions(&board.B.Rook, to+1, to-1)
		}
	case queenCastleFlag:
		if piece.isWhite {
			st.CastleWKing = false
			st.CastleWQueen = false
			board.swapPiecePositions(&board.W.Rook, to-2, to+1)
		} else {
			st.CastleBKing = false
			st.CastleBQueen = false
			board.swapPiecePositions(&board.B.Rook, to-2, to+1)
		}
	}

	// Capture move
	if captureFlag&GetFlag(move) > 0 {
		if GetFlag(move) == epCaptureFlag {
			if currentState.IsWhiteTurn {
				RemoveFromBitBoard(&board.B.Pawn, to-8)
				st.Capture = board.PieceInfoArr[to-8]
				board.PieceInfoArr[to-8] = nil
			} else {
				RemoveFromBitBoard(&board.W.Pawn, to+8)
				st.Capture = board.PieceInfoArr[to+8]
				board.PieceInfoArr[to+8] = nil
			}
		} else {
			RemoveFromBitBoard(board.PieceInfoArr[to].thisBitBoard, to)
			st.Capture = board.PieceInfoArr[to]
		}
		// Rook Castle Partner captured -> Cannot Castle that side
		if currentState.CastleWKing && to == 7 {
			st.CastleWKing = false
		} else if currentState.CastleWQueen && to == 0 {
			st.CastleWQueen = false
		} else if currentState.CastleBKing && to == 63 {
			st.CastleBKing = false
		} else if currentState.CastleBQueen && to == 56 {
			st.CastleBQueen = false
		}
		// All captures reset half move clock
		st.HalfMoveClock = 0
	}
	// Update PieceInfoArr from -> to
	board.PieceInfoArr[to] = piece
	board.PieceInfoArr[from] = nil
	// Promotion move
	if GetFlag(move)&0b1000 > 0 {
		st.PrePromotionBitBoard = piece.thisBitBoard
		switch GetFlag(move) & 0b11 {
		case 0: // Knight
			if piece.isWhite {
				piece.thisBitBoard = &board.W.Knight
			} else {
				piece.thisBitBoard = &board.B.Knight
			}
		case 1: // Rook
			if piece.isWhite {
				piece.thisBitBoard = &board.W.Rook
			} else {
				piece.thisBitBoard = &board.B.Rook
			}
		case 2: // Bishop
			if piece.isWhite {
				piece.thisBitBoard = &board.W.Bishop
			} else {
				piece.thisBitBoard = &board.B.Bishop
			}
		case 3: // Queen
			if piece.isWhite {
				piece.thisBitBoard = &board.W.Queen
			} else {
				piece.thisBitBoard = &board.B.Queen
			}
		}
	}
	PlaceOnBitBoard(piece.thisBitBoard, to)
	// Rooks moving -> Castle side negated
	// Hard coded because faster than map lookup?
	if st.CastleWKing && piece.pieceTYPE == ROOK && from == 7 {
		st.CastleWKing = false
	} else if st.CastleWQueen && piece.pieceTYPE == ROOK && from == 0 {
		st.CastleWQueen = false
	} else if st.CastleBKing && piece.pieceTYPE == ROOK && from == 63 {
		st.CastleBKing = false
	} else if st.CastleBQueen && piece.pieceTYPE == ROOK && from == 56 {
		st.CastleBQueen = false
	}
	// King moving -> Both castle sides negated
	if (st.CastleWKing || st.CastleWQueen) && piece.thisBitBoard == &board.W.King {
		st.CastleWKing = false
		st.CastleWQueen = false
	}
	if (st.CastleBKing || st.CastleBQueen) && piece.thisBitBoard == &board.B.King {
		st.CastleBKing = false
		st.CastleBQueen = false
	}

	board.PushNewState(st)
}

func (board *Board) UnMakeMove() {
	topState := board.PopTopState()
	move := topState.PrecedentMove
	if move == NULL_MOVE {
		panic(fmt.Sprintf("No precedent move recorded for current state:\n%v", topState))
	}

	from := GetStartingPosition(move)
	to := GetTargetPosition(move)

	piece := board.PieceInfoArr[to]

	if piece == nil {
		log.Printf("From: %d, To: %d", from, to)
	}

	// Start by removing this piece from the bitBoard it was originally, it is essentially in limbo now
	RemoveFromBitBoard(piece.thisBitBoard, to)

	// Reverse engineer castling
	switch GetFlag(move) {
	case kingCastleFlag: // From left of king's "to" position to his right
		if piece.isWhite {
			board.swapPiecePositions(&board.W.Rook, to-1, to+1)
		} else {
			board.swapPiecePositions(&board.B.Rook, to-1, to+1)
		}
	case queenCastleFlag: // From right of king's "to" position to his left
		if piece.isWhite {
			board.swapPiecePositions(&board.W.Rook, to+1, to-2)
		} else {
			board.swapPiecePositions(&board.B.Rook, to+1, to-2)
		}
	}

	// Put piece on from position in InfoArr
	board.PieceInfoArr[from] = piece

	// If move was a capture, replace back captured piece in InfoArr on "to" and move it onto its bitboard,
	// otherwise simply nil the infoarr spot as nothing exists there now
	if captureFlag&GetFlag(move) > 0 {
		if GetFlag(move) == epCaptureFlag {
			if topState.IsWhiteTurn {
				board.PieceInfoArr[to+8] = topState.Capture
				PlaceOnBitBoard(topState.Capture.thisBitBoard, to+8)
			} else {
				board.PieceInfoArr[to-8] = topState.Capture
				PlaceOnBitBoard(topState.Capture.thisBitBoard, to-8)
			}
		} else {
			board.PieceInfoArr[to] = topState.Capture
			PlaceOnBitBoard(topState.Capture.thisBitBoard, to)
		}
	} else {
		board.PieceInfoArr[to] = nil
	}

	// If promoted, get the bitboard the piece belonged to pre-promotion
	if GetFlag(move)&0b1000 > 0 {
		piece.thisBitBoard = topState.PrePromotionBitBoard
	}

	// Place back onto original bitboard position
	PlaceOnBitBoard(piece.thisBitBoard, from)

}

func (board *Board) swapPiecePositions(bitboard *BitBoard, startPosition, targetPosition Position) {
	RemoveFromBitBoard(bitboard, startPosition)
	PlaceOnBitBoard(bitboard, targetPosition)
	board.PieceInfoArr[targetPosition] = board.PieceInfoArr[startPosition]
	board.PieceInfoArr[startPosition] = nil
}

func MoveToStringWithFlag(move Move) string {
	return fmt.Sprintf("%s%s, flag: %s", PositionToSquare(GetStartingPosition(move)), PositionToSquare(GetTargetPosition(move)), FlagToString(GetFlag(move)))
}

func MoveToString(move Move) string {
	promoSuffix := ""
	if GetFlag(move) == queenPromoCaptureFlag || GetFlag(move) == queenPromotionFlag {
		promoSuffix = "q"
	} else if GetFlag(move) == knightPromoCaptureFlag || GetFlag(move) == knightPromotionFlag {
		promoSuffix = "n"
	} else if GetFlag(move) == rookPromoCaptureFlag || GetFlag(move) == rookPromotionFlag {
		promoSuffix = "r"
	} else if GetFlag(move) == bishopPromoCaptureFlag || GetFlag(move) == bishopPromotionFlag {
		promoSuffix = "b"
	}
	return fmt.Sprintf("%s%s%s", PositionToSquare(GetStartingPosition(move)), PositionToSquare(GetTargetPosition(move)), promoSuffix)
}
