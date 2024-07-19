package chessengine

import (
	"cmp"
	"fmt"
	"slices"
	"time"
)

const (
	MATE_SCORE             = -65535
	MIN_VALUE              = -2147483648
	MAX_VALUE              = 2147483648
	MAX_EXTENSION_DEPTH    = 8
	MAX_QSEARCH_DEPTH      = 30
	MAX_SEARCH_DEPTH       = 63
	DELTAPRUNE_MARGIN      = 200
	LATE_GAME_PHASE_CUTOFF = 4
	triangleTableSize      = ((MAX_SEARCH_DEPTH+MAX_EXTENSION_DEPTH)*(MAX_SEARCH_DEPTH+MAX_EXTENSION_DEPTH) + (MAX_SEARCH_DEPTH + MAX_EXTENSION_DEPTH)) / 2
	squareTableSize        = (MAX_SEARCH_DEPTH + MAX_EXTENSION_DEPTH) * (MAX_SEARCH_DEPTH + MAX_EXTENSION_DEPTH)
)

type searchInfo struct {
	nodeCount uint64
	startTime time.Time
	depth     int8
	score     int
	seldepth  int8
	multipv   byte
	qNodes    uint64
	pvNodes   uint64
	allNodes  uint64
	cutNodes  uint64
}

var latestSearchInfo searchInfo

var bestEvalThisIteration int
var currentSearchTurn byte

var DebugMode = false
var savedPV [MAX_SEARCH_DEPTH]Move

var searchMovePool [MAX_SEARCH_DEPTH][MAX_MOVE_COUNT]Move
var qsearchMovePool [MAX_QSEARCH_DEPTH][MAX_CAPTURE_COUNT]Move

var pv [squareTableSize]Move
var pvPtr int

func (board *Board) StartSearchNoDepth(startTime time.Time, cancelChannel chan struct{}) Move {
	return board.initSearch(startTime, MAX_SEARCH_DEPTH, cancelChannel)
}

func (board *Board) StartSearchDepth(startTime time.Time, max_depth int8, cancelChannel chan struct{}) Move {
	return board.initSearch(startTime, max_depth, cancelChannel)
}

