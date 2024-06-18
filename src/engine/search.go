package chessengine

import (
	"cmp"
	"fmt"
	"slices"
	"time"
)

const (
	MATE_SCORE          = -65535
	MIN_VALUE           = -2147483648
	MAX_VALUE           = 2147483648
	MAX_EXTENSION_DEPTH = 8
	MAX_DEPTH           = 128
)

type searchInfo struct {
	nodeCount uint64
	startTime time.Time
	depth     byte
	score     int
	seldepth  byte
	multipv   byte
}

var latestSearchInfo searchInfo

var bestMoveThisIteration, bestMove Move
var bestEvalThisIteration int

var DebugMode = false
var PV [MAX_DEPTH]uint16

func (board *Board) StartSearch(startTime time.Time, cancelChannel chan int) Move {
	var depth byte = 1
	bestMove = NULL_MOVE
	DebugCollisions = 0
	DebugNewEntries = 0
	PV = [MAX_DEPTH]uint16{}
	latestSearchInfo = searchInfo{startTime: startTime, multipv: 1}

	for depth < MAX_DEPTH {
		bestMoveThisIteration = NULL_MOVE
		bestEvalThisIteration = MIN_VALUE

		latestSearchInfo.depth = depth
		latestSearchInfo.nodeCount = 0
		latestSearchInfo.seldepth = 0

		board.search(depth, 0, MIN_VALUE, MAX_VALUE, 0, cancelChannel)

		if bestMoveThisIteration.enc != NULL_MOVE.enc {
			PV[0] = bestMoveThisIteration.enc
			board.updatePV(1, depth)
			depth++
			bestMove = bestMoveThisIteration
			latestSearchInfo.score = bestEvalThisIteration
			fmt.Println(engineInfoString())
		}

		select {
		case <-cancelChannel:
			return bestMove
		default:
		}
	}

	return bestMove
}

// search performs an alpha-beta pruning of minimax search on the chess board up to the specified depth.
// It returns the best evaluation score found for the current position.
// The function uses alpha-beta pruning to reduce the number of positions that need to be evaluated.
// The depth parameter specifies the maximum depth to search.
// The plyFromRoot parameter specifies the current ply (half-move) count from the root position.
// The alpha and beta parameters define the current alpha-beta window.
// The numExtensions parameter specifies the number of extensions to apply during the search.
// The cancelChannel parameter is used to cancel the search if needed.
// If the search is cancelled, the function returns 0.
func (board *Board) search(depth, plyFromRoot byte, alpha, beta int, numExtensions byte, cancelChannel chan int) int {
	// Check if the search has been cancelled
	select {
	case <-cancelChannel:
		return 0
	default:
	}

	latestSearchInfo.nodeCount++

	if plyFromRoot > 0 {
		// Fifty move rule
		if board.GetTopState().HalfMoveClock >= 100 {
			return 0
		}

		// Threefold repetition
		if board.RepetitionPositionHistory[board.GetTopState().ZobristKey] == 3 {
			return 0
		}
	}

	if ttScore := probeHash(depth, alpha, beta, board.GetTopState().ZobristKey); ttScore != MIN_VALUE {
		if plyFromRoot == 0 {
			bestMoveThisIteration = GetEntry(board.GetTopState().ZobristKey).best
			bestEvalThisIteration = ttScore
		}
		return ttScore
	}

	if depth == 0 {
		eval := board.quiescenceSearch(alpha, beta, plyFromRoot, cancelChannel)
		return eval
	}

	moveList := make([]Move, 0, MAX_MOVE_COUNT)
	moveList = board.GenerateMoves(ALL, moveList)
	board.moveordering(false, plyFromRoot, moveList)

	if len(moveList) == 0 {
		if board.InCheck() {
			return MATE_SCORE + int(plyFromRoot) // Checkmate
		} else {
			return 0 // Stalemate
		}
	}

	hashF := hashfALPHA
	bestMoveThisPath := NULL_MOVE

	for _, move := range moveList {
		board.MakeMove(move)
		extension := extendSearch(board, move, numExtensions)
		score := -board.search(depth-1+extension, plyFromRoot+1, -beta, -alpha, numExtensions+extension, cancelChannel)
		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			return 0
		default:
		}

		if score >= beta {
			// If the score is greater than or equal to beta,
			// it means that the opponent has a better move to choose.
			// We record this information in the transposition table.
			recordHash(depth, beta, hashfBETA, move, board.GetTopState().ZobristKey)
			return beta
		}
		if score > alpha { // This move is better than the current best move
			bestMoveThisPath = move
			hashF = hashfEXACT
			alpha = score
			if plyFromRoot == 0 {
				bestMoveThisIteration = move  // Update the best move for this iteration
				bestEvalThisIteration = score // Update the best evaluation score for this iteration
			}
		}
	}

	recordHash(depth, alpha, hashF, bestMoveThisPath, board.GetTopState().ZobristKey) // Record the best move for this position
	return alpha
}

func (board *Board) quiescenceSearch(alpha, beta int, plyFromRoot byte, cancelChannel chan int) int {
	select { // Check if the search has been cancelled
	case <-cancelChannel:
		return 0
	default:
	}

	latestSearchInfo.nodeCount++

	eval := board.Evaluate()
	if eval >= beta {
		return beta
	}
	if eval > alpha {
		alpha = eval
		latestSearchInfo.seldepth = max(plyFromRoot, latestSearchInfo.seldepth)
	}

	captureMoveList := make([]Move, 0, MAX_CAPTURE_COUNT)
	captureMoveList = board.GenerateMoves(CAPTURE, captureMoveList)
	board.moveordering(true, 0, captureMoveList)

	for _, move := range captureMoveList {
		board.MakeMove(move)
		eval := -board.quiescenceSearch(-beta, -alpha, plyFromRoot+1, cancelChannel)
		board.UnMakeMove()

		if eval >= beta {
			return beta
		}
		if eval > alpha {
			alpha = eval
		}
	}

	return alpha
}

