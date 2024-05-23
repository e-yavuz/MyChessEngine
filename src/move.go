package chessengine

/*
	Contains
		info for move + all possible flags
		Constant indicies for each possible square
*/

const (
	A1 = iota
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
	Encoding uint16
}

func NewMove(from, to uint16, flag uint16) *Move {
	return &Move{
		Encoding: (from & 0b111111) + ((to & 0b111111) << 6) + (flag << 12),
	}
}

func (move *Move) GetStartingPosition() byte {
	return byte(move.Encoding & 0b111111)
}

func (move *Move) GetTargetPosition() byte {
	return byte((move.Encoding >> 6) & 0b111111)
}

/*
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
	quietFlag              = 0b0000
	doublePawnPushFlag     = 0b0001
	kingCastleFlag         = 0b0010
	queenCastleFlag        = 0b0011
	captureFlag            = 0b0100
	epCaptureFlag          = 0b0101
	knightPromotionFlag    = 0b1000
	queenPromotionFlag     = 0b1011
	knightPromoCaptureFlag = 0b1100
	queenPromoCaptureFlag  = 0b1111
)

func (move *Move) GetFlag() byte {
	return byte(move.Encoding >> 12)
}

// Invariant: Assumes move is legal
func (board *Board) MakeMove(move *Move) {

	from := move.GetStartingPosition()
	to := move.GetTargetPosition()

	piece := board.PieceInfoArr[from]
	prev := board.GetTopState()

	//Copy over from prev to a new state
	st := &StateInfo{
		IsWhiteTurn:       !prev.IsWhiteTurn,
		DrawCounter:       prev.DrawCounter + 1,
		TurnCounter:       prev.TurnCounter,
		CastleWKing:       prev.CastleWKing,
		CastleBKing:       prev.CastleBKing,
		CastleWQueen:      prev.CastleWQueen,
		CastleBQueen:      prev.CastleBQueen,
		EnPassantPosition: NullPosition,
	}

	if !prev.IsWhiteTurn {
		st.TurnCounter += 1
		if piece.ThisBitBoard == &board.Bpawn {
			st.DrawCounter = 0
		}
	} else if piece.ThisBitBoard == &board.Wpawn {
		st.DrawCounter = 0
	}

	//Start by removing this piece from the bitBoard it was originally, it is essentially in limbo now
	piece.ThisBitBoard.RemoveFromBitBoard(move.GetStartingPosition())

	//Now need to check flags to see what is happening!
	switch move.GetFlag() {
	case doublePawnPushFlag:
		st.EnPassantPosition = (from + to) / 2
	case kingCastleFlag:
		if piece.IsWhite {
			st.CastleWKing = false
			st.CastleWQueen = false
			board.swapPiecePositions(&board.Wrook, to+1, to-1)
		} else {
			st.CastleBKing = false
			st.CastleBQueen = false
			board.swapPiecePositions(&board.Brook, to+1, to-1)
		}
	case queenCastleFlag:
		if piece.IsWhite {
			st.CastleWKing = false
			st.CastleWQueen = false
			board.swapPiecePositions(&board.Wrook, to-1, to+1)
		} else {
			st.CastleBKing = false
			st.CastleBQueen = false
			board.swapPiecePositions(&board.Brook, to-1, to+1)
		}
	}

	if captureFlag&move.GetFlag() > 0 {
		board.PieceInfoArr[to].ThisBitBoard.RemoveFromBitBoard(to)
		st.Capture = board.PieceInfoArr[to]
		st.DrawCounter = 0
	}

	board.PieceInfoArr[to] = piece
	board.PieceInfoArr[from] = nil

	if move.GetFlag()&0b1000 > 0 {
		st.PrePromotionBitBoard = piece.ThisBitBoard
		switch move.GetFlag() & 1 {
		case 0: //Knight
			if piece.IsWhite {
				piece.ThisBitBoard = &board.Wknight
			} else {
				piece.ThisBitBoard = &board.Bknight
			}
		case 1: //Queen
			if piece.IsWhite {
				piece.ThisBitBoard = &board.Wqueen
			} else {
				piece.ThisBitBoard = &board.Bqueen
			}
		}
	}
	piece.ThisBitBoard.PlaceOnBitBoard(to)

	board.PushNewState(st)
}

func (board *Board) UnMakeMove(move *Move) {
	topState := board.PopTopState()

	from := move.GetStartingPosition()
	to := move.GetTargetPosition()

	piece := board.PieceInfoArr[to]

	//Start by removing this piece from the bitBoard it was originally, it is essentially in limbo now
	piece.ThisBitBoard.RemoveFromBitBoard(to)

	//Reverse engineer castling
	switch move.GetFlag() {
	case kingCastleFlag: // From left of king's "to" position to his right
		if piece.IsWhite {
			board.swapPiecePositions(&board.Wrook, to-1, to+1)
		} else {
			board.swapPiecePositions(&board.Brook, to-1, to+1)
		}
	case queenCastleFlag: // From right of king's "to" position to his left
		if piece.IsWhite {
			board.swapPiecePositions(&board.Wrook, to+1, to-1)
		} else {
			board.swapPiecePositions(&board.Brook, to+1, to-1)
		}
	}

	//Put piece on from position in InfoArr
	board.PieceInfoArr[from] = piece

	//If move was a capture, replace back captured piece in InfoArr on "to" and move it onto its bitboard,
	//otherwise simply nil the infoarr spot as nothing exists there now
	if captureFlag&move.GetFlag() > 0 {
		board.PieceInfoArr[to] = topState.Capture
		topState.Capture.ThisBitBoard.PlaceOnBitBoard(to)
	} else {
		board.PieceInfoArr[to] = nil
	}

	//If promoted, get the bitboard the piece belonged to pre-promotion
	if move.GetFlag()&0b1000 > 0 {
		piece.ThisBitBoard = topState.PrePromotionBitBoard
	}

	//Place back onto original bitboard position
	piece.ThisBitBoard.PlaceOnBitBoard(from)

}

func (board *Board) swapPiecePositions(bitboard *BitBoard, startPosition, targetPosition byte) {
	bitboard.RemoveFromBitBoard(startPosition)
	bitboard.PlaceOnBitBoard(targetPosition)
	board.PieceInfoArr[targetPosition] = board.PieceInfoArr[startPosition]
	board.PieceInfoArr[startPosition] = nil
}
