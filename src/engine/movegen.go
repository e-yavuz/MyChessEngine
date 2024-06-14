package chessengine

const (
	CAPTURE = iota
	EVASION
	QUIET
	ALL
)

const (
	PAWN = iota
	KNIGHT
	ROOK
	BISHOP
	QUEEN
	KING
	NULL_PIECE
)

const MAX_MOVE_COUNT = 218
const MAX_CAPTURE_COUNT = 74
const MAX_QUIET_COUNT = MAX_MOVE_COUNT - MAX_CAPTURE_COUNT // 144

func generateSliding(board *Board, thisBitBoard BitBoard, targetBitBoard BitBoard, pieceType int, genType int, moveList *[]Move) {
	currentState := board.GetTopState()
	if thisBitBoard == 0 {
		return
	}

	var flag Flag
	switch genType {
	case CAPTURE:
		flag = captureFlag
	case QUIET:
		flag = quietFlag
	}

	// Sliding pieces variables
	var totalOccupancyBitBoard, friendlyOccupancyBitBoard, enemyOccupancyBitBoard BitBoard

	if currentState.IsWhiteTurn {
		friendlyOccupancyBitBoard = board.W.OccupancyBitBoard()
		enemyOccupancyBitBoard = board.B.OccupancyBitBoard()
	} else {
		friendlyOccupancyBitBoard = board.B.OccupancyBitBoard()
		enemyOccupancyBitBoard = board.W.OccupancyBitBoard()
	}
	totalOccupancyBitBoard = friendlyOccupancyBitBoard | enemyOccupancyBitBoard

	switch pieceType {
	case ROOK:
		for from := PopLSB(&thisBitBoard); from != INVALID_POSITION; from = PopLSB(&thisBitBoard) {
			validPositions := GetRookMoves(from,
				RookMask(from)&totalOccupancyBitBoard) // Possible positions &'s with total occupancy = blockers
			validPositions &= ^friendlyOccupancyBitBoard // Removes possible captures that would be captures of friendly pieces
			// this is a result of including friendly piece captures in potential moves for faster magic BB's
			validPositions &= targetBitBoard
			for to := PopLSB(&validPositions); to != INVALID_POSITION; to = PopLSB(&validPositions) {
				*moveList = append(*moveList, NewMove(from, to, flag))
			}
		}

	case BISHOP:
		for from := PopLSB(&thisBitBoard); from != INVALID_POSITION; from = PopLSB(&thisBitBoard) {
			validPositions := GetBishopMoves(from,
				BishopMask(from)&totalOccupancyBitBoard) // Possible positions &'s with total occupancy = blockers
			validPositions &= ^friendlyOccupancyBitBoard // Removes possible captures that would be captures of friendly pieces
			// this is a result of including friendly piece captures in potential moves for faster magic BB's
			validPositions &= targetBitBoard
			for to := PopLSB(&validPositions); to != INVALID_POSITION; to = PopLSB(&validPositions) {
				*moveList = append(*moveList, NewMove(from, to, flag))
			}
		}

	case QUEEN:
		for from := PopLSB(&thisBitBoard); from != INVALID_POSITION; from = PopLSB(&thisBitBoard) {
			rookMoves := GetRookMoves(from,
				RookMask(from)&totalOccupancyBitBoard)
			bishopMoves := GetBishopMoves(from,
				BishopMask(from)&totalOccupancyBitBoard)
			// queen = rook + bishop!
			validPositions := (rookMoves | bishopMoves) & ^friendlyOccupancyBitBoard & targetBitBoard
			for to := PopLSB(&validPositions); to != INVALID_POSITION; to = PopLSB(&validPositions) {
				*moveList = append(*moveList, NewMove(from, to, flag))
			}
		}
	}

}

