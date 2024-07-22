package chessengine

const (
	CAPTURE = iota
	EVASION
	QUIET
	ALL
)

var knightMoveBoard = [64]BitBoard{}
var kingMoveBoard = [64]BitBoard{}

var friendBitboard BitBoard
var enemyBitBoard BitBoard
var totalBitBoard BitBoard

const MAX_MOVE_COUNT = 218
const MAX_CAPTURE_COUNT = 74
const MAX_QUIET_COUNT = MAX_MOVE_COUNT - MAX_CAPTURE_COUNT // 144

func init() {
	for i := 0; i < 64; i++ {
		bitboard := BitBoard(1) << i
		knightMoveBoard[i] = ((Shift(bitboard, N+N+E) & ^Col1Full) | //NNE
			(Shift(bitboard, N+N+W) & ^Col8Full) | //NNW
			(Shift(bitboard, S+S+E) & ^Col1Full) | //SSE
			(Shift(bitboard, S+S+W) & ^Col8Full) | //SSW
			(Shift(bitboard, E+E+S) & ^(Col1Full | Shift(Col1Full, E))) | //EES
			(Shift(bitboard, E+E+N) & ^(Col1Full | Shift(Col1Full, E))) | //EEN
			(Shift(bitboard, W+W+S) & ^(Col8Full | Shift(Col8Full, W))) | //WWS
			(Shift(bitboard, W+W+N) & ^(Col8Full | Shift(Col8Full, W)))) //WWN
		kingMoveBoard[i] = Shift(bitboard, N) |
			Shift(bitboard, S) |
			(Shift(bitboard, E) & ^Col1Full) |
			(Shift(bitboard, W) & ^Col8Full) |
			(Shift(bitboard, N+E) & ^Col1Full) |
			(Shift(bitboard, N+W) & ^Col8Full) |
			(Shift(bitboard, S+E) & ^Col1Full) |
			(Shift(bitboard, S+W) & ^Col8Full)
	}
}