/*
Extends the search by +1 based upon whether current move:
 1. Leaves the baord in check
 2. Puts a pawn in position to be promoted
*/
func extendSearch(board *Board, move Move, numExtensions byte) byte {
	if numExtensions >= MAX_EXTENSION_DEPTH {
		return 0
	}

	if board.InCheck() { // Checks are very interesting to extend out
		return 1
	}

	pieceType := board.PieceInfoArr[getTargetPosition(move)].pieceTYPE
	moveRank := (getTargetPosition(move) & 0b111000) >> 3

	if pieceType == PAWN && (moveRank == 1 || moveRank == 6) { // Pawn on verge of promotion, also interesting
		return 1
	}

	return 0
}

/*
Source: https://lup.lub.lu.se/luur/download?func=downloadFile&recordOId=9069249&fileOId=9069251

MVV-LVA (MostValuable Victim – Least Valuable Aggressor). In chess, moves, where an
opponent’s piece of high value is taken by a piece of low value of one’s own, are
often good moves.

The MVV-LVA heuristic uses this reasoning by applying a score
to each move based on the difference in value between a taking piece and the
piece that is taken.

The higher the value of the taken piece, and the lower value of
the taking piece, the higher the priority score.

For example, a very promising move, when available, is a move where one player may take
the opponent’s queen with their own pawn and this move is thus given the highest priority
by the MVV-LVA heuristic.
*/
func (board *Board) moveordering(inQSearch bool, plyFromRoot byte, moveList []Move) {
	for i := range moveList {
		if !inQSearch && moveList[i].enc == PV[plyFromRoot] {
			moveList[i].priority = 127
			continue
		}
		var takenPiece int

		switch GetFlag(moveList[i]) {
		case captureFlag:
			takenPiece = board.PieceInfoArr[getTargetPosition(moveList[i])].pieceTYPE
		case epCaptureFlag:
			takenPiece = PAWN
		case knightPromotionFlag, bishopPromotionFlag, rookPromotionFlag, queenPromotionFlag, knightPromoCaptureFlag, bishopPromoCaptureFlag, rookPromoCaptureFlag, queenPromoCaptureFlag:
			moveList[i].priority = 100
			continue
		default:
			continue
		}

		switch board.PieceInfoArr[getStartingPosition(moveList[i])].pieceTYPE { // Aggressor
		case PAWN:
			moveList[i].priority += -1
		case KNIGHT:
			moveList[i].priority += -3
		case BISHOP:
			moveList[i].priority += -3
		case ROOK:
			moveList[i].priority += -5
		case QUEEN:
			moveList[i].priority += -9
		case KING:
			moveList[i].priority += -10
		}

		switch takenPiece { // Victim
		case PAWN:
			moveList[i].priority += 10
		case KNIGHT:
			moveList[i].priority += 30
		case BISHOP:
			moveList[i].priority += 30
		case ROOK:
			moveList[i].priority += 50
		case QUEEN:
			moveList[i].priority += 90
		}
	}

	slices.SortFunc(moveList, func(a, b Move) int {
		return cmp.Compare(b.priority, a.priority)
	})
}

// updatePV updates the principal variation (PV) based on the current board state.
// It takes two parameters: plyFromRoot and depth.
// - plyFromRoot represents the number of plies (half-moves) from the root of the search tree.
// - depth represents the maximum depth of the search.
// The function updates the PV by making moves on the board until the specified depth is reached.
// It uses the transposition table (ttMove) to determine the best move at each ply.
// If the best move is a null move or an invalid move, the function stops updating the PV.
// Finally, it restores the board to its original state by undoing the moves made during the update.
func (board *Board) updatePV(plyFromRoot, depth byte) {
	for i := byte(0); i < plyFromRoot; i++ {
		board.MakeMove(Move{enc: PV[i], priority: 0})
	}
	for ; plyFromRoot < depth; plyFromRoot++ {
		ttMove := GetEntry(board.GetTopState().ZobristKey).best
		if ttMove.enc == NULL_MOVE.enc || !board.validMove(ttMove) {
			break
		}
		PV[plyFromRoot] = ttMove.enc
		board.MakeMove(ttMove)
	}
	for ; plyFromRoot > 0; plyFromRoot-- {
		board.UnMakeMove()
	}
}

/*
	Outputs the engine info from a given search depth following the format:

info depth <depth> seldepth <maxdepth searched> multipv <principal variations> score cp <score> nodes <nodecount> nps <nodes / time> time <time taken in ms> pv <pv>
*/
func engineInfoString() string {
	elapsedTime := time.Since(latestSearchInfo.startTime)
	nps := int64(float64(latestSearchInfo.nodeCount) / elapsedTime.Seconds())
	// Convert PV chain up to depth to a single string seperated by " "
	pvString := ""
	for i := byte(0); i < latestSearchInfo.depth; i++ {
		if PV[i] == NULL_MOVE.enc {
			break
		}
		pvString += MoveToString(Move{enc: PV[i], priority: 0}) + " "
	}
	// Strip surrounding whitespace from PV string
	pvString = pvString[:len(pvString)-1]

	return fmt.Sprintf("info depth %d seldepth %d multipv %d score cp %d nodes %d nps %d time %d pv %s",
		latestSearchInfo.depth, latestSearchInfo.seldepth, latestSearchInfo.multipv,
		latestSearchInfo.score, latestSearchInfo.nodeCount, nps, elapsedTime.Milliseconds(), pvString)

}
