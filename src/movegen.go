package chessengine

const (
	CAPTURE = iota
	EVASION
	QUIET
)

// TODOmed GeneratePsuedoLegalMoves
func (board *Board) GeneratePsuedoLegalMoves(targetBitBoard BitBoard, genType int) {
}

// TODOhigh specific generation
func (board *Board) Generate_Specific(targetBitBoard BitBoard, genType int, moveList *[]Move) {
	// Current state of board, includes who's turn it is, any EnPassant possibility, along with Castling Rights
	currentState := board.GetTopState()

	var pieces Pieces
	// var from, to Position
	// Get the pieces that will generate moves
	if currentState.IsWhiteTurn {
		pieces = board.W
	} else {
		pieces = board.B
	}

	// Pawn Logic, only run if there exists pawns
	if pieces.Pawn > 0 {
		// Always internally keep White as the team facing up North
		// thus determine if it is White's turn, pawn's must move North and vice versa
		var pawnPushDirection Direction
		if currentState.IsWhiteTurn {
			pawnPushDirection = N
		} else {
			pawnPushDirection = S
		}

		switch genType { // Either generate only captures or only quiet moves
		case CAPTURE:
			// Identify possible captures moving diagonally
			// inverting by side columns to prevent going off the board
			eastCapture := Shift(pieces.Pawn&^Col8Full, pawnPushDirection+E) & targetBitBoard
			westCapture := Shift(pieces.Pawn&^Col1Full, pawnPushDirection+W) & targetBitBoard
			// All captures which end up on the top or bottom rows, i.e. a capture + promotion
			eastCapturePromo := eastCapture & Row8Full & Row1Full
			westCapturePromo := westCapture & Row8Full & Row1Full
			// Consider both knights and queen for possibility to be promotion piece
			moveListHelper(eastCapturePromo, pawnPushDirection+E, knightPromoCaptureFlag, moveList)
			moveListHelper(eastCapturePromo, pawnPushDirection+E, queenPromoCaptureFlag, moveList)
			moveListHelper(westCapturePromo, pawnPushDirection+W, knightPromoCaptureFlag, moveList)
			moveListHelper(westCapturePromo, pawnPushDirection+W, queenPromoCaptureFlag, moveList)
			// Check if en passant is even possible this turn
			var eastCaptureEnPassant, westCaptureEnPassant BitBoard
			if currentState.EnPassantPosition != NULL_POSITION {
				eastCaptureEnPassant = Shift(pieces.Pawn&^Col8Full, pawnPushDirection+E) & (1 << currentState.EnPassantPosition) // If so see if
				westCaptureEnPassant = Shift(pieces.Pawn&^Col1Full, pawnPushDirection+W) & (1 << currentState.EnPassantPosition) // en passant exists on capture spots
				moveListHelper(eastCaptureEnPassant, pawnPushDirection+E, epCaptureFlag, moveList)
				moveListHelper(westCaptureEnPassant, pawnPushDirection+W, epCaptureFlag, moveList)
			}
			// Finally consider all remaining capture moves by inverting the special
			// ones along with AND'ing on top of original
			eastCaptureQuiet := eastCapture & ^eastCapturePromo & ^eastCaptureEnPassant
			moveListHelper(eastCaptureQuiet, pawnPushDirection+E, captureFlag, moveList)
			westCaptureQuiet := westCapture & ^westCapturePromo & ^westCaptureEnPassant
			moveListHelper(westCaptureQuiet, pawnPushDirection+W, captureFlag, moveList)
		case QUIET:
			// Consider all pawns not on Row's 1 and 8
			// (idk abt this being needed as you always promote pawns on these rows?)
			// then move them inwards 1 space and AND with possible positons to land on with targetBitBoard
			singlePush := Shift(pieces.Pawn & ^Row1Full & ^Row8Full, pawnPushDirection) & targetBitBoard
			// Consider all moves that land on the non-promotion squares as quiet
			moveListHelper(singlePush&NonPromotionFull, pawnPushDirection, quietFlag, moveList)
			// Consider all moves that land on the promotion squares and flag with knight/queen promotion
			moveListHelper(singlePush&PromotionFull, pawnPushDirection, knightPromotionFlag, moveList)
			moveListHelper(singlePush&PromotionFull, pawnPushDirection, queenPromotionFlag, moveList)
			// Take all pawn's on Rows 2 and 7 (eligible for double pawn push)
			// then see if they can push forward one square onto a valid space
			// and that it is into the interior with NonPromotionFull (smart doggo moment :D)
			// Finally, push inwards 1 square again and see if the output is a valid space to land on
			doublePush := Shift(Shift(
				pieces.Pawn&(Row2Full|Row7Full),
				pawnPushDirection)&targetBitBoard&NonPromotionFull,
				pawnPushDirection) & targetBitBoard
			moveListHelper(doublePush, pawnPushDirection*2, doublePawnPushFlag, moveList)
		}
	}

	// Knight
	if pieces.Knight > 0 {
		NNE := Shift(pieces.Knight, N+N+E) & ^Col1Full & targetBitBoard
		NNW := Shift(pieces.Knight, N+N+W) & ^Col8Full & targetBitBoard

		SSE := Shift(pieces.Knight, S+S+E) & ^Col1Full & targetBitBoard
		SSW := Shift(pieces.Knight, S+S+W) & ^Col8Full & targetBitBoard

		EES := Shift(pieces.Knight, E+E+S) & ^(Col1Full | Shift(Col1Full, E) | Row8Full) & targetBitBoard
		EEN := Shift(pieces.Knight, E+E+N) & ^(Col1Full | Shift(Col1Full, E) | Row1Full) & targetBitBoard

		WWS := Shift(pieces.Knight, W+W+S) & ^(Col1Full | Shift(Col8Full, W) | Row8Full) & targetBitBoard
		WWN := Shift(pieces.Knight, W+W+N) & ^(Col8Full | Shift(Col8Full, W) | Row1Full) & targetBitBoard

		var knightFlag Flag
		switch genType {
		case CAPTURE:
			knightFlag = captureFlag
		case QUIET:
			knightFlag = quietFlag
		}
		moveListHelper(NNE, N+N+E, knightFlag, moveList)
		moveListHelper(NNW, N+N+W, knightFlag, moveList)
		moveListHelper(SSE, S+S+E, knightFlag, moveList)
		moveListHelper(SSW, S+S+W, knightFlag, moveList)
		moveListHelper(EES, E+E+S, knightFlag, moveList)
		moveListHelper(EEN, E+E+N, knightFlag, moveList)
		moveListHelper(WWS, W+W+S, knightFlag, moveList)
		moveListHelper(WWN, W+W+N, knightFlag, moveList)
	}

	// Sliding Pieces?
}