func generateSliding(thisBitBoard BitBoard, targetBitBoard BitBoard, pieceType int, genType int, moveList *[]Move) {
	// currentState := board.GetTopState()
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

	switch pieceType {
	case ROOK:
		for from := PopLSB(&thisBitBoard); from != INVALID_POSITION; from = PopLSB(&thisBitBoard) {
			validPositions := GetRookMoves(from,
				RookMask(from)&totalBitBoard) // Possible positions &'s with total occupancy = blockers
			validPositions &= ^friendBitboard // Removes possible captures that would be captures of friendly pieces
			// this is a result of including friendly piece captures in potential moves for faster magic BB's
			validPositions &= targetBitBoard
			for to := PopLSB(&validPositions); to != INVALID_POSITION; to = PopLSB(&validPositions) {
				*moveList = append(*moveList, NewMove(from, to, flag))
			}
		}

	case BISHOP:
		for from := PopLSB(&thisBitBoard); from != INVALID_POSITION; from = PopLSB(&thisBitBoard) {
			validPositions := GetBishopMoves(from,
				BishopMask(from)&totalBitBoard) // Possible positions &'s with total occupancy = blockers
			validPositions &= ^friendBitboard // Removes possible captures that would be captures of friendly pieces
			// this is a result of including friendly piece captures in potential moves for faster magic BB's
			validPositions &= targetBitBoard
			for to := PopLSB(&validPositions); to != INVALID_POSITION; to = PopLSB(&validPositions) {
				*moveList = append(*moveList, NewMove(from, to, flag))
			}
		}

	case QUEEN:
		for from := PopLSB(&thisBitBoard); from != INVALID_POSITION; from = PopLSB(&thisBitBoard) {
			rookMoves := GetRookMoves(from,
				RookMask(from)&totalBitBoard)
			bishopMoves := GetBishopMoves(from,
				BishopMask(from)&totalBitBoard)
			// queen = rook + bishop!
			validPositions := (rookMoves | bishopMoves) & ^friendBitboard & targetBitBoard
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
		if currentState.TurnColor == WHITE {
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
			moveListHelper(eastCapturePromo, pawnPushDirection+E, moveList,
				knightPromoCaptureFlag, queenPromoCaptureFlag,
				rookPromoCaptureFlag, bishopPromoCaptureFlag)
			moveListHelper(westCapturePromo, pawnPushDirection+W, moveList,
				knightPromoCaptureFlag, queenPromoCaptureFlag,
				rookPromoCaptureFlag, bishopPromoCaptureFlag)
			// Check if en passant is even possible this turn
			var eastCaptureEnPassant, westCaptureEnPassant BitBoard
			if currentState.EnPassantPosition != INVALID_POSITION {
				eastCaptureEnPassant = Shift(thisBitBoard&^Col8Full, pawnPushDirection+E) & (1 << currentState.EnPassantPosition) // If so see if
				westCaptureEnPassant = Shift(thisBitBoard&^Col1Full, pawnPushDirection+W) & (1 << currentState.EnPassantPosition) // en passant exists on capture spots
				moveListHelper(eastCaptureEnPassant, pawnPushDirection+E, moveList, epCaptureFlag)
				moveListHelper(westCaptureEnPassant, pawnPushDirection+W, moveList, epCaptureFlag)
			}
			// Finally consider all remaining capture moves by inverting the special
			// ones along with AND'ing on top of original
			eastCaptureQuiet := eastCapture & ^eastCapturePromo & ^eastCaptureEnPassant
			moveListHelper(eastCaptureQuiet, pawnPushDirection+E, moveList, captureFlag)
			westCaptureQuiet := westCapture & ^westCapturePromo & ^westCaptureEnPassant
			moveListHelper(westCaptureQuiet, pawnPushDirection+W, moveList, captureFlag)
		case QUIET:
			// Consider all pawns not on Row's 1 and 8
			// (idk abt this being needed as you always promote pawns on these rows?)
			// then move them inwards 1 space and AND with possible positons to land on with targetBitBoard
			singlePush := Shift(thisBitBoard & ^Row1Full & ^Row8Full, pawnPushDirection) & targetBitBoard
			// Consider all moves that land on the non-promotion squares as quiet
			moveListHelper(singlePush&NonPromotionFull, pawnPushDirection, moveList, quietFlag)
			// Consider all moves that land on the promotion squares and flag with promotion flags
			moveListHelper(singlePush&PromotionFull, pawnPushDirection, moveList,
				knightPromotionFlag, queenPromotionFlag, rookPromotionFlag, bishopPromotionFlag)
			// Take all pawn's on Rows 2 and 7 (eligible for double pawn push)
			// then see if they can push forward one square onto an unoccupied space
			// and that it is into the interior with NonPromotionFull (smart doggo moment :D)
			// Finally, push inwards 1 square again and see if the output is a target space
			doublePush := Shift(Shift(
				thisBitBoard&(Row2Full|Row7Full),
				pawnPushDirection)&NonPromotionFull&^(totalBitBoard),
				pawnPushDirection) & targetBitBoard
			moveListHelper(doublePush, pawnPushDirection*2, moveList, doublePawnPushFlag)
		}

	case KNIGHT:
		var knightFlag Flag
		switch genType {
		case CAPTURE:
			knightFlag = captureFlag
		case QUIET:
			knightFlag = quietFlag
		}

		for from := PopLSB(&thisBitBoard); from != INVALID_POSITION; from = PopLSB(&thisBitBoard) {
			// Generate all possible knight moves from the current position
			// then AND with the targetBitBoard to see if it is a valid move
			// Finally, add to the moveList with the appropriate flag
			validPositions := knightMoveBoard[from] & targetBitBoard
			for to := PopLSB(&validPositions); to != INVALID_POSITION; to = PopLSB(&validPositions) {
				*moveList = append(*moveList, NewMove(from, to, knightFlag))
			}
		}
	}
}

func enemyPieceAttackBitBoard(board *Board) (retval BitBoard) {
	// Current state of board
	currentState := board.GetTopState()
	var enemyPieces Pieces
	var friendlyKing BitBoard
	// Get the opposing enemyPieces that will generate moves
	if currentState.TurnColor == WHITE {
		enemyPieces = board.B
		friendlyKing = board.W.King
	} else {
		enemyPieces = board.W
		friendlyKing = board.B.King
	}

	//Pawns
	if enemyPieces.Pawn != 0 {
		var pawnPushDirection Direction
		if currentState.TurnColor == WHITE {
			pawnPushDirection = S
		} else {
			pawnPushDirection = N
		}
		retval |= Shift(enemyPieces.Pawn&^Col8Full, pawnPushDirection+E) | Shift(enemyPieces.Pawn&^Col1Full, pawnPushDirection+W)
	}

	//Knights
	if enemyPieces.Knight != 0 {
		for from := PopLSB(&enemyPieces.Knight); from != INVALID_POSITION; from = PopLSB(&enemyPieces.Knight) {
			retval |= knightMoveBoard[from]
		}
	}

	// Sliding enemyPieces variables, total occupancy - our king to prevent moving along a check ray
	modifiedTotal := totalBitBoard & ^friendlyKing

	// Rook
	if enemyPieces.Rook != 0 {
		for from := PopLSB(&enemyPieces.Rook); from != INVALID_POSITION; from = PopLSB(&enemyPieces.Rook) {
			retval |= GetRookMoves(from,
				RookMask(from)&modifiedTotal)
		}
	}

	// Bishop
	if enemyPieces.Bishop != 0 {
		for from := PopLSB(&enemyPieces.Bishop); from != INVALID_POSITION; from = PopLSB(&enemyPieces.Bishop) {
			retval |= GetBishopMoves(from,
				BishopMask(from)&modifiedTotal)
		}
	}

	// Queen
	if enemyPieces.Queen != 0 {
		for from := PopLSB(&enemyPieces.Queen); from != INVALID_POSITION; from = PopLSB(&enemyPieces.Queen) {
			rookMoves := GetRookMoves(from,
				RookMask(from)&modifiedTotal)
			bishopMoves := GetBishopMoves(from,
				BishopMask(from)&modifiedTotal)
			retval |= (rookMoves | bishopMoves)
		}
	}

	// King
	retval |= kingMoveBoard[PopLSB(&enemyPieces.King)]

	return retval
}

func generateKing(board *Board, targetBitBoard BitBoard, genType int, inCheck bool, moveList *[]Move) {
	// Current state of board, includes who's turn it is, any EnPassant possibility, along with Castling Rights
	currentState := board.GetTopState()

	var pieces Pieces
	var from, to Position
	var castleKing, castleQueen bool
	// Get the pieces that will generate moves
	if currentState.TurnColor == WHITE {
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

	validPositions := kingMoveBoard[from] & targetBitBoard

	var kingFlag Flag
	switch genType {
	case CAPTURE:
		kingFlag = captureFlag
	case QUIET:
		kingFlag = quietFlag

		// Check king can move at least 2 to the right (emptiness included implicitly)
		if !inCheck &&
			castleKing &&
			(getIntermediaryRay(from, from+3)&targetBitBoard) == getIntermediaryRay(from, from+3) {
			*moveList = append(*moveList, NewMove(from, from+2, kingCastleFlag))
		}
		// Check queen side castle empty + king can move at least 2 to the left
		if !inCheck &&
			castleQueen &&
			(getIntermediaryRay(from, from-4)&(totalBitBoard)) == 0 &&
			(getIntermediaryRay(from, from-3)&targetBitBoard) == getIntermediaryRay(from, from-3) {
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
	if currentState.TurnColor == WHITE {
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
		if currentState.TurnColor == WHITE {
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
			placeOnBitBoard(&pinnedPiecesBitBoard, to)
		}
	}

	// Knight
	if enemyPieces.Knight != 0 {
		validCaptures := knightMoveBoard[from] & enemyPieces.Knight // All possible knight positions
		for to = PopLSB(&validCaptures); to != INVALID_POSITION; to = PopLSB(&validCaptures) {
			*checkingPieces = append(*checkingPieces,
				CheckerInfo{
					*board.PieceInfoArr[to], to, 0})
			placeOnBitBoard(&pinnedPiecesBitBoard, to)
		}
	}

	// Rook/Queen
	if enemyPieces.Rook != 0 || enemyPieces.Queen != 0 {
		possibleCheckers := GetRookMoves(from, RookMask(from)&enemyBitBoard)
		possibleCheckers &= enemyPieces.Queen | enemyPieces.Rook

		for to = PopLSB(&possibleCheckers); to != INVALID_POSITION; to = PopLSB(&possibleCheckers) {
			checkRay := getIntermediaryRay(from, to)
			checkerPieceInfo := CheckerInfo{*board.PieceInfoArr[to], to, checkRay}

			checkRayPinned := checkRay & friendBitboard
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
				placeOnBitBoard(&pinnedPiecesBitBoard, friendlyPinnedPosition)
			}
		}
	}

	// Bishop/Queen
	if enemyPieces.Bishop != 0 || enemyPieces.Queen != 0 {
		possibleCheckers := GetBishopMoves(from, BishopMask(from)&enemyBitBoard)
		possibleCheckers &= enemyPieces.Queen | enemyPieces.Bishop

		for to = PopLSB(&possibleCheckers); to != INVALID_POSITION; to = PopLSB(&possibleCheckers) {
			checkRay := getIntermediaryRay(from, to)
			checkerPieceInfo := CheckerInfo{*board.PieceInfoArr[to], to, checkRay}

			checkRayPinned := checkRay & friendBitboard
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
				placeOnBitBoard(&pinnedPiecesBitBoard, friendlyPinnedPosition)
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
			generateSliding(thisBitBoard, targetBitBoard, ROOK, genType, moveList)
		case BISHOP:
			generateSliding(thisBitBoard, targetBitBoard, BISHOP, genType, moveList)
		case QUEEN:
			generateSliding(thisBitBoard, targetBitBoard, QUEEN, genType, moveList)
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

	if currentState.TurnColor == WHITE {
		enemyPieces = &board.B
		friendlyPieces = &board.W
	} else {
		enemyPieces = &board.W
		friendlyPieces = &board.B
	}

	if friendlyPieces.King == 0 || enemyPieces.King == 0 {
		return moveList
	}

	friendBitboard = friendlyPieces.OccupancyBitBoard()
	enemyBitBoard = enemyPieces.OccupancyBitBoard()
	totalBitBoard = friendBitboard | enemyBitBoard

	pinnedPieces, pinnedPiecesBitBoard, checkingPieces := generateCheck(board)
	inCheck := len(*checkingPieces) > 0

	switch len(*checkingPieces) {
	case 0:
		// No checkers
		switch genType {
		case CAPTURE:
			targetBitBoard = enemyBitBoard
		case QUIET:
			targetBitBoard = ^totalBitBoard
		}
		generatePinned(board, genType, pinnedPieces, &moveList)
		generateSliding(friendlyPieces.Queen&^pinnedPiecesBitBoard, targetBitBoard, QUEEN, genType, &moveList)
		generateSliding(friendlyPieces.Bishop&^pinnedPiecesBitBoard, targetBitBoard, BISHOP, genType, &moveList)
		generateSliding(friendlyPieces.Rook&^pinnedPiecesBitBoard, targetBitBoard, ROOK, genType, &moveList)
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
		generateSliding(friendlyPieces.Queen&^pinnedPiecesBitBoard, targetBitBoard, QUEEN, genType, &moveList)
		generateSliding(friendlyPieces.Bishop&^pinnedPiecesBitBoard, targetBitBoard, BISHOP, genType, &moveList)
		generateSliding(friendlyPieces.Rook&^pinnedPiecesBitBoard, targetBitBoard, ROOK, genType, &moveList)
		generateNonSliding(board, friendlyPieces.Knight&^pinnedPiecesBitBoard, targetBitBoard, KNIGHT, genType, &moveList)
		generateNonSliding(board, friendlyPieces.Pawn&^pinnedPiecesBitBoard, targetBitBoard, PAWN, genType, &moveList)

	}
	switch genType {
	case CAPTURE:
		generateKing(board,
			enemyBitBoard & ^enemyPieceAttackBitBoard(board),
			CAPTURE,
			inCheck,
			&moveList)
	case QUIET:
		generateKing(board,
			^(totalBitBoard) & ^enemyPieceAttackBitBoard(board),
			QUIET,
			inCheck,
			&moveList)
	}

	return moveList
}

// Helper function to generate moves for a bitboard of pieces given a static direction
func moveListHelper(bitboard BitBoard, moveDir Direction, moveList *[]Move, flags ...Flag) {
	for to := PopLSB(&bitboard); to != INVALID_POSITION; to = PopLSB(&bitboard) {
		for _, flag := range flags {
			*moveList = append(*moveList, NewMove(to-Position(moveDir), to, flag))
		}
	}
}

// Checks wether or not a given square is attacked by a given color
func (board *Board) isAttacked(position Position, color int8) bool {
	// Current state of board, includes who's turn it is, any EnPassant possibility, along with Castling Rights
	var enemyPieces *Pieces

	targetBitboard := BitBoard(1) << position
	totalBitBoard := board.W.OccupancyBitBoard() | board.B.OccupancyBitBoard()
	// Get the pieces that will generate moves
	if color == WHITE {
		enemyPieces = &board.W
	} else if color == BLACK {
		enemyPieces = &board.B
	} else {
		panic("Wrong color value given in isAttacked()")
	}

	// Idea is to treat king as every possible non-king enemy piece,
	// then seeing if a reversal of the capture is possible -> "captured" piece has the king in check
	var attackers BitBoard

	// Pawn
	if enemyPieces.Pawn != 0 {
		var pawnPushDirection Direction
		if color == WHITE {
			pawnPushDirection = S
		} else if color == BLACK {
			pawnPushDirection = N
		}
		// Check for possible pawn captures
		// inverting by side columns to prevent going off the board
		attackers = (Shift(targetBitboard, pawnPushDirection+E) & enemyPieces.Pawn &^ Col1Full) |
			(Shift(targetBitboard, pawnPushDirection+W) & enemyPieces.Pawn &^ Col8Full)
		if attackers != 0 {
			return true
		}
	}

	// Knight
	if enemyPieces.Knight != 0 {
		attackers := knightMoveBoard[position] & enemyPieces.Knight // All possible knight positions
		if attackers > 0 {
			return true
		}
	}

	// Rook/Queen
	if enemyPieces.Rook != 0 || enemyPieces.Queen != 0 {
		attackers := GetRookMoves(position, RookMask(position)&totalBitBoard)
		attackers &= enemyPieces.Queen | enemyPieces.Rook

		if attackers > 0 {
			return true
		}
	}

	// Bishop/Queen
	if enemyPieces.Bishop != 0 || enemyPieces.Queen != 0 {
		attackers := GetBishopMoves(position, BishopMask(position)&totalBitBoard)
		attackers &= enemyPieces.Queen | enemyPieces.Bishop

		if attackers > 0 {
			return true
		}
	}

	// King
	{
		attackers := kingMoveBoard[position] & enemyPieces.King
		if attackers > 0 {
			return true
		}
	}

	return false
}
