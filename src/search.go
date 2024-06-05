package chessengine

import (
	"sort"
)

const (
	MATE_SCORE = -65535
	MIN_VALUE  = -2147483648
	MAX_VALUE  = 2147483648
)

var bestMoveThisIteration, bestMove Move
var bestEvalThisIteration, bestEval int

func (board *Board) StartSearch(cancelChannel chan int) (Move, int, byte) {
	var depth byte = 1
	bestMove = NULL_MOVE

	for {
		bestMoveThisIteration = NULL_MOVE
		bestEvalThisIteration = MIN_VALUE
		board.Search(depth, 0, MIN_VALUE, MAX_VALUE, cancelChannel)

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

func (board *Board) Search(depth, plyFromRoot byte, alpha, beta int, cancelChannel chan int) int {
	// Check if the search has been cancelled
	select {
	case <-cancelChannel:
		return 0
	default:
	}

	if ttScore := ProbeHash(depth, alpha, beta, board.GetTopState().ZobristKey); ttScore != MIN_VALUE {
		if plyFromRoot == 0 {
			bestMoveThisIteration = GetEntry(board.GetTopState().ZobristKey).best
			bestEvalThisIteration = ttScore
		}
		return ttScore
	}

	if depth == 0 {
		eval := board.QuiescenceSearch(alpha, beta, cancelChannel)
		// RecordHash(0, eval, hashfEXACT, NULL_MOVE, board.GetTopState().ZobristKey)
		return eval
	}

	totalMoveList := *getAllMoves(board, plyFromRoot)

	if len(totalMoveList) == 0 {
		if board.inCheck() {
			return MATE_SCORE + int(plyFromRoot) // Checkmate
		} else {
			return 0 // Stalemate
		}
	}

	hashF := hashfALPHA
	bestMoveThisPath := NULL_MOVE

	for _, move := range totalMoveList {
		board.MakeMove(move)

		score := -board.Search(depth-1, plyFromRoot+1, -beta, -alpha, cancelChannel)

		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			return 0
		default:
		}

		if score >= beta {
			RecordHash(depth, beta, hashfBETA, move, board.GetTopState().ZobristKey)
			return beta
		}
		if score > alpha {
			bestMoveThisPath = move
			hashF = hashfEXACT
			alpha = score
			if plyFromRoot == 0 {
				bestMoveThisIteration = move
				bestEvalThisIteration = score
			}
		}
	}

	RecordHash(depth, alpha, hashF, bestMoveThisPath, board.GetTopState().ZobristKey)
	return alpha
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
		eval := -board.QuiescenceSearch(-beta, -alpha, cancelChannel)
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

func getAllMoves(board *Board, plyFromRoot byte) *[]Move {
	var totalMoveList []Move
	captureList := *board.generateMoves(CAPTURE)
	quietList := *board.generateMoves(QUIET)

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
