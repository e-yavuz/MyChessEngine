package chessengine

import (
	"fmt"
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

type Move struct {
	enc uint16
	// 0-5: from
	// 6-11: to
	// 12-15: flags

	priority int8
	// used in move ordering
}

var NULL_MOVE = Move{enc: 0, priority: 0}

type Position = byte
type Flag = uint16

func positionToSquare(position Position) string {
	col := position & 0b111
	row := (position >> 3) & 0b111

	return string(rune('a'+col)) + string(rune('1'+row))
}

// Returns a bitboard of 1's between the given two positions, (excluding the two positions) if they are along a diagonal, row, or column
func getIntermediaryRay(from, to Position) BitBoard {
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

func newMove(from, to Position, flag uint16) Move {
	return Move{enc: uint16(from) + (uint16(to) << 6) + (flag << 12)}
}

func getStartingPosition(move Move) Position {
	return Position(move.enc & 0x3F)
}

func getTargetPosition(move Move) Position {
	return Position((move.enc & 0xFC0) >> 6)
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
	case bishopPromotionFlag:
		return "bishopPromotionFlag"
	case rookPromotionFlag:
		return "rookPromotionFlag"
	case queenPromotionFlag:
		return "queenPromotionFlag"
	case knightPromoCaptureFlag:
		return "knightPromoCaptureFlag"
	case bishopPromoCaptureFlag:
		return "bishopPromoCaptureFlag"
	case rookPromoCaptureFlag:
		return "rookPromoCaptureFlag"
	case queenPromoCaptureFlag:
		return "queenPromoCaptureFlag"
	default:
		return "InvalidFlag"
	}
}

/*
FLAGS
  - quietFlag              Flag = 0b0000
  - doublePawnPushFlag     Flag = 0b0001
  - kingCastleFlag         Flag = 0b0010
  - queenCastleFlag        Flag = 0b0011
  - captureFlag            Flag = 0b0100
  - epCaptureFlag          Flag = 0b0101
  - knightPromotionFlag    Flag = 0b1000
  - rookPromotionFlag      Flag = 0b1001
  - bishopPromotionFlag    Flag = 0b1010
  - queenPromotionFlag     Flag = 0b1011
  - knightPromoCaptureFlag Flag = 0b1100
  - rookPromoCaptureFlag   Flag = 0b1101
  - bishopPromoCaptureFlag Flag = 0b1110
  - queenPromoCaptureFlag  Flag = 0b1111
*/
func GetFlag(move Move) Flag {
	return move.enc >> 12
}

func isTacticalMove(move Move) bool {
	return GetFlag(move) >= captureFlag
}

// Trys to make a move by generating possible moves at ply 1, then checking if the move in UCI format
// is in list by seeing if the possible moves -> UCI == moveUCI (i.e. e2e4 is in the list)
// returns true if move is in list and makes the move
func (board *Board) TryMoveUCI(move string) (Move, bool) {
	moveList := make([]Move, 0, MAX_MOVE_COUNT)
	moveList = board.GenerateMoves(ALL, moveList)

	for _, possibleMove := range moveList {
		if MoveToString(possibleMove) == move {
			return possibleMove, true
		}
	}
	return NULL_MOVE, false
}

// Generates all possible moves for the current board state
// and returns [true/false] if the move is in the list
func (board *Board) validMove(move Move) bool {
	if move == NULL_MOVE {
		return false
	}
	moveList := make([]Move, 0, MAX_MOVE_COUNT)
	moveList = board.GenerateMoves(ALL, moveList)

	for _, possibleMove := range moveList {
		if possibleMove.enc == move.enc {
			return true
		}
	}
	return false
}

// Invariant: Assumes move is legal
func (board *Board) MakeMove(move Move) {

	from := getStartingPosition(move)
	to := getTargetPosition(move)

	currentState := board.GetTopState()

	// Copy over from currentState to a new state
	st := &StateInfo{
		IsWhiteTurn:       !currentState.IsWhiteTurn,
		HalfMoveClock:     currentState.HalfMoveClock + 1,
		TurnCounter:       currentState.TurnCounter,
		ZobristKey:        currentState.ZobristKey,
		CastleState:       currentState.CastleState,
		useOpeningBook:    currentState.useOpeningBook,
		EnPassantPosition: INVALID_POSITION,
		PrecedentMove:     move,
		inCheck:           currentState.inCheck,
	}

	if !currentState.IsWhiteTurn {
		st.TurnCounter += 1
	}

	if move == NULL_MOVE {
		board.pushNewState(st)
		st.inCheck = board.isCheck()
		return
	}

	piece := board.PieceInfoArr[from]
	if piece == nil {
		panic(fmt.Sprintf("%s begins on empty square BestMove: %s", MoveToString(move), MoveToString(pv[0])))
	}

	if piece.thisBitBoard == &board.B.Pawn || piece.thisBitBoard == &board.W.Pawn {
		st.HalfMoveClock = 0
	}

	// Start by removing this piece from the bitBoard it was originally, it is essentially in limbo now
	removeFromBitBoard(piece.thisBitBoard, getStartingPosition(move))

	// Now need to check flags to see what is happening!
	switch GetFlag(move) {
	case doublePawnPushFlag:
		st.EnPassantPosition = (from + to) / 2
	case kingCastleFlag:
		if piece.isWhite {
			st.setCastleWKing(false)
			st.setCastleWQueen(false)
			board.swapPiecePositions(&board.W.Rook, to+1, to-1)
		} else {
			st.setCastleBKing(false)
			st.setCastleBQueen(false)
			board.swapPiecePositions(&board.B.Rook, to+1, to-1)
		}
	case queenCastleFlag:
		if piece.isWhite {
			st.setCastleWKing(false)
			st.setCastleWQueen(false)
			board.swapPiecePositions(&board.W.Rook, to-2, to+1)
		} else {
			st.setCastleBKing(false)
			st.setCastleBQueen(false)
			board.swapPiecePositions(&board.B.Rook, to-2, to+1)
		}
	}

	// Capture move
	if captureFlag&GetFlag(move) > 0 {
		if GetFlag(move) == epCaptureFlag {
			if currentState.IsWhiteTurn {
				removeFromBitBoard(&board.B.Pawn, to-8)
				st.Capture = board.PieceInfoArr[to-8]
				board.PieceInfoArr[to-8] = nil
			} else {
				removeFromBitBoard(&board.W.Pawn, to+8)
				st.Capture = board.PieceInfoArr[to+8]
				board.PieceInfoArr[to+8] = nil
			}
		} else {
			removeFromBitBoard(board.PieceInfoArr[to].thisBitBoard, to)
			st.Capture = board.PieceInfoArr[to]
		}
		// Rook Castle Partner captured -> Cannot Castle that side
		if currentState.getCastleWKing() && to == 7 {
			st.setCastleWKing(false)
		} else if currentState.getCastleWQueen() && to == 0 {
			st.setCastleWQueen(false)
		} else if currentState.getCastleBKing() && to == 63 {
			st.setCastleBKing(false)
		} else if currentState.getCastleBQueen() && to == 56 {
			st.setCastleBQueen(false)
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
			piece.pieceTYPE = KNIGHT
		case 1: // Rook
			if piece.isWhite {
				piece.thisBitBoard = &board.W.Rook
			} else {
				piece.thisBitBoard = &board.B.Rook
			}
			piece.pieceTYPE = ROOK
		case 2: // Bishop
			if piece.isWhite {
				piece.thisBitBoard = &board.W.Bishop
			} else {
				piece.thisBitBoard = &board.B.Bishop
			}
			piece.pieceTYPE = BISHOP
		case 3: // Queen
			if piece.isWhite {
				piece.thisBitBoard = &board.W.Queen
			} else {
				piece.thisBitBoard = &board.B.Queen
			}
			piece.pieceTYPE = QUEEN
		}
	}
	placeOnBitBoard(piece.thisBitBoard, to)
	// Rooks moving -> Castle side negated
	// Hard coded because faster than map lookup?
	if st.getCastleWKing() && piece.pieceTYPE == ROOK && from == 7 {
		st.setCastleWKing(false)
	} else if st.getCastleWQueen() && piece.pieceTYPE == ROOK && from == 0 {
		st.setCastleWQueen(false)
	} else if st.getCastleBKing() && piece.pieceTYPE == ROOK && from == 63 {
		st.setCastleBKing(false)
	} else if st.getCastleBQueen() && piece.pieceTYPE == ROOK && from == 56 {
		st.setCastleBQueen(false)
	}
	// King moving -> Both castle sides negated
	if (st.getCastleWKing() || st.getCastleWQueen()) && piece.thisBitBoard == &board.W.King {
		st.setCastleWKing(false)
		st.setCastleWQueen(false)
	}
	if (st.getCastleBKing() || st.getCastleBQueen()) && piece.thisBitBoard == &board.B.King {
		st.setCastleBKing(false)
		st.setCastleBQueen(false)
	}

	board.pushNewState(st)
	st.inCheck = board.isCheck()
	board.updateZobristHash()
	board.RepetitionPositionHistory[board.GetTopState().ZobristKey] += 1
}

func (board *Board) UnMakeMove() {
	topState := board.PopTopState()
	move := topState.PrecedentMove
	if move == NULL_MOVE {
		return
	}

	board.RepetitionPositionHistory[topState.ZobristKey] -= 1

	from := getStartingPosition(move)
	to := getTargetPosition(move)

	piece := board.PieceInfoArr[to]

	if piece == nil || from == to {
		panic(fmt.Sprintf("%s begins on empty square", MoveToString(move)))
	}

	// Start by removing this piece from the bitBoard it was originally, it is essentially in limbo now
	removeFromBitBoard(piece.thisBitBoard, to)

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
				placeOnBitBoard(topState.Capture.thisBitBoard, to+8)
			} else {
				board.PieceInfoArr[to-8] = topState.Capture
				placeOnBitBoard(topState.Capture.thisBitBoard, to-8)
			}
		} else {
			board.PieceInfoArr[to] = topState.Capture
			placeOnBitBoard(topState.Capture.thisBitBoard, to)
		}
	} else {
		board.PieceInfoArr[to] = nil
	}

	// If promoted, get the bitboard the piece belonged to pre-promotion
	if GetFlag(move)&0b1000 > 0 {
		piece.thisBitBoard = topState.PrePromotionBitBoard
		piece.pieceTYPE = PAWN
	}

	// Place back onto original bitboard position
	placeOnBitBoard(piece.thisBitBoard, from)
}

func (board *Board) swapPiecePositions(bitboard *BitBoard, startPosition, targetPosition Position) {
	removeFromBitBoard(bitboard, startPosition)
	placeOnBitBoard(bitboard, targetPosition)
	board.PieceInfoArr[targetPosition] = board.PieceInfoArr[startPosition]
	board.PieceInfoArr[startPosition] = nil
}

func MoveToStringWithFlag(move Move) string {
	return fmt.Sprintf("%s%s, flag: %s", positionToSquare(getStartingPosition(move)), positionToSquare(getTargetPosition(move)), FlagToString(GetFlag(move)))
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
	return fmt.Sprintf("%s%s%s", positionToSquare(getStartingPosition(move)), positionToSquare(getTargetPosition(move)), promoSuffix)
}

func EncToString(enc uint16) string {
	return fmt.Sprintf("%s%s", positionToSquare(Position(enc&0x3F)), positionToSquare(Position((enc&0xFC0)>>6)))
}

// func compareMove(a, b Move) int {
// 	return int(a.priority - b.priority)
// }
