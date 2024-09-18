package chessengine

import (
	"fmt"
	"time"
)

var DebugMode = false

const (
	MATE_SCORE                  = -65535
	DRAW_SCORE                  = 0
	MIN_VALUE                   = -2147483648
	MAX_VALUE                   = 2147483648
	MAX_EXTENSION_DEPTH         = 8
	MAX_QSEARCH_DEPTH           = 30
	MAX_SEARCH_DEPTH            = 63
	MAX_POSSIBLE_DEPTH          = MAX_EXTENSION_DEPTH + MAX_SEARCH_DEPTH
	NULL_MOVE_REDUCTION    int8 = 2
	DELTAPRUNE_MARGIN           = 200
	LATE_GAME_PHASE_CUTOFF      = 4
	triangleTableSize           = ((MAX_SEARCH_DEPTH+MAX_EXTENSION_DEPTH)*(MAX_SEARCH_DEPTH+MAX_EXTENSION_DEPTH) + (MAX_POSSIBLE_DEPTH)) / 2
	squareTableSize             = (MAX_POSSIBLE_DEPTH) * (MAX_POSSIBLE_DEPTH)
)

type searchInfo struct {
	startTime time.Time
	depth     int8
	score     int
	seldepth  int8
	multipv   byte
	leafNodes uint64
	debug     debugInfo
}

var prevIterationNodeCount uint64 // for branching factor calculation

type debugInfo struct {
	qNodes           uint64
	qNodeDeltaPrunes uint64

	pvNodes  uint64
	allNodes uint64
	cutNodes uint64

	D1pvNodes  uint64
	D1allNodes uint64
	D1cutNodes uint64

	reducedNodes  uint64
	amountReduced uint64

	siblingNodes uint64

	researchedNodes       uint64
	researchedReduceNodes uint64

	probePVNodes  uint64
	probeALLNodes uint64
	probeCUTNodes uint64

	probePVNodesCorrect  uint64
	probeALLNodesCorrect uint64
	probeCUTNodesCorrect uint64
}

var latestSearchInfo searchInfo

var bestEvalThisIteration int
var currentSearchTurn byte

var savedPV [MAX_POSSIBLE_DEPTH]Move // Principal variation lists

var searchMovePool [MAX_POSSIBLE_DEPTH][MAX_MOVE_COUNT]Move    // Move pool for main-search
var qsearchMovePool [MAX_QSEARCH_DEPTH][MAX_CAPTURE_COUNT]Move // Move pool for quiescence search

var pv [squareTableSize]Move
var pvPtr int

func (board *Board) StartSearchNoDepth(startTime time.Time, cancelChannel chan struct{}) Move {
	return board.initSearch(startTime, MAX_SEARCH_DEPTH, cancelChannel)
}

func (board *Board) StartSearchDepth(startTime time.Time, max_depth int8, cancelChannel chan struct{}) (score Move, nodes uint64, timeTaken int64) {
	return board.initSearch(startTime, max_depth, cancelChannel), latestSearchInfo.leafNodes, time.Since(startTime).Milliseconds()
}

