package chessengine

import "fmt"

var bestMoveChain []Move

func (board *Board) StartSearchDebug(cancelChannel chan int) (Move, int16, []Move) {
	var depth byte = 1
	bestMove = NULL_MOVE
	DebugCollisions = 0
	DebugNewEntries = 0

	for {
		bestMoveThisIteration = NULL_MOVE
		bestEvalThisIteration = MIN_VALUE
		_, moveChain := board.searchDebug(depth, 0, MIN_VALUE, MAX_VALUE, 0, cancelChannel)

		if bestMoveThisIteration != NULL_MOVE {
			depth++
			bestMove = bestMoveThisIteration
			bestEval = bestEvalThisIteration
			fmt.Print("[ ")
			for _, move := range moveChain {
				fmt.Print(MoveToString(move) + " ")
			}
			fmt.Print("]")
			fmt.Println()
		}

		select {
		case <-cancelChannel:
			return bestMove, bestEval, bestMoveChain
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
func (board *Board) searchDebug(depth, plyFromRoot byte, alpha, beta int16, numExtensions byte, cancelChannel chan int) (int16, []Move) {
	// Check if the search has been cancelled
	select {
	case <-cancelChannel:
		return 0, []Move{}
	default:
	}

	if plyFromRoot > 0 {
		// Fifty move rule
		if board.GetTopState().HalfMoveClock >= 100 {
			return 0, []Move{}
		}

		// Threefold repetition
		if board.RepetitionPositionHistory[board.GetTopState().ZobristKey] == 3 {
			return 0, []Move{}
		}
	}

	if ttScore := probeHash(depth, alpha, beta, board.GetTopState().ZobristKey); ttScore != MIN_VALUE {
		if plyFromRoot == 0 {
			bestMoveThisIteration = GetEntry(board.GetTopState().ZobristKey).best
			bestEvalThisIteration = ttScore
		}
		return ttScore, []Move{GetEntry(board.GetTopState().ZobristKey).best}
	}

	if depth == 0 {
		eval := board.quiescenceSearch(alpha, beta, plyFromRoot+1, cancelChannel)
		return eval, []Move{}
	}

	moveList := make([]Move, 0, MAX_MOVE_COUNT+1)
	moveList = getAllMoves(board, plyFromRoot, moveList)

	if len(moveList) == 0 {
		if board.InCheck() {
			return MATE_SCORE + int16(plyFromRoot), []Move{} // Checkmate
		} else {
			return 0, []Move{} // Stalemate
		}
	}

	hashF := hashfALPHA
	bestMoveThisPath := NULL_MOVE
	moveChain := make([]Move, 0, depth)

	for _, move := range moveList {
		board.MakeMove(move)
		extension := extendSearch(board, move, numExtensions)
		score, newMoveChain := board.searchDebug(depth-1+extension, plyFromRoot+1, -beta, -alpha, numExtensions+extension, cancelChannel)
		score *= -1
		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			return 0, []Move{}
		default:
		}

		if score >= beta {
			// If the score is greater than or equal to beta,
			// it means that the opponent has a better move to choose.
			// We record this information in the transposition table.
			recordHash(depth, beta, hashfBETA, move, board.GetTopState().ZobristKey)
			return beta, []Move{}
		}
		if score > alpha { // This move is better than the current best move
			bestMoveThisPath = move
			moveChain = newMoveChain
			hashF = hashfEXACT
			alpha = score
			if plyFromRoot == 0 {
				bestMoveThisIteration = move  // Update the best move for this iteration
				bestEvalThisIteration = score // Update the best evaluation score for this iteration
			}
		}
	}

	recordHash(depth, alpha, hashF, bestMoveThisPath, board.GetTopState().ZobristKey) // Record the best move for this position
	moveChain = append([]Move{bestMoveThisPath}, moveChain...)
	return alpha, moveChain
}