func (board *Board) enemyPieceAttackBitBoard() (retval BitBoard) {
	// Current state of board
	currentState := board.GetTopState()
	var pieces Pieces
	// Get the opposing pieces that will generate moves
	if currentState.IsWhiteTurn {
		pieces = board.B
	} else {
		pieces = board.W
	}

	//Pawns
	{
		var pawnPushDirection Direction
		if currentState.IsWhiteTurn {
			pawnPushDirection = S
		} else {
			pawnPushDirection = N
		}
		retval |= Shift(pieces.Pawn&^Col8Full, pawnPushDirection+E) | Shift(pieces.Pawn&^Col1Full, pawnPushDirection+W)
	}

	//Knights
	retval |= Shift(pieces.Knight, N+N+E) & ^Col1Full                                   //NNE
	retval |= Shift(pieces.Knight, N+N+W) & ^Col8Full                                   //NNW
	retval |= Shift(pieces.Knight, S+S+E) & ^Col1Full                                   //SSE
	retval |= Shift(pieces.Knight, S+S+W) & ^Col8Full                                   //SSW
	retval |= Shift(pieces.Knight, E+E+S) & ^(Col1Full | Shift(Col1Full, E) | Row8Full) //EES
	retval |= Shift(pieces.Knight, E+E+N) & ^(Col1Full | Shift(Col1Full, E) | Row1Full) //EEN
	retval |= Shift(pieces.Knight, W+W+S) & ^(Col1Full | Shift(Col8Full, W) | Row8Full) //WWS
	retval |= Shift(pieces.Knight, W+W+N) & ^(Col8Full | Shift(Col8Full, W) | Row1Full) //WWN

	//TODOmed enemyPieceAttackBitBoard Sliding Pieces

	return retval
}

func (board *Board) Generate_King(targetBitBoard BitBoard, genType int, moveList *[]Move) {
	// // Current state of board, includes who's turn it is, any EnPassant possibility, along with Castling Rights
	// currentState := board.GetTopState()

	// var pieces Pieces
	// // var from, to Position
	// // Get the pieces that will generate moves
	// if currentState.IsWhiteTurn {
	// 	pieces = board.W
	// } else {
	// 	pieces = board.B
	// }

	// kingN := Shift(pieces.King, N) & targetBitBoard
	// kingS := Shift(pieces.King, S) & targetBitBoard
	// kingE := Shift(pieces.King, E) & ^Col1Full
	// kingW := Shift(pieces.King, W) & ^Col8Full
	// kingNE := Shift(kingE, N) & targetBitBoard
	// kingSE := Shift(kingE, S) & targetBitBoard
	// kingNW := Shift(kingW, N) & targetBitBoard
	// kingSW := Shift(kingW, S) & targetBitBoard
	// kingE &= targetBitBoard
	// kingW &= targetBitBoard

	// var kingFlag Flag
	// switch genType {
	// case CAPTURE:
	// 	kingFlag = captureFlag
	// case QUIET:
	// 	kingFlag = quietFlag
	// }
}

// TODOmed GenerateLegalMoves
func GenerateLegalMoves() {

}

func moveListHelper(bitboard BitBoard, moveDir Direction, flag Flag, moveList *[]Move) {
	for to := PopLSB(&bitboard); to != NULL_POSITION; to = PopLSB(&bitboard) {
		*moveList = append(*moveList, NewMove(to, to-Position(moveDir), flag))
	}
}