func (board *Board) initSearch(startTime time.Time, max_depth int8, cancelChannel chan struct{}) Move {
	var depth int8 = 1
	pvPtr = 0
	savedPV = [MAX_SEARCH_DEPTH]Move{}
	if DebugMode {
		TTDebugReset(board)
	}

	currentSearchTurn = board.GetTopState().TurnCounter
	bestEvalThisIteration = MIN_VALUE

	for depth <= max_depth {

		latestSearchInfo = searchInfo{startTime: startTime, multipv: 1, depth: depth}

		board.search(depth, 0, MIN_VALUE, MAX_VALUE, 0, cancelChannel)

		if pv[0] != NULL_MOVE {
			copy(savedPV[:depth], pv[:depth])
			depth++
			latestSearchInfo.score = bestEvalThisIteration

			fmt.Println(engineInfoString())
			if DebugMode {
				fmt.Print(TTDebugInfo())
			}
		}

		select {
		case <-cancelChannel:
			return savedPV[0]
		default:
		}
	}

	return savedPV[0]
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
func (board *Board) search(depth, plyFromRoot int8, alpha, beta int, numExtensions int8, cancelChannel chan struct{}) int {
	// Check if the search has been cancelled
	select {
	case <-cancelChannel:
		return bestEvalThisIteration
	default:
	}

	if plyFromRoot > 0 {
		// Fifty move rule
		if board.GetTopState().HalfMoveClock >= 100 {
			latestSearchInfo.nodeCount++
			return 0
		}

		// Threefold repetition
		if board.RepetitionPositionHistory[board.GetTopState().ZobristKey] == 3 {
			latestSearchInfo.nodeCount++
			return 0
		}
	}

	if depth <= 0 {
		eval := board.quiescenceSearch(alpha, beta, plyFromRoot, 0, cancelChannel)
		return eval
	}

	probeScore, probeNodeType, probeMove := probeHash(depth, currentSearchTurn, alpha, beta, board.GetTopState().ZobristKey)
	if probeScore != MIN_VALUE {
		if plyFromRoot == 0 && probeNodeType == PVnode {
			bestEvalThisIteration = probeScore
		}
		return probeScore
	}

	latestSearchInfo.nodeCount++

	moveList := board.GenerateMoves(ALL, searchMovePool[plyFromRoot][:0])
	board.moveordering(false, savedPV[plyFromRoot], probeMove, moveList)

	if len(moveList) == 0 {
		if board.InCheck() {
			return MATE_SCORE + int(plyFromRoot) // Checkmate
		} else {
			return 0 // Stalemate
		}
	}

	nodeType := ALLnode
	this_pvPtr := pvPtr
	pv[pvPtr] = NULL_MOVE // initialize empty PV
	pvPtr += int(MAX_SEARCH_DEPTH + MAX_EXTENSION_DEPTH)

	{
		move := moveList[0]
		// using fail soft with negamax:
		board.MakeMove(move)
		extension := extendSearch(board, move, numExtensions)
		score := -board.search(depth-1+extension, plyFromRoot+1, -beta, -alpha, numExtensions+extension, cancelChannel)
		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			return bestEvalThisIteration
		default:
		}

		if score >= beta {
			// If the score is greater than or equal to beta,
			// it means that the opponent has a better move to choose.
			// We record this information in the transposition table.
			recordHash(depth, CUTnode, currentSearchTurn, beta, NULL_MOVE, board.GetTopState().ZobristKey)
			latestSearchInfo.cutNodes++
			pvPtr = this_pvPtr
			return beta
		}
		if score > alpha { // This move is better than the current best move
			updatePVTable(this_pvPtr, move, depth)
			nodeType = PVnode
			alpha = score
			if plyFromRoot == 0 {
				bestEvalThisIteration = score // Update the best evaluation score for this iteration
			}
		}
	}

	for _, move := range moveList[1:] {
		board.MakeMove(move)

		extension := extendSearch(board, move, numExtensions)
		score := -board.search(depth-1+extension, plyFromRoot+1, -alpha-1, -alpha, numExtensions+extension, cancelChannel)
		if score > alpha {
			score = -board.search(depth-1+extension, plyFromRoot+1, -beta, -alpha, numExtensions+extension, cancelChannel)
		}

		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			return bestEvalThisIteration
		default:
		}

		if score >= beta {
			recordHash(depth, CUTnode, currentSearchTurn, beta, NULL_MOVE, board.GetTopState().ZobristKey)
			latestSearchInfo.cutNodes++
			pvPtr = this_pvPtr
			return beta
		}
		if score > alpha { // This move is better than the current best move
			updatePVTable(this_pvPtr, move, depth)
			nodeType = PVnode
			alpha = score
			if plyFromRoot == 0 {
				bestEvalThisIteration = score // Update the best evaluation score for this iteration
			}
		}
	}

	if nodeType == ALLnode {
		latestSearchInfo.allNodes++
	} else {
		latestSearchInfo.pvNodes++
	}

	pvPtr = this_pvPtr
	recordHash(depth, nodeType, currentSearchTurn, alpha, pv[this_pvPtr], board.GetTopState().ZobristKey) // Record the best move for this position
	return alpha
}