func generateNonSliding(board *Board, thisBitBoard BitBoard, targetBitBoard BitBoard, pieceType int, genType int, moveList *[]Move) {
	// Current state of board, includes who's turn it is, any EnPassant possibility, along with Castling Rights
	currentState := board.GetTopState()
	if thisBitBoard == 0 {
		return
	}

	switch pieceType {
	case PAWN:
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
			eastCapture := Shift(thisBitBoard&^Col8Full, pawnPushDirection+E) & targetBitBoard
			westCapture := Shift(thisBitBoard&^Col1Full, pawnPushDirection+W) & targetBitBoard
			// All captures which end up on the top or bottom rows, i.e. a capture + promotion
			eastCapturePromo := eastCapture & (Row8Full | Row1Full)
			westCapturePromo := westCapture & (Row8Full | Row1Full)
			// Consider both knights and queen for possibility to be promotion piece
			moveListHelper(eastCapturePromo, pawnPushDirection+E, knightPromoCaptureFlag, moveList)
			moveListHelper(eastCapturePromo, pawnPushDirection+E, queenPromoCaptureFlag, moveList)
			moveListHelper(eastCapturePromo, pawnPushDirection+E, rookPromoCaptureFlag, moveList)
			moveListHelper(eastCapturePromo, pawnPushDirection+E, bishopPromoCaptureFlag, moveList)
			moveListHelper(westCapturePromo, pawnPushDirection+W, knightPromoCaptureFlag, moveList)
			moveListHelper(westCapturePromo, pawnPushDirection+W, queenPromoCaptureFlag, moveList)
			moveListHelper(westCapturePromo, pawnPushDirection+W, rookPromoCaptureFlag, moveList)
			moveListHelper(westCapturePromo, pawnPushDirection+W, bishopPromoCaptureFlag, moveList)
			// Check if en passant is even possible this turn
			var eastCaptureEnPassant, westCaptureEnPassant BitBoard
			if currentState.EnPassantPosition != INVALID_POSITION {
				eastCaptureEnPassant = Shift(thisBitBoard&^Col8Full, pawnPushDirection+E) & (1 << currentState.EnPassantPosition) // If so see if
				westCaptureEnPassant = Shift(thisBitBoard&^Col1Full, pawnPushDirection+W) & (1 << currentState.EnPassantPosition) // en passant exists on capture spots
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
			singlePush := Shift(thisBitBoard & ^Row1Full & ^Row8Full, pawnPushDirection) & targetBitBoard
			// Consider all moves that land on the non-promotion squares as quiet
			moveListHelper(singlePush&NonPromotionFull, pawnPushDirection, quietFlag, moveList)
			// Consider all moves that land on the promotion squares and flag with promotion flags
			moveListHelper(singlePush&PromotionFull, pawnPushDirection, knightPromotionFlag, moveList)
			moveListHelper(singlePush&PromotionFull, pawnPushDirection, queenPromotionFlag, moveList)
			moveListHelper(singlePush&PromotionFull, pawnPushDirection, bishopPromotionFlag, moveList)
			moveListHelper(singlePush&PromotionFull, pawnPushDirection, rookPromotionFlag, moveList)
			// Take all pawn's on Rows 2 and 7 (eligible for double pawn push)
			// then see if they can push forward one square onto an unoccupied space
			// and that it is into the interior with NonPromotionFull (smart doggo moment :D)
			// Finally, push inwards 1 square again and see if the output is a target space
			doublePush := Shift(Shift(
				thisBitBoard&(Row2Full|Row7Full),
				pawnPushDirection)&NonPromotionFull&^(board.W.OccupancyBitBoard()|board.B.OccupancyBitBoard()),
				pawnPushDirection) & targetBitBoard
			moveListHelper(doublePush, pawnPushDirection*2, doublePawnPushFlag, moveList)
		}

	case KNIGHT:
		NNE := Shift(thisBitBoard, N+N+E) & ^Col1Full & targetBitBoard
		NNW := Shift(thisBitBoard, N+N+W) & ^Col8Full & targetBitBoard

		SSE := Shift(thisBitBoard, S+S+E) & ^Col1Full & targetBitBoard
		SSW := Shift(thisBitBoard, S+S+W) & ^Col8Full & targetBitBoard

		EES := Shift(thisBitBoard, E+E+S) & ^(Col1Full | Shift(Col1Full, E)) & targetBitBoard
		EEN := Shift(thisBitBoard, E+E+N) & ^(Col1Full | Shift(Col1Full, E)) & targetBitBoard

		WWS := Shift(thisBitBoard, W+W+S) & ^(Col8Full | Shift(Col8Full, W)) & targetBitBoard
		WWN := Shift(thisBitBoard, W+W+N) & ^(Col8Full | Shift(Col8Full, W)) & targetBitBoard

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
}

func enemyPieceAttackBitBoard(board *Board) (retval BitBoard) {
	// Current state of board
	currentState := board.GetTopState()
	var enemyPieces Pieces
	var friendlyKing BitBoard
	// Get the opposing enemyPieces that will generate moves
	if currentState.IsWhiteTurn {
		enemyPieces = board.B
		friendlyKing = board.W.King
	} else {
		enemyPieces = board.W
		friendlyKing = board.B.King
	}

	//Pawns
	if enemyPieces.Pawn != 0 {
		var pawnPushDirection Direction
		if currentState.IsWhiteTurn {
			pawnPushDirection = S
		} else {
			pawnPushDirection = N
		}
		retval |= Shift(enemyPieces.Pawn&^Col8Full, pawnPushDirection+E) | Shift(enemyPieces.Pawn&^Col1Full, pawnPushDirection+W)
	}

	//Knights
	if enemyPieces.Knight != 0 {
		retval |= Shift(enemyPieces.Knight, N+N+E) & ^Col1Full                        //NNE
		retval |= Shift(enemyPieces.Knight, N+N+W) & ^Col8Full                        //NNW
		retval |= Shift(enemyPieces.Knight, S+S+E) & ^Col1Full                        //SSE
		retval |= Shift(enemyPieces.Knight, S+S+W) & ^Col8Full                        //SSW
		retval |= Shift(enemyPieces.Knight, E+E+S) & ^(Col1Full | Shift(Col1Full, E)) //EES
		retval |= Shift(enemyPieces.Knight, E+E+N) & ^(Col1Full | Shift(Col1Full, E)) //EEN
		retval |= Shift(enemyPieces.Knight, W+W+S) & ^(Col8Full | Shift(Col8Full, W)) //WWS
		retval |= Shift(enemyPieces.Knight, W+W+N) & ^(Col8Full | Shift(Col8Full, W)) //WWN
	}

	// Sliding enemyPieces variables, total occypancy - our king to prevent moving along a check ray
	totalOccupancyBitBoard := (board.W.OccupancyBitBoard() | board.B.OccupancyBitBoard()) & ^friendlyKing

	// Rook
	if enemyPieces.Rook != 0 {
		locations := enemyPieces.Rook
		for from := PopLSB(&locations); from != INVALID_POSITION; from = PopLSB(&locations) {
			retval |= GetRookMoves(from,
				RookMask(from)&totalOccupancyBitBoard)
		}
	}

	// Bishop
	if enemyPieces.Bishop != 0 {
		locations := enemyPieces.Bishop
		for from := PopLSB(&locations); from != INVALID_POSITION; from = PopLSB(&locations) {
			retval |= GetBishopMoves(from,
				BishopMask(from)&totalOccupancyBitBoard)
		}
	}

	// Queen
	if enemyPieces.Queen != 0 {
		locations := enemyPieces.Queen
		for from := PopLSB(&locations); from != INVALID_POSITION; from = PopLSB(&locations) {
			rookMoves := GetRookMoves(from,
				RookMask(from)&totalOccupancyBitBoard)
			bishopMoves := GetBishopMoves(from,
				BishopMask(from)&totalOccupancyBitBoard)
			retval |= (rookMoves | bishopMoves)
		}
	}

	// King
	retval |= Shift(enemyPieces.King, N) |
		Shift(enemyPieces.King, S) |
		(Shift(enemyPieces.King, E) & ^Col1Full) |
		(Shift(enemyPieces.King, W) & ^Col8Full) |
		(Shift(enemyPieces.King, N+E) & ^Col1Full) |
		(Shift(enemyPieces.King, N+W) & ^Col8Full) |
		(Shift(enemyPieces.King, S+E) & ^Col1Full) |
		(Shift(enemyPieces.King, S+W) & ^Col8Full)

	return retval
}

func generateKing(board *Board, targetBitBoard BitBoard, genType int, inCheck bool, moveList *[]Move) {
	// Current state of board, includes who's turn it is, any EnPassant possibility, along with Castling Rights
	currentState := board.GetTopState()

	var pieces Pieces
	var from, to Position
	var castleKing, castleQueen bool
	// Get the pieces that will generate moves
	if currentState.IsWhiteTurn {
		pieces = board.W
		castleKing = currentState.getCastleWKing()
		castleQueen = currentState.getCastleWQueen()
	} else {
		pieces = board.B
		castleKing = currentState.getCastleBKing()
		castleQueen = currentState.getCastleBQueen()
	}
	{
		temp := pieces.King
		from = PopLSB(&temp)
	}

	validPositions := Shift(pieces.King, N) |
		Shift(pieces.King, S) |
		(Shift(pieces.King, E) & ^Col1Full) |
		(Shift(pieces.King, W) & ^Col8Full) |
		(Shift(pieces.King, N+E) & ^Col1Full) |
		(Shift(pieces.King, N+W) & ^Col8Full) |
		(Shift(pieces.King, S+E) & ^Col1Full) |
		(Shift(pieces.King, S+W) & ^Col8Full)
	validPositions &= targetBitBoard

	var kingFlag Flag
	switch genType {
	case CAPTURE:
		kingFlag = captureFlag
	case QUIET:
		kingFlag = quietFlag

		// Check king can move at least 2 to the right (emptiness included implicitly)
		if !inCheck &&
			castleKing &&
			(GetIntermediaryRay(from, from+3)&targetBitBoard) == GetIntermediaryRay(from, from+3) {
			*moveList = append(*moveList, NewMove(from, from+2, kingCastleFlag))
		}
		// Check queen side castle empty + king can move at least 2 to the left
		if !inCheck &&
			castleQueen &&
			(GetIntermediaryRay(from, from-4)&(board.W.OccupancyBitBoard()|board.B.OccupancyBitBoard())) == GetIntermediaryRay(from, from-4) &&
			(GetIntermediaryRay(from, from-3)&targetBitBoard) == GetIntermediaryRay(from, from-3) {
			*moveList = append(*moveList, NewMove(from, from-2, queenCastleFlag))
		}
	}

	for to = PopLSB(&validPositions); to != INVALID_POSITION; to = PopLSB(&validPositions) {
		*moveList = append(*moveList, NewMove(from, to, kingFlag))
	}
}

/*
Scan from king according to https://peterellisjones.com/posts/generating-legal-chess-moves-efficiently/
for hop pieces like pawn and knight, just add to checkers list of []PieceInfofor sliding
pieces, send out rays from the king in all 4+4 directions (rook then bishop)scan for rook/queen,
then bishop/queen along the rays (no mask)if there is a piece,take the ray seperating the king and
sliding enemy piece, and &it with friendly occupancy board, if there is 0 pieces, the sliding piece
goes into the checkers listif only 1 friendly piece, then add that piece to pinned pieces list of
[]PieceInfo if 2 or more, then ignore the sliding piece!
*/
func generateCheck(board *Board) (pinnedPieces *[]PinnedPieceInfo, pinnedPiecesBitBoard BitBoard, checkingPieces *[]CheckerInfo) {
	// Current state of board, includes who's turn it is, any EnPassant possibility, along with Castling Rights
	currentState := board.GetTopState()

	checkingPieces = new([]CheckerInfo)
	pinnedPieces = new([]PinnedPieceInfo)

	var enemyPieces, friendlyPieces *Pieces
	var from, to Position
	var friendlyKing BitBoard
	// Get the pieces that will generate moves
	if currentState.IsWhiteTurn {
		friendlyPieces = &board.W
		enemyPieces = &board.B
	} else {
		friendlyPieces = &board.B
		enemyPieces = &board.W
	}

	{
		friendlyKing = friendlyPieces.King
		temp := friendlyKing
		from = PopLSB(&temp)
	}

	// Idea is to treat king as every possible non-king enemy piece,
	// then seeing if a reversal of the capture is possible -> "captured" piece has the king in check

	// Pawn
	if enemyPieces.Pawn != 0 {
		var pawnPushDirection Direction
		if currentState.IsWhiteTurn {
			pawnPushDirection = N
		} else {
			pawnPushDirection = S
		}
		// Check for possible pawn captures
		// inverting by side columns to prevent going off the board
		validCaptures := (Shift(friendlyKing, pawnPushDirection+E) & enemyPieces.Pawn &^ Col1Full) |
			(Shift(friendlyKing, pawnPushDirection+W) & enemyPieces.Pawn &^ Col8Full)
		for to = PopLSB(&validCaptures); to != INVALID_POSITION; to = PopLSB(&validCaptures) {
			*checkingPieces = append(*checkingPieces,
				CheckerInfo{
					*board.PieceInfoArr[to], to, 0})
			PlaceOnBitBoard(&pinnedPiecesBitBoard, to)
		}
	}

	// Knight
	if enemyPieces.Knight != 0 {
		validCaptures := ((Shift(friendlyKing, N+N+E) & ^Col1Full) | //NNE
			(Shift(friendlyKing, N+N+W) & ^Col8Full) | //NNW
			(Shift(friendlyKing, S+S+E) & ^Col1Full) | //SSE
			(Shift(friendlyKing, S+S+W) & ^Col8Full) | //SSW
			(Shift(friendlyKing, E+E+S) & ^(Col1Full | Shift(Col1Full, E))) | //EES
			(Shift(friendlyKing, E+E+N) & ^(Col1Full | Shift(Col1Full, E))) | //EEN
			(Shift(friendlyKing, W+W+S) & ^(Col8Full | Shift(Col8Full, W))) | //WWS
			(Shift(friendlyKing, W+W+N) & ^(Col8Full | Shift(Col8Full, W)))) & //WWN
			enemyPieces.Knight // All possible knight positions
		for to = PopLSB(&validCaptures); to != INVALID_POSITION; to = PopLSB(&validCaptures) {
			*checkingPieces = append(*checkingPieces,
				CheckerInfo{
					*board.PieceInfoArr[to], to, 0})
			PlaceOnBitBoard(&pinnedPiecesBitBoard, to)
		}
	}

	friendlyOccupancyBitBoard := friendlyPieces.OccupancyBitBoard()
	enemyOccupancyBitBoard := enemyPieces.OccupancyBitBoard()

	// Rook/Queen
	if enemyPieces.Rook != 0 || enemyPieces.Queen != 0 {
		possibleCheckers := GetRookMoves(from, RookMask(from)&enemyOccupancyBitBoard)
		possibleCheckers &= enemyPieces.Queen | enemyPieces.Rook

		for to = PopLSB(&possibleCheckers); to != INVALID_POSITION; to = PopLSB(&possibleCheckers) {
			checkRay := GetIntermediaryRay(from, to)
			checkerPieceInfo := CheckerInfo{*board.PieceInfoArr[to], to, checkRay}

			checkRayPinned := checkRay & friendlyOccupancyBitBoard
			friendlyPinnedPosition := PopLSB(&checkRayPinned)

			if friendlyPinnedPosition == INVALID_POSITION { // No pinned pieces -> + to CheckerInfo
				*checkingPieces = append(*checkingPieces, checkerPieceInfo)
			} else if PopLSB(&checkRayPinned) != INVALID_POSITION { // 2+ "pinned" pieces -> go to next potential checker
				continue
			} else { // 1 pinned piece -> + to PinnedInfo
				*pinnedPieces = append(*pinnedPieces,
					PinnedPieceInfo{
						checkerPieceInfo,
						*board.PieceInfoArr[friendlyPinnedPosition],
						friendlyPinnedPosition,
						checkRay & ^(BitBoard(1) << friendlyPinnedPosition)})
				PlaceOnBitBoard(&pinnedPiecesBitBoard, friendlyPinnedPosition)
			}
		}
	}

	// Bishop/Queen
	if enemyPieces.Bishop != 0 || enemyPieces.Queen != 0 {
		possibleCheckers := GetBishopMoves(from, BishopMask(from)&enemyOccupancyBitBoard)
		possibleCheckers &= enemyPieces.Queen | enemyPieces.Bishop

		for to = PopLSB(&possibleCheckers); to != INVALID_POSITION; to = PopLSB(&possibleCheckers) {
			checkRay := GetIntermediaryRay(from, to)
			checkerPieceInfo := CheckerInfo{*board.PieceInfoArr[to], to, checkRay}

			checkRayPinned := checkRay & friendlyOccupancyBitBoard
			friendlyPinnedPosition := PopLSB(&checkRayPinned)

			if friendlyPinnedPosition == INVALID_POSITION { // No pinned pieces -> + to CheckerInfo
				*checkingPieces = append(*checkingPieces, checkerPieceInfo)
			} else if PopLSB(&checkRayPinned) != INVALID_POSITION { // 2+ "pinned" pieces -> go to next potential checker
				continue
			} else { // 1 pinned piece -> + to PinnedInfo
				*pinnedPieces = append(*pinnedPieces,
					PinnedPieceInfo{
						checkerPieceInfo,
						*board.PieceInfoArr[friendlyPinnedPosition],
						friendlyPinnedPosition,
						checkRay & ^(BitBoard(1) << friendlyPinnedPosition)})
				PlaceOnBitBoard(&pinnedPiecesBitBoard, friendlyPinnedPosition)
			}
		}
	}

	return pinnedPieces, pinnedPiecesBitBoard, checkingPieces
}

func generatePinned(board *Board, genType int, pinnedPieces *[]PinnedPieceInfo, moveList *[]Move) {
	for _, pinnedPiece := range *pinnedPieces {
		thisBitBoard := BitBoard(1) << pinnedPiece.position
		var targetBitBoard BitBoard

		switch genType {
		case CAPTURE:
			targetBitBoard = BitBoard(1) << pinnedPiece.checkerInfo.position
		case QUIET:
			targetBitBoard = pinnedPiece.possibleQuiet
		}

		switch pinnedPiece.pieceInfo.pieceTYPE {
		case PAWN:
			generateNonSliding(board, thisBitBoard, targetBitBoard, PAWN, genType, moveList)
		case KNIGHT:
			generateNonSliding(board, thisBitBoard, targetBitBoard, KNIGHT, genType, moveList)
		case ROOK:
			generateSliding(board, thisBitBoard, targetBitBoard, ROOK, genType, moveList)
		case BISHOP:
			generateSliding(board, thisBitBoard, targetBitBoard, BISHOP, genType, moveList)
		case QUEEN:
			generateSliding(board, thisBitBoard, targetBitBoard, QUEEN, genType, moveList)
		}
	}
}

func (board *Board) GenerateMoves(genType int, moveList []Move) []Move {
	if genType == ALL {
		moveList = board.GenerateMoves(CAPTURE, moveList)
		moveList = board.GenerateMoves(QUIET, moveList)
		return moveList
	}
	currentState := board.GetTopState()

	var friendlyPieces, enemyPieces *Pieces
	var targetBitBoard BitBoard

	if currentState.IsWhiteTurn {
		enemyPieces = &board.B
		friendlyPieces = &board.W
	} else {
		enemyPieces = &board.W
		friendlyPieces = &board.B
	}

	if friendlyPieces.King == 0 || enemyPieces.King == 0 {
		return []Move{}
	}

	pinnedPieces, pinnedPiecesBitBoard, checkingPieces := generateCheck(board)
	inCheck := len(*checkingPieces) > 0

	switch len(*checkingPieces) {
	case 0:
		// No checkers
		switch genType {
		case CAPTURE:
			targetBitBoard = enemyPieces.OccupancyBitBoard()
		case QUIET:
			targetBitBoard = ^(friendlyPieces.OccupancyBitBoard() | enemyPieces.OccupancyBitBoard())
		}
		generatePinned(board, genType, pinnedPieces, &moveList)
		generateSliding(board, friendlyPieces.Queen&^pinnedPiecesBitBoard, targetBitBoard, QUEEN, genType, &moveList)
		generateSliding(board, friendlyPieces.Bishop&^pinnedPiecesBitBoard, targetBitBoard, BISHOP, genType, &moveList)
		generateSliding(board, friendlyPieces.Rook&^pinnedPiecesBitBoard, targetBitBoard, ROOK, genType, &moveList)
		generateNonSliding(board, friendlyPieces.Knight&^pinnedPiecesBitBoard, targetBitBoard, KNIGHT, genType, &moveList)
		generateNonSliding(board, friendlyPieces.Pawn&^pinnedPiecesBitBoard, targetBitBoard, PAWN, genType, &moveList)
	case 1:
		// 1 Checker, quiet target bitboard is intermediary ray,
		// capture target bitboard is the checker
		switch genType {
		case CAPTURE:
			targetBitBoard = BitBoard(1) << (*checkingPieces)[0].position
		case QUIET:
			targetBitBoard = (*checkingPieces)[0].intermediaryRay
		}
		generateSliding(board, friendlyPieces.Queen&^pinnedPiecesBitBoard, targetBitBoard, QUEEN, genType, &moveList)
		generateSliding(board, friendlyPieces.Bishop&^pinnedPiecesBitBoard, targetBitBoard, BISHOP, genType, &moveList)
		generateSliding(board, friendlyPieces.Rook&^pinnedPiecesBitBoard, targetBitBoard, ROOK, genType, &moveList)
		generateNonSliding(board, friendlyPieces.Knight&^pinnedPiecesBitBoard, targetBitBoard, KNIGHT, genType, &moveList)
		generateNonSliding(board, friendlyPieces.Pawn&^pinnedPiecesBitBoard, targetBitBoard, PAWN, genType, &moveList)

	}
	switch genType {
	case CAPTURE:
		generateKing(board,
			enemyPieces.OccupancyBitBoard() & ^enemyPieceAttackBitBoard(board),
			CAPTURE,
			inCheck,
			&moveList)
	case QUIET:
		generateKing(board,
			^(friendlyPieces.OccupancyBitBoard()|enemyPieces.OccupancyBitBoard()) & ^enemyPieceAttackBitBoard(board),
			QUIET,
			inCheck,
			&moveList)
	}

	return moveList
}

// Helper function to generate moves for a bitboard of pieces given a static direction
func moveListHelper(bitboard BitBoard, moveDir Direction, flag Flag, moveList *[]Move) {
	for to := PopLSB(&bitboard); to != INVALID_POSITION; to = PopLSB(&bitboard) {
		*moveList = append(*moveList, NewMove(to-Position(moveDir), to, flag))
	}
}

// Simplified generateCheck to reduce computation as there is no need for pinned pieces, etc... used in search
func (board *Board) InCheck() bool {
	// Current state of board, includes who's turn it is, any EnPassant possibility, along with Castling Rights
	currentState := board.GetTopState()

	var enemyPieces, friendlyPieces *Pieces
	var from Position
	var friendlyKing BitBoard
	// Get the pieces that will generate moves
	if currentState.IsWhiteTurn {
		friendlyPieces = &board.W
		enemyPieces = &board.B
	} else {
		friendlyPieces = &board.B
		enemyPieces = &board.W
	}

	{
		friendlyKing = friendlyPieces.King
		if friendlyKing == 0 {
			return true
		}
		temp := friendlyKing
		from = PopLSB(&temp)
	}

	// Idea is to treat king as every possible non-king enemy piece,
	// then seeing if a reversal of the capture is possible -> "captured" piece has the king in check

	// Pawn
	if enemyPieces.Pawn != 0 {
		var pawnPushDirection Direction
		if currentState.IsWhiteTurn {
			pawnPushDirection = N
		} else {
			pawnPushDirection = S
		}
		// Check for possible pawn captures
		// inverting by side columns to prevent going off the board
		validCaptures := (Shift(friendlyKing, pawnPushDirection+E) & enemyPieces.Pawn &^ Col1Full) |
			(Shift(friendlyKing, pawnPushDirection+W) & enemyPieces.Pawn &^ Col8Full)
		if validCaptures > 0 {
			return true
		}
	}

	// Knight
	if enemyPieces.Knight != 0 {
		validCaptures := ((Shift(friendlyKing, N+N+E) & ^Col1Full) | //NNE
			(Shift(friendlyKing, N+N+W) & ^Col8Full) | //NNW
			(Shift(friendlyKing, S+S+E) & ^Col1Full) | //SSE
			(Shift(friendlyKing, S+S+W) & ^Col8Full) | //SSW
			(Shift(friendlyKing, E+E+S) & ^(Col1Full | Shift(Col1Full, E))) | //EES
			(Shift(friendlyKing, E+E+N) & ^(Col1Full | Shift(Col1Full, E))) | //EEN
			(Shift(friendlyKing, W+W+S) & ^(Col8Full | Shift(Col8Full, W))) | //WWS
			(Shift(friendlyKing, W+W+N) & ^(Col8Full | Shift(Col8Full, W)))) & //WWN
			enemyPieces.Knight // All possible knight positions
		if validCaptures > 0 {
			return true
		}
	}

	totalOccupancyBitBoard := enemyPieces.OccupancyBitBoard() | friendlyPieces.OccupancyBitBoard()

	// Rook/Queen
	if enemyPieces.Rook != 0 || enemyPieces.Queen != 0 {
		possibleCheckers := GetRookMoves(from, RookMask(from)&totalOccupancyBitBoard)
		possibleCheckers &= enemyPieces.Queen | enemyPieces.Rook

		if possibleCheckers > 0 {
			return true
		}
	}

	// Bishop/Queen
	if enemyPieces.Bishop != 0 || enemyPieces.Queen != 0 {
		possibleCheckers := GetBishopMoves(from, BishopMask(from)&totalOccupancyBitBoard)
		possibleCheckers &= enemyPieces.Queen | enemyPieces.Bishop

		if possibleCheckers > 0 {
			return true
		}
	}

	return false
}
