package chessengine

import (
	"cmp"
	"fmt"
	"slices"
)

// Basic move-ordering constants
const (
	PV_MOVE_SCORE      int16 = 32767
	TT_MOVE_SCORE      int16 = 32766
	QUEEN_PROMO_SCORE  int16 = 32765
	NORMAL_PROMO_SCORE int16 = 32764
)

const CAPTURE_OFFSET int16 = 1000 // Used for CAPTURE ORDERING

// BASIC MVV_LVA constants
// var basic_mvv_lvaTable = [30]int16{
// 	19, 11, 9, 8, 1, 0, // victim P, attacker P, N, B, R, Q, K
// 	23, 18, 14, 10, 4, 2, // victim N, attacker P, N, B, R, Q, K
// 	24, 20, 17, 12, 5, 3, // victim B, attacker P, N, B, R, Q, K
// 	26, 22, 21, 16, 7, 6, // victim R, attacker P, N, B, R, Q, K
// 	29, 28, 27, 25, 15, 13, // victim Q, attacker P, N, B, R, Q, K
// }

const BASIC_MVV_TABLE_ROW_SIZE = 6

// DEFENDED MVV_LVA constants
var mvv_lvaTable = [25]int16{
	14, 7, 5, 4, 0, // victim P, attacker P, N, B, R, Q
	24, 13, 9, 6, 1, // victim N, attacker P, N, B, R, Q
	25, 15, 12, 8, 2, // victim B, attacker P, N, B, R, Q
	39, 23, 22, 11, 3, // victim R, attacker P, N, B, R, Q
	48, 47, 46, 38, 10, // victim Q, attacker P, N, B, R, Q
}

const MVV_LVA_TABLE_ROW_SIZE int = 5

// UNDEFENDED MVV_LVA constants
var blind_mvv_lvaTable = [30]int16{
	21, 20, 19, 18, 17, 16, // victim P, attacker P, N, B, R, Q, K
	31, 30, 29, 28, 27, 26, // victim N, attacker P, N, B, R, Q, K
	37, 36, 35, 34, 33, 32, // victim B, attacker P, N, B, R, Q, K
	45, 44, 43, 42, 41, 40, // victim R, attacker P, N, B, R, Q, K
	54, 53, 52, 51, 50, 49, // victim Q, attacker P, N, B, R, Q, K
}

const BLIND_TABLE_ROW_SIZE int = 6

// History heuristic variables
var history [2][6][64]int16

// History heuristic constants
const (
	HISTORY_MULTIPLIER     int16 = 32
	HISTORY_AGE_MULTIPLIER int16 = 31
	HISTORY_MAX_HISTORY    int16 = 31 * 31
)

// Killer heuristic variables
var killerMoves [MAX_POSSIBLE_DEPTH][2]Move
var killerMovesCounter [MAX_POSSIBLE_DEPTH][64][64]uint16

const KILLER_MOVE_SCORE int16 = 997