func (board *Board) initSearch(startTime time.Time, max_depth int8, cancelChannel chan struct{}) Move {
	var depth int8 = 1
	pvPtr = 0
	pv = [squareTableSize]Move{}
	savedPV = [MAX_POSSIBLE_DEPTH]Move{}
	resetHistory()
	resetKillers()
	if DebugMode {
		TTDebugReset(board)
	}

	currentSearchTurn = board.GetTopState().TurnCounter
	bestEvalThisIteration = MIN_VALUE

	for depth <= max_depth {

		latestSearchInfo = searchInfo{startTime: startTime, multipv: 1, depth: depth}
		ageHistory()
		pv = [squareTableSize]Move{}
		// killerMovesCounter = [MAX_POSSIBLE_DEPTH][64][64]uint16{}

		board.search(depth, 0, MIN_VALUE, MAX_VALUE, 0, false, true, cancelChannel)

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
func (board *Board) search(depth, plyFromRoot int8, alpha, beta int, numExtensions int8, searchReduced, doNullMove bool, cancelChannel chan struct{}) int {
	// Check if the search has been cancelled
	select {
	case <-cancelChannel:
		return bestEvalThisIteration
	default:
	}

	if plyFromRoot > 0 {
		// Fifty move rule, Insufficient Material, Threefold repetition
		if board.GetTopState().HalfMoveClock >= 100 ||
			isInsufficientMaterial(board) ||
			board.RepetitionPositionHistory[board.GetTopState().ZobristKey] == 3 {
			latestSearchInfo.leafNodes++
			return DRAW_SCORE
		}
	}

	if depth <= 0 {
		eval := board.quiescenceSearch(alpha, beta, 0, cancelChannel)
		latestSearchInfo.leafNodes++
		return eval
	}

	probeScore, probeNodeType, probeMove := probeHash(depth, currentSearchTurn, alpha, beta, board.GetTopState().ZobristKey)
	if probeScore != MIN_VALUE {
		if plyFromRoot == 0 && probeNodeType == PVnode {
			bestEvalThisIteration = probeScore
		}
		return probeScore
	}

	// lazyEval, _, egPhase := board.Evaluate()
	inPVNode := alpha != beta-1

	// Null Move Pruning, https://www.chessprogramming.org/Null_Move_Pruning
	// if doNullMove && !inPVNode && depth > 1 && egPhase != 24 && !board.InCheck() && lazyEval >= beta-50 {
	// 	board.MakeMove(NULL_MOVE)
	// 	nullScore := -board.search(depth-1-NULL_MOVE_REDUCTION, plyFromRoot+1, -beta, -beta+1, numExtensions, searchReduced, false, cancelChannel)
	// 	board.UnMakeMove()
	// 	if nullScore >= beta {
	// 		if nullScore >= -MATE_SCORE-(MAX_SEARCH_DEPTH+MAX_EXTENSION_DEPTH) {
	// 			nullScore = beta
	// 		}

	// 		return nullScore
	// 	}
	// }

	moveList := board.GenerateMoves(ALL, searchMovePool[plyFromRoot][:0])
	board.moveordering(savedPV[plyFromRoot], probeMove, plyFromRoot, moveList)

	if len(moveList) == 0 {
		latestSearchInfo.leafNodes++
		if board.InCheck() {
			return MATE_SCORE + int(plyFromRoot) // Checkmate
		} else {
			return 0 // Stalemate
		}
	}

	if DebugMode {
		switch probeNodeType {
		case PVnode:
			latestSearchInfo.debug.probePVNodes++
		case ALLnode:
			latestSearchInfo.debug.probeALLNodes++
		case CUTnode:
			latestSearchInfo.debug.probeCUTNodes++
		}
	}

	nodeType := ALLnode
	this_pvPtr := pvPtr
	pv[pvPtr] = NULL_MOVE // initialize empty PV
	pvPtr += int(MAX_POSSIBLE_DEPTH)

	wasInCheck := board.InCheck()

	var bestScore int

	{
		move := moveList[0]
		// using fail soft with negamax:
		board.MakeMove(move)
		extension := extendSearch(board, move, numExtensions)
		bestScore = -board.search(depth-1+extension, plyFromRoot+1, -beta, -alpha, numExtensions+extension, false, true, cancelChannel)
		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			return bestEvalThisIteration
		default:
		}

		if bestScore >= beta {
			// If the score is greater than or equal to beta,
			// it means that the opponent has a better move to choose.
			// We record this information in the transposition table.
			recordHash(depth, CUTnode, currentSearchTurn, bestScore, NULL_MOVE, board.GetTopState().ZobristKey)
			pvPtr = this_pvPtr
			// Killer Heuristic, https://www.chessprogramming.org/Killer_Heuristic
			// if !board.InCheck() && move != savedPV[plyFromRoot] {
			// 	updateKiller(plyFromRoot, move)
			// }
			// History Heuristic, https://www.chessprogramming.org/History_Heuristic
			// No need to loop since this is the first move
			updateHistory(board, move, int16(depth*depth))

			if DebugMode {
				latestSearchInfo.debug.cutNodes++
				if depth == 1 {
					latestSearchInfo.debug.D1cutNodes++
				}
				if probeNodeType == CUTnode {
					latestSearchInfo.debug.probeCUTNodesCorrect++
				}
			}
			return bestScore
		}
		if bestScore > alpha { // This move is better than the current best move
			if plyFromRoot == 0 {
				bestEvalThisIteration = bestScore // Update the best evaluation score for this iteration
			}
			updatePVTable(this_pvPtr, move, depth)
			nodeType = PVnode
			alpha = bestScore
		}
	}

	for i, move := range moveList[1:] {

		var score int
		needFullSearch := true

		if DebugMode {
			latestSearchInfo.debug.siblingNodes++
		}

		board.MakeMove(move)

		// Move extensions,
		extension := extendSearch(board, move, numExtensions)

		// Late move reductions, https://www.chessprogramming.org/Late_Move_Reductions
		var reduceAmount int8
		if depth >= 3 && !wasInCheck && !board.InCheck() && extension == 0 && !inPVNode && isQuietMove(move) && !searchReduced {
			reduceAmount = depth / 3 // Senpai engine LMR implementation
			if DebugMode {
				latestSearchInfo.debug.reducedNodes++
			}
			latestSearchInfo.debug.amountReduced += uint64(reduceAmount)
			score = -board.search(depth-1-reduceAmount, plyFromRoot+1, -alpha-1, -alpha, numExtensions+extension, true, true, cancelChannel)
			needFullSearch = (score > alpha && score < beta)
		}

		// PVS Search, https://www.chessprogramming.org/Principal_Variation_Search
		if needFullSearch {
			score = -board.search(depth-1+extension, plyFromRoot+1, -alpha-1, -alpha, numExtensions+extension, (reduceAmount != 0) || searchReduced, true, cancelChannel)
			if DebugMode && reduceAmount != 0 {
				latestSearchInfo.debug.researchedReduceNodes++
			}
			needFullSearch = (score > alpha && score < beta)
		}

		// Full search
		if needFullSearch {
			score = -board.search(depth-1+extension, plyFromRoot+1, -beta, -alpha, numExtensions+extension, false, true, cancelChannel)
			if DebugMode {
				latestSearchInfo.debug.researchedNodes++
			}
			alpha = max(alpha, score)
		}

		board.UnMakeMove()

		// Check if the search has been cancelled
		select {
		case <-cancelChannel:
			return bestEvalThisIteration
		default:
		}

		if score >= beta {
			recordHash(depth, CUTnode, currentSearchTurn, score, NULL_MOVE, board.GetTopState().ZobristKey)
			pvPtr = this_pvPtr
			if DebugMode {
				latestSearchInfo.debug.cutNodes++
				if depth == 1 {
					latestSearchInfo.debug.D1cutNodes++
				}
				if probeNodeType == CUTnode {
					latestSearchInfo.debug.probeCUTNodesCorrect++
				}
			}
			// Killer Heuristic, https://www.chessprogramming.org/Killer_Heuristic
			// if !board.InCheck() && move != savedPV[plyFromRoot] {
			// 	updateKiller(plyFromRoot, move)
			// }
			// History Heuristic, https://www.chessprogramming.org/History_Heuristic
			if updateHistory(board, move, int16(depth*depth)) { // If this move is a quiet move
				// Update all quiet moves before this move to have negative value
				for j := i - 1; j >= 0 && updateHistory(board, moveList[j], -int16(depth*depth)); j-- {
				}
			}
			return score
		}
		if score > bestScore { // This move is better than the current best move
			if plyFromRoot == 0 {
				bestEvalThisIteration = score // Update the best evaluation score for this iteration
			}
			updatePVTable(this_pvPtr, move, depth)
			nodeType = PVnode
			bestScore = score
		}
	}
	if DebugMode {
		if nodeType == ALLnode {
			latestSearchInfo.debug.allNodes++
			if depth == 1 {
				latestSearchInfo.debug.D1allNodes++
			}
			if probeNodeType == ALLnode {
				latestSearchInfo.debug.probeALLNodesCorrect++
			}
		} else if nodeType == PVnode {
			latestSearchInfo.debug.pvNodes++
			if depth == 1 {
				latestSearchInfo.debug.D1pvNodes++
			}
			if probeNodeType == PVnode {
				latestSearchInfo.debug.probePVNodesCorrect++
			}
		}
	}

	pvPtr = this_pvPtr
	recordHash(depth, nodeType, currentSearchTurn, bestScore, pv[this_pvPtr], board.GetTopState().ZobristKey) // Record the best move for this position
	return bestScore
}

func (board *Board) quiescenceSearch(alpha, beta int, plyFromSearch int8, cancelChannel chan struct{}) int {
	select { // Check if the search has been cancelled
	case <-cancelChannel:
		return bestEvalThisIteration
	default:
	}

	latestSearchInfo.debug.qNodes++

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
	}

	if plyFromSearch >= MAX_QSEARCH_DEPTH {
		return alpha
	}

	latestSearchInfo.seldepth = max(plyFromSearch, latestSearchInfo.seldepth)

	captureMoveList := board.GenerateMoves(CAPTURE, qsearchMovePool[plyFromSearch][:0])
	board.quiescence_moveordering(captureMoveList)

	for _, move := range captureMoveList {
		// Delta pruning cut
		if mgPhase > LATE_GAME_PHASE_CUTOFF && getTargetPieceValue(board, move, egPhase) < deltaPrune {
			latestSearchInfo.debug.qNodeDeltaPrunes++
			continue
		}

		board.MakeMove(move)
		eval := -board.quiescenceSearch(-beta, -alpha, plyFromSearch+1, cancelChannel)
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

	// pieceType := board.PieceInfoArr[getTargetPosition(move)].pieceTYPE
	// moveRank := (getTargetPosition(move) & 0b111000) >> 3

	// if pieceType == PAWN && (moveRank == 1 || moveRank == 6) { // Pawn on verge of promotion, also interesting
	// 	return 1
	// }

	return 0
}

func getTargetPieceValue(board *Board, move Move, gamePhase int) int {
	targetPosition := getTargetPosition(move)

	// En-passant fix
	if GetFlag(move) == epCaptureFlag {
		if board.GetTopState().TurnColor == WHITE {
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
	if move.enc == NULL_MOVE.enc {
		panic("PV Table Invalid Update w/ Null move")
	}
	child_pvPtr := pvPtr
	tempPtr := this_pvPtr
	pv[tempPtr] = move
	tempPtr++
	for i := int8(0); i < depth-1 && pv[child_pvPtr] != NULL_MOVE; i++ { // copy child PV behind it
		pv[tempPtr] = pv[child_pvPtr]
		tempPtr++
		child_pvPtr++
	}
}

/*
	Outputs the engine info from a given search depth following the format:

info depth <depth> seldepth <maxdepth searched> multipv <principal variations> score cp <score> nodes <nodecount> nps <nodes / time> time <time taken in ms> pv <pv>
*/
func engineInfoString() (retval string) {
	elapsedTime := time.Since(latestSearchInfo.startTime)
	nps := int64(float64(latestSearchInfo.leafNodes) / elapsedTime.Seconds())
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
	checkmateScore := scoreIsCheckmate(latestSearchInfo.score)

	if checkmateScore != 0 {
		retval += fmt.Sprintf("info depth %d seldepth %d multipv %d score mate %d nodes %d nps %d hashfull %d time %d pv %s",
			latestSearchInfo.depth, latestSearchInfo.seldepth+latestSearchInfo.depth, latestSearchInfo.multipv,
			checkmateScore, latestSearchInfo.leafNodes, nps, hashFill, elapsedTime.Milliseconds(), pvString)
	} else {
		retval += fmt.Sprintf("info depth %d seldepth %d multipv %d score cp %d nodes %d nps %d hashfull %d time %d pv %s",
			latestSearchInfo.depth, latestSearchInfo.seldepth+latestSearchInfo.depth, latestSearchInfo.multipv,
			latestSearchInfo.score, latestSearchInfo.leafNodes, nps, hashFill, elapsedTime.Milliseconds(), pvString)
	}

	if DebugMode {
		totalNodeCount := latestSearchInfo.debug.allNodes + latestSearchInfo.debug.pvNodes + latestSearchInfo.debug.cutNodes
		D1totalNodeCount := latestSearchInfo.debug.D1allNodes + latestSearchInfo.debug.D1pvNodes + latestSearchInfo.debug.D1cutNodes
		retval += fmt.Sprintf("\nDebug Info of nodes:\n\tpvNodes: %d(%0.2f%%)\n\tallNodes: %d(%0.2f%%)\n\tcutNodes: %d(%0.2f%%)\n", latestSearchInfo.debug.pvNodes, 100*float32(latestSearchInfo.debug.pvNodes)/float32(totalNodeCount), latestSearchInfo.debug.allNodes, 100*float32(latestSearchInfo.debug.allNodes)/float32(totalNodeCount), latestSearchInfo.debug.cutNodes, 100*float32(latestSearchInfo.debug.cutNodes)/float32(totalNodeCount))
		retval += fmt.Sprintf("\nDebug Info of depth-1 nodes:\n\tpvNodes: %d(%0.2f%%)\n\tallNodes: %d(%0.2f%%)\n\tcutNodes: %d(%0.2f%%)\n", latestSearchInfo.debug.D1pvNodes, 100*float32(latestSearchInfo.debug.D1pvNodes)/float32(D1totalNodeCount), latestSearchInfo.debug.D1allNodes, 100*float32(latestSearchInfo.debug.D1allNodes)/float32(D1totalNodeCount), latestSearchInfo.debug.D1cutNodes, 100*float32(latestSearchInfo.debug.D1cutNodes)/float32(D1totalNodeCount))
		retval += fmt.Sprintf("\nDebug Info of probe nodes:\n\tpvNodes: %d(Correct: %0.2f%%)\n\tallNodes: %d(Correct: %0.2f%%)\n\tcutNodes: %d(Correct: %0.2f%%)\n", latestSearchInfo.debug.probePVNodes, 100*float32(latestSearchInfo.debug.probePVNodesCorrect)/float32(latestSearchInfo.debug.probePVNodes), latestSearchInfo.debug.probeALLNodes, 100*float32(latestSearchInfo.debug.probeALLNodesCorrect)/float32(latestSearchInfo.debug.probeALLNodes), latestSearchInfo.debug.probeCUTNodes, 100*float32(latestSearchInfo.debug.probeCUTNodesCorrect)/float32(latestSearchInfo.debug.probeCUTNodes))
		retval += fmt.Sprintf("\nDebug Info of quiescence nodes:\n\tqNodes: %d\n\tqNodes delta Pruned: %d(%0.2f%%)\n", latestSearchInfo.debug.qNodes-latestSearchInfo.leafNodes, latestSearchInfo.debug.qNodeDeltaPrunes, 100*float32(latestSearchInfo.debug.qNodeDeltaPrunes)/float32(latestSearchInfo.debug.qNodes-latestSearchInfo.leafNodes))
		retval += fmt.Sprintf("\nDebug Info of sibling nodes:\n\tsiblingNodes: %d\n\tsiblingNodes re-searched: %d(%0.2f%%)\n", latestSearchInfo.debug.siblingNodes, latestSearchInfo.debug.researchedNodes, 100*float32(latestSearchInfo.debug.researchedNodes)/float32(latestSearchInfo.debug.siblingNodes))
		retval += fmt.Sprintf("\nDebug Info of reduced nodes:\n\treducedNodes: %d(%0.2f%%)\n\taverage Reduce Amount: %0.2f\n\treducedNodes re-searched: %d(%0.2f%%)\n", latestSearchInfo.debug.reducedNodes, 100*float32(latestSearchInfo.debug.reducedNodes)/float32(latestSearchInfo.debug.siblingNodes), float32(latestSearchInfo.debug.amountReduced)/float32(latestSearchInfo.debug.reducedNodes), latestSearchInfo.debug.researchedReduceNodes, 100*float32(latestSearchInfo.debug.researchedReduceNodes)/float32(latestSearchInfo.debug.reducedNodes))
		if latestSearchInfo.depth > 1 {
			// Return branching factor in relation to previous iteration
			retval += fmt.Sprintf("\nDebug Info of Effective Branching Factor:\n\t( N(D) / N(D-1) )\n\t%d/%d(%0.2f)\n", totalNodeCount, prevIterationNodeCount, float32(totalNodeCount)/float32(prevIterationNodeCount))
		}
		// Mean/Average Branching Factor
		retval += fmt.Sprintf("\nDebug Info of Average Branching Factor:\n\t # of all nodes / # of non terminal nodes\n\t%d/%d(%0.2f)\n", totalNodeCount+latestSearchInfo.leafNodes, totalNodeCount, float32(totalNodeCount+latestSearchInfo.leafNodes)/float32(totalNodeCount))
		prevIterationNodeCount = totalNodeCount
	}

	return retval
}