func (board *Board) quiescenceSearch(alpha, beta int, plyFromRoot, plyFromSearch int8, cancelChannel chan struct{}) int {
	select { // Check if the search has been cancelled
	case <-cancelChannel:
		return bestEvalThisIteration
	default:
	}

	latestSearchInfo.qNodes++

	eval, mgPhase, egPhase := board.Evaluate()
	if eval >= beta {
		return beta
	}

	/*
		Delta Pruning
		https://www.chessprogramming.org/Delta_Pruning
			a technique similar in concept to futility pruning,
			only used in the quiescence search.

			It works as follows:
				before we make a capture, we test whether the captured piece value
				plus some safety margin (typically around 200 centipawns) are enough
				to raise alpha for the current node.
	*/
	deltaPrune := alpha - board.EvaluateMaterial(mgPhase, egPhase) - DELTAPRUNE_MARGIN

	if eval > alpha {
		alpha = eval
		latestSearchInfo.seldepth = max(plyFromRoot-latestSearchInfo.depth, latestSearchInfo.seldepth)
	}

	if plyFromSearch >= MAX_QSEARCH_DEPTH {
		return alpha
	}

	captureMoveList := board.GenerateMoves(CAPTURE, qsearchMovePool[plyFromSearch][:0])
	board.moveordering(true, NULL_MOVE, NULL_MOVE, captureMoveList)

	for _, move := range captureMoveList {
		// Delta pruning cut
		if mgPhase > LATE_GAME_PHASE_CUTOFF && getTargetPieceValue(board, move, egPhase) < deltaPrune {
			continue
		}

		board.MakeMove(move)
		eval := -board.quiescenceSearch(-beta, -alpha, plyFromRoot+1, plyFromSearch+1, cancelChannel)
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
func extendSearch(board *Board, move Move, numExtensions int8) int8 {
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
func (board *Board) moveordering(usePV_TT bool, PVMove Move, TTMove Move, moveList []Move) {
	for i := range moveList {
		if !usePV_TT {
			if moveList[i].enc == PVMove.enc {
				moveList[i].priority = 127
				continue
			}
			if moveList[i].enc == TTMove.enc {
				moveList[i].priority = 126
				continue
			}
		}
		var takenPiece int

		switch GetFlag(moveList[i]) {
		case captureFlag:
			takenPiece = board.PieceInfoArr[getTargetPosition(moveList[i])].pieceTYPE
		case epCaptureFlag:
			takenPiece = PAWN
		case knightPromotionFlag, bishopPromotionFlag, rookPromotionFlag, knightPromoCaptureFlag, bishopPromoCaptureFlag, rookPromoCaptureFlag:
			moveList[i].priority = 100 // Promotions are always good
			continue
		case queenPromotionFlag, queenPromoCaptureFlag:
			moveList[i].priority = 101 // Normally queen promo = best promo
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

func getTargetPieceValue(board *Board, move Move, gamePhase int) int {
	targetPosition := getTargetPosition(move)

	// En-passant fix
	if GetFlag(move) == epCaptureFlag {
		if board.GetTopState().IsWhiteTurn {
			targetPosition -= 8
		} else {
			targetPosition += 8
		}
	}

	return GetPieceValue(*board.PieceInfoArr[targetPosition],
		targetPosition,
		gamePhase)
}

func updatePVTable(this_pvPtr int, move Move, depth int8) {
	child_pvPtr := pvPtr
	pvPtr = this_pvPtr
	pv[pvPtr] = move
	pvPtr++
	for i := int8(0); i < depth-1 && pv[child_pvPtr] != NULL_MOVE; i++ { // copy child PV behind it
		pv[pvPtr] = pv[child_pvPtr]
		pvPtr++
		child_pvPtr++
	}
}

/*
	Outputs the engine info from a given search depth following the format:

info depth <depth> seldepth <maxdepth searched> multipv <principal variations> score cp <score> nodes <nodecount> nps <nodes / time> time <time taken in ms> pv <pv>
*/
func engineInfoString() (retval string) {
	elapsedTime := time.Since(latestSearchInfo.startTime)
	nps := int64(float64(latestSearchInfo.nodeCount) / elapsedTime.Seconds())
	hashFill := int(float64(DebugTableSize) / float64(TableCapacity) * 1000)
	// Convert PV chain up to depth to a single string seperated by " "
	pvString := ""
	for i := int8(0); i < latestSearchInfo.depth; i++ {
		if savedPV[i] == NULL_MOVE {
			break
		}
		pvString += MoveToString(savedPV[i]) + " "
	}
	// Strip surrounding whitespace from PV string
	pvString = pvString[:len(pvString)-1]

	retval += fmt.Sprintf("info depth %d seldepth %d multipv %d score cp %d nodes %d nps %d hashfull %d time %d pv %s",
		latestSearchInfo.depth, latestSearchInfo.seldepth+latestSearchInfo.depth, latestSearchInfo.multipv,
		latestSearchInfo.score, latestSearchInfo.nodeCount, nps, hashFill, elapsedTime.Milliseconds(), pvString)

	if DebugMode {
		retval += fmt.Sprintf("\nDebug Info:\n\tpvNodes: %d(%0.2f%%)\n\tallNodes: %d(%0.2f%%)\n\tcutNodes: %d(%0.2f%%)\n\tqNodes: %d\n", latestSearchInfo.pvNodes, 100*float32(latestSearchInfo.pvNodes)/float32(latestSearchInfo.nodeCount), latestSearchInfo.allNodes, 100*float32(latestSearchInfo.allNodes)/float32(latestSearchInfo.nodeCount), latestSearchInfo.cutNodes, 100*float32(latestSearchInfo.cutNodes)/float32(latestSearchInfo.nodeCount), latestSearchInfo.qNodes)
	}

	return retval
}
