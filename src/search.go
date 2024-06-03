package chessengine

import (
	"sort"
)

const (
	MATE_SCORE = -65535
	MIN_VALUE  = -2147483648
	MAX_VALUE  = 2147483648
)

func (board *Board) Search(depth, plyFromRoot byte, alpha, beta int, bestMove Move, cancelChannel chan int) (Move, int) {
	hashF := hashfALPHA

	// Check if the search has been cancelled
	select {
	case <-cancelChannel:
		return NULL_MOVE, 0
	default:
	}

	if ttMove, ttScore := ProbeHash(depth, alpha, beta, board.GetTopState().ZobristKey); ttScore != MIN_VALUE {
		return ttMove, ttScore
	}

	// Check if the search reached the maximum depth
	if depth == 0 {
		eval := board.QuiescenceSearch(alpha, beta, cancelChannel)
		RecordHash(0, eval, hashfEXACT, NULL_MOVE, board.GetTopState().ZobristKey)
		return NULL_MOVE, eval
	}

	// Sort the capture list in descending order, hopefully improving alpha-beta pruning
	captureMoveList := *board.generateMoves(CAPTURE)
	sort.Slice(captureMoveList, func(i, j int) bool {
		return captureMoveList[i] > captureMoveList[j]
	})

	// Generate the total move list, don't sort Quiet moves because a lot of them are just positioning?
	totalMoveList := append(captureMoveList, *board.generateMoves(QUIET)...)

	// if best_move != NULL_MOVE, add it to the front of the list
	if bestMove != NULL_MOVE {
		totalMoveList = append([]Move{bestMove}, totalMoveList...)
	}

	if len(totalMoveList) == 0 {
		if board.inCheck() {
			return NULL_MOVE, MATE_SCORE + int(plyFromRoot)
		} else {
			return NULL_MOVE, 0
		}
	}

	for _, move := range totalMoveList {
		board.MakeMove(move)

		_, score := board.Search(depth-1, plyFromRoot+1, -beta, -alpha, NULL_MOVE, cancelChannel)
		score = -score

		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			if alpha == MIN_VALUE {
				return NULL_MOVE, 0
			} else {
				return bestMove, alpha
			}
		default:
		}

		if score >= beta {
			RecordHash(depth, beta, hashfBETA, move, board.GetTopState().ZobristKey)
			return NULL_MOVE, beta
		}
		if score > alpha {
			bestMove = move
			hashF = hashfEXACT
			alpha = score
		}
	}

	RecordHash(depth, alpha, hashF, bestMove, board.GetTopState().ZobristKey)
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

	captureMoveList := *board.generateMoves(CAPTURE)
	sort.Slice(captureMoveList, func(i, j int) bool {
		return captureMoveList[i] > captureMoveList[j]
	})

	for _, move := range captureMoveList {
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
