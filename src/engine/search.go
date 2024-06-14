package chessengine

import (
	"sort"
)

const (
	MATE_SCORE          int16 = MIN_VALUE + 1
	MIN_VALUE           int16 = -32767
	MAX_VALUE           int16 = 32767
	MAX_EXTENSION_DEPTH       = 8
	MAX_DEPTH                 = 64
)

const (
	killerMoveBias = 100
)

var bestMoveThisIteration, bestMove Move
var bestEvalThisIteration, bestEval int16

var killerMoves [MAX_DEPTH]map[Move]bool

func init() {
	for i := range killerMoves {
		killerMoves[i] = make(map[Move]bool, MAX_MOVE_COUNT)
	}
}

func (board *Board) StartSearch(cancelChannel chan int) Move {
	var depth byte = 1
	bestMove = NULL_MOVE
	for i := range killerMoves {
		for k := range killerMoves[i] {
			delete(killerMoves[i], k)
		}
	}

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
			return bestMove
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
func (board *Board) search(depth, plyFromRoot byte, alpha, beta int16, numExtensions byte, cancelChannel chan int) int16 {
	// Check if the search has been cancelled
	select {
	case <-cancelChannel:
		return 0
	default:
	}

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
		eval := board.quiescenceSearch(alpha, beta, plyFromRoot+1, cancelChannel)
		return eval
	}

	moveList := make([]Move, 0, MAX_MOVE_COUNT+1)
	moveList = getAllMoves(board, plyFromRoot, moveList)

	if len(moveList) == 0 {
		if board.InCheck() {
			return MATE_SCORE + int16(plyFromRoot) // Checkmate
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
			killerMoves[plyFromRoot][move] = true
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

func (board *Board) quiescenceSearch(alpha, beta int16, plyFromRoot byte, cancelChannel chan int) int16 {
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

	captureMoveList := make([]Move, 0, MAX_CAPTURE_COUNT)
	captureMoveList = board.GenerateMoves(CAPTURE, captureMoveList)
	board.moveordering(captureMoveList, plyFromRoot)
	sort.Slice(captureMoveList, func(i, j int) bool {
		return compareMove(captureMoveList[i], captureMoveList[j])
	})

	for _, move := range captureMoveList {
		board.MakeMove(move)
		eval := -board.quiescenceSearch(-beta, -alpha, plyFromRoot+1, cancelChannel)
		board.UnMakeMove()

		if eval >= beta {
			killerMoves[plyFromRoot][move] = true
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
func getAllMoves(board *Board, plyFromRoot byte, moveList []Move) []Move {
	if bestMove != NULL_MOVE && plyFromRoot == 0 {
		moveList = append(moveList, bestMove)
	}
	moveList = board.GenerateMoves(ALL, moveList)

	if bestMove != NULL_MOVE && plyFromRoot == 0 {
		sort.Slice(moveList[1:], func(i, j int) bool {
			return compareMove(moveList[i], moveList[j])
		})
	} else {
		sort.Slice(moveList, func(i, j int) bool {
			return compareMove(moveList[i], moveList[j])
		})
	}

	return moveList
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
func (board *Board) moveordering(moveList []Move, plyFromRoot byte) {
	for i := range moveList {
		ok, val := killerMoves[plyFromRoot][moveList[i]]
		if ok && val {
			moveList[i].priority = killerMoveBias
			continue
		}
		movingPiece := board.PieceInfoArr[GetStartingPosition(moveList[i])].pieceTYPE
		moveFlag := GetFlag(moveList[i])
		var takenPiece int

		if moveFlag&^captureFlag == 0 {
			takenPiece = board.PieceInfoArr[GetTargetPosition(moveList[i])].pieceTYPE
		} else if moveFlag&^epCaptureFlag == 0 {
			takenPiece = PAWN
		} else {
			continue
		}

		switch movingPiece {
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

		switch takenPiece {
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
}