func init() {
	// for i := range basic_mvv_lvaTable {
	// 	basic_mvv_lvaTable[i] += CAPTURE_OFFSET
	// }
	for i := range mvv_lvaTable {
		mvv_lvaTable[i] += CAPTURE_OFFSET
	}
	for i := range blind_mvv_lvaTable {
		blind_mvv_lvaTable[i] += CAPTURE_OFFSET
	}
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
func (board *Board) moveordering(PVMove Move, TTMove Move, plyFromRoot int8, moveList []Move) {
	for i := range moveList {
		if moveList[i].enc == PVMove.enc {
			moveList[i].priority = PV_MOVE_SCORE
			continue
		}
		if moveList[i].enc == TTMove.enc {
			moveList[i].priority = TT_MOVE_SCORE
			continue
		}

		switch GetFlag(moveList[i]) {
		case captureFlag:
			moveList[i].priority = board.mvv_lva_score(board.PieceInfoArr[getStartingPosition(moveList[i])].pieceTYPE,
				board.PieceInfoArr[getTargetPosition(moveList[i])].pieceTYPE, getTargetPosition(moveList[i]))
		case epCaptureFlag:
			moveList[i].priority = board.mvv_lva_score(board.PieceInfoArr[getStartingPosition(moveList[i])].pieceTYPE,
				PAWN, getTargetPosition(moveList[i]))
		case knightPromotionFlag, bishopPromotionFlag, rookPromotionFlag, knightPromoCaptureFlag, bishopPromoCaptureFlag, rookPromoCaptureFlag:
			moveList[i].priority = NORMAL_PROMO_SCORE // Promotions are always good
			continue
		case queenPromotionFlag, queenPromoCaptureFlag:
			moveList[i].priority = QUEEN_PROMO_SCORE // Normally queen promo = best promo
			continue
		default: // Quiet moves
			moveList[i].priority = getKiller(plyFromRoot, moveList[i])
			if moveList[i].priority == 0 {
				moveList[i].priority = getHistory(board, moveList[i])
			}
		}

	}

	slices.SortFunc(moveList, func(a, b Move) int {
		return cmp.Compare(b.priority, a.priority)
	})
}

func (board *Board) quiescence_moveordering(moveList []Move) {
	for i := range moveList {
		switch GetFlag(moveList[i]) {
		case captureFlag:
			moveList[i].priority = board.mvv_lva_score(board.PieceInfoArr[getStartingPosition(moveList[i])].pieceTYPE,
				board.PieceInfoArr[getTargetPosition(moveList[i])].pieceTYPE, getTargetPosition(moveList[i]))
		case epCaptureFlag:
			moveList[i].priority = board.mvv_lva_score(board.PieceInfoArr[getStartingPosition(moveList[i])].pieceTYPE,
				PAWN, getTargetPosition(moveList[i]))
		case knightPromoCaptureFlag, bishopPromoCaptureFlag, rookPromoCaptureFlag:
			moveList[i].priority = NORMAL_PROMO_SCORE // Promotions are always good
			continue
		case queenPromoCaptureFlag:
			moveList[i].priority = QUEEN_PROMO_SCORE // Normally queen promo = best promo
			continue
		default: // Quiet Moves
			panic("Quiessence move ordering should only be used for captures, but this flag is not a capture!")
		}
	}

	slices.SortFunc(moveList, func(a, b Move) int {
		return cmp.Compare(b.priority, a.priority)
	})
}

func (board *Board) mvv_lva_score(aggressorPieceType, victimPieceType int, capturePosition Position) (priority int16) {
	if board.isAttacked(capturePosition, board.GetTopState().TurnColor^1) {
		// if (victimPieceType*MVV_LVA_TABLE_ROW_SIZE)+aggressorPieceType > len(basic_mvv_lvaTable) {
		// 	fmt.Println(board.DisplayBoard())
		// 	panic("wtf")
		// }
		return mvv_lvaTable[(victimPieceType*MVV_LVA_TABLE_ROW_SIZE)+aggressorPieceType]
	} else { /* BLIND (Better or Lesser If Not Defended) CAPTURES, if the victim is undefended, then the priority is increased */
		return blind_mvv_lvaTable[(victimPieceType*BLIND_TABLE_ROW_SIZE)+aggressorPieceType]
	}
	// return basic_mvv_lvaTable[(victimPieceType*BASIC_MVV_TABLE_ROW_SIZE)+aggressorPieceType]
}

func updateHistory(board *Board, move Move, bonus int16) bool {
	if isQuietMove(move) {
		side2move := board.GetTopState().TurnColor
		piece := board.PieceInfoArr[getStartingPosition(move)].pieceTYPE
		to := getTargetPosition(move)
		clampedBonus := max(min(bonus, HISTORY_MAX_HISTORY), -HISTORY_MAX_HISTORY)

		history[side2move][piece][to] +=
			clampedBonus -
				(history[side2move][piece][to] * max(-clampedBonus, clampedBonus) / HISTORY_MAX_HISTORY)
		return true
	}
	return false
}

func getHistory(board *Board, move Move) int16 {
	side2move := board.GetTopState().TurnColor
	piece := board.PieceInfoArr[getStartingPosition(move)].pieceTYPE
	to := getTargetPosition(move)
	return history[side2move][piece][to]
}

func resetHistory() {
	history = [2][6][64]int16{}
}

func ageHistory() {
	for piece := PAWN; piece <= KING; piece++ {
		for to := A1; to <= H8; to++ {
			history[WHITE][piece][to] = (history[WHITE][piece][to] * HISTORY_AGE_MULTIPLIER) / HISTORY_MULTIPLIER
			history[BLACK][piece][to] = (history[BLACK][piece][to] * HISTORY_AGE_MULTIPLIER) / HISTORY_MULTIPLIER
		}
	}
}

func resetKillers() {
	killerMoves = [MAX_POSSIBLE_DEPTH][2]Move{}
	killerMovesCounter = [MAX_POSSIBLE_DEPTH][64][64]uint16{}
}

func updateKiller(depth int8, move Move) {
	if !isQuietMove(move) {
		return
	}

	firstMove := killerMoves[depth][0]
	secondMove := killerMoves[depth][1]

	if move.enc == firstMove.enc {
		return
	}

	killerMovesCounter[depth][getStartingPosition(move)][getTargetPosition(move)] = min(65535, killerMovesCounter[depth][getStartingPosition(move)][getTargetPosition(move)]+1)

	if killerMovesCounter[depth][getStartingPosition(move)][getTargetPosition(move)] == 65535 {
		fmt.Println("Killer move counter overflow!")
	}

	if killerMovesCounter[depth][getStartingPosition(move)][getTargetPosition(move)] >
		killerMovesCounter[depth][getStartingPosition(firstMove)][getTargetPosition(firstMove)] {
		killerMoves[depth][1] = killerMoves[depth][0]
		killerMoves[depth][0] = move
	} else if killerMovesCounter[depth][getStartingPosition(move)][getTargetPosition(move)] >
		killerMovesCounter[depth][getStartingPosition(secondMove)][getTargetPosition(secondMove)] {
		killerMoves[depth][1] = move
	}
}

func getKiller(depth int8, move Move) int16 {
	switch move.enc {
	case killerMoves[depth][0].enc:
		return KILLER_MOVE_SCORE + 2
	case killerMoves[depth][1].enc:
		return KILLER_MOVE_SCORE + 1
	default:
		return 0
	}
}
