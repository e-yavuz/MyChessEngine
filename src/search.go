package chessengine

import (
	"sort"
)

const (
	MATE_SCORE = -65535
	INT_MIN    = -2147483648
	INT_MAX    = 2147483648
)

func (board *Board) Search(depth, plyFromRoot, alpha, beta int, bestMove Move, cancelChannel chan int) (Move, int) {
	if depth == 0 {
		return NULL_MOVE, board.QuiescenceSearch(alpha, beta, cancelChannel)
	}

	// Check if the search has been cancelled
	select {
	case <-cancelChannel:
		return NULL_MOVE, 0
	default:
	}

	totalMoveList := append(*board.generateMoves(CAPTURE), *board.generateMoves(QUIET)...)
	sort.Slice(totalMoveList, func(i, j int) bool {
		return totalMoveList[i] > totalMoveList[j]
	})

	// if best_move != NULL_MOVE, add it to the front of the list
	if bestMove != NULL_MOVE {
		totalMoveList = append([]Move{bestMove}, totalMoveList...)
	}

	if len(totalMoveList) == 0 {
		if board.inCheck() {
			return NULL_MOVE, MATE_SCORE - depth
		} else {
			return NULL_MOVE, 0
		}
	}

	for _, move := range totalMoveList {
		board.MakeMove(move)

		board.Search(depth-1, plyFromRoot+1, -beta, -alpha, NULL_MOVE, cancelChannel)
		_, score := board.Search(depth-1, plyFromRoot+1, -beta, -alpha, NULL_MOVE, cancelChannel)
		score = -score

		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			if alpha == INT_MIN {
				return NULL_MOVE, 0
			} else {
				return bestMove, alpha
			}
		default:
		}

		if score >= beta {
			return NULL_MOVE, beta
		}
		if score > alpha {
			if plyFromRoot == 0 {
				bestMove = move
			}
			alpha = score
		}
	}

	return bestMove, alpha
}

func (board *Board) QuiescenceSearch(alpha, beta int, cancelChannel chan int) int {
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

	captureMoveList := board.generateMoves(CAPTURE)

	for _, move := range *captureMoveList {
		board.MakeMove(move)
		score := -board.QuiescenceSearch(-beta, -alpha, cancelChannel)
		board.UnMakeMove()

		select { // Check if the search has been cancelled
		case <-cancelChannel:
			return 0
		default:
		}

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}
