package chessengine

//TODOhigh Make tests for MakeMove and UnMakeMove!

// TODOmed GeneratePsuedoLegalMoves
func GeneratePsuedoLegalMoves() {
	return
}

// TODOmed GenerateLegalMoves
func GenerateLegalMoves() {
	return
}

// Invariant: Assumes move is legal
func (board *Board) MakeMove(move *Move) {

	from := move.GetStartingPosition()
	to := move.GetTargetPosition()

	piece := board.PieceInfoMap[from]
	prev := board.GetTopState()

	//Copy over from prev to a new state
	st := &StateInfo{
		IsWhiteTurn:       !prev.IsWhiteTurn,
		DrawCounter:       prev.DrawCounter + 1,
		TurnCounter:       prev.TurnCounter + 1,
		CastleWKing:       prev.CastleWKing,
		CastleBKing:       prev.CastleBKing,
		CastleWQueen:      prev.CastleWQueen,
		CastleBQueen:      prev.CastleBQueen,
		EnPassantPosition: NullPosition,
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
		board.PieceInfoMap[to].ThisBitBoard.RemoveFromBitBoard(to)
		st.Capture = board.PieceInfoMap[to]
		st.DrawCounter = 0
	}

	board.PieceInfoMap[to] = piece
	delete(board.PieceInfoMap, from)

	if move.GetFlag()&0b1000 > 0 {
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

	piece := board.PieceInfoMap[to]

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

	board.PieceInfoMap[from] = piece

	if captureFlag&move.GetFlag() > 0 {
		board.PieceInfoMap[to] = topState.Capture
		topState.Capture.ThisBitBoard.PlaceOnBitBoard(to)
	} else {
		delete(board.PieceInfoMap, to)
	}

	if move.GetFlag()&0b1000 > 0 {
		piece.ThisBitBoard = topState.PrePromotionBitBoard
	}
	piece.ThisBitBoard.PlaceOnBitBoard(from)

}

func (board *Board) swapPiecePositions(bitboard *BitBoard, startPosition, targetPosition byte) {
	bitboard.RemoveFromBitBoard(startPosition)
	bitboard.PlaceOnBitBoard(targetPosition)
	board.PieceInfoMap[targetPosition] = board.PieceInfoMap[startPosition]
	delete(board.PieceInfoMap, startPosition)
}
