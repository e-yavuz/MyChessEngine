package chessengine

import (
	"sort"
)

const (
	MATE_SCORE          = -65535
	MIN_VALUE           = -2147483648
	MAX_VALUE           = 2147483648
	MAX_EXTENSION_DEPTH = 3
)

var bestMoveThisIteration, bestMove Move
var bestEvalThisIteration, bestEval int

func (board *Board) StartSearch(cancelChannel chan int) (Move, int, byte) {
	var depth byte = 1
	bestMove = NULL_MOVE

	for {
		bestMoveThisIteration = NULL_MOVE
		bestEvalThisIteration = MIN_VALUE
		board.search(depth, 0, MIN_VALUE, MAX_VALUE, 0, cancelChannel)

		if bestMoveThisIteration != NULL_MOVE {
			depth++
			bestMove = bestMoveThisIteration
			bestEval = bestEvalThisIteration
		}

		select {
		case <-cancelChannel:
			return bestMove, bestEval, depth
		default:
		}
	}
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

	if ttScore := probeHash(depth, alpha, beta, board.GetTopState().ZobristKey); ttScore != MIN_VALUE {
		if plyFromRoot == 0 {
			bestMoveThisIteration = GetEntry(board.GetTopState().ZobristKey).best
			bestEvalThisIteration = ttScore
		}
		return ttScore
	}

	if depth == 0 {
		eval := board.quiescenceSearch(alpha, beta, cancelChannel)
		// recordHash(0, eval, hashfEXACT, NULL_MOVE, board.GetTopState().ZobristKey)
		return eval
	}

	totalMoveList := *getAllMoves(board, plyFromRoot)

	if len(totalMoveList) == 0 {
		if board.InCheck() {
			return MATE_SCORE + int(plyFromRoot) // Checkmate
		} else {
			return 0 // Stalemate
		}
	}

	hashF := hashfALPHA
	bestMoveThisPath := NULL_MOVE

	for _, move := range totalMoveList {
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
			// If the score is greater than or equal to beta, it means that the opponent has a better move to choose.
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

func (board *Board) quiescenceSearch(alpha, beta int, cancelChannel chan int) int {
	select { // Check if the search has been cancelled
	case <-cancelChannel:
		return 0
	default:
	}

	eval := board.Evaluate()
	if eval >= beta {
		return beta
	}
	if eval > alpha {
		alpha = eval
	}

	captureMoveList := *board.GenerateMoves(CAPTURE)
	sort.Slice(captureMoveList, func(i, j int) bool {
		return captureMoveList[i] > captureMoveList[j]
	})

	for _, move := range captureMoveList {
		board.MakeMove(move)
		eval := -board.quiescenceSearch(-beta, -alpha, cancelChannel)
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
Returns a list of all possible moves for the current board state.
The list is sorted such that the best move is first.
*/
func getAllMoves(board *Board, plyFromRoot byte) *[]Move {
	var totalMoveList []Move
	captureList := *board.GenerateMoves(CAPTURE)
	quietList := *board.GenerateMoves(QUIET)

	if bestMove != NULL_MOVE && plyFromRoot == 0 {
		totalMoveList = make([]Move, 0, len(captureList)+len(quietList)+1)
		totalMoveList = append(totalMoveList, bestMove)
		totalMoveList = append(totalMoveList, captureList...)
		totalMoveList = append(totalMoveList, quietList...)
		sort.Slice(totalMoveList[1:], func(i, j int) bool {
			return totalMoveList[i] > totalMoveList[j]
		})
	} else {
		totalMoveList = make([]Move, 0, len(captureList)+len(quietList))
		totalMoveList = append(totalMoveList, captureList...)
		totalMoveList = append(totalMoveList, quietList...)
		sort.Slice(totalMoveList, func(i, j int) bool {
			return totalMoveList[i] > totalMoveList[j]
		})
	}

	return &totalMoveList
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

	pieceType := board.PieceInfoArr[GetTargetPosition(move)].pieceTYPE
	moveRank := (GetTargetPosition(move) & 0b111000) >> 3

	if pieceType == PAWN && (moveRank == 1 || moveRank == 6) { // Pawn on verge of promotion, also interesting
		return 1
	}

	return 0
}
