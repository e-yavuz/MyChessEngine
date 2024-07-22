package chessengine

import (
	"cmp"
	"slices"
)

const INTERMEDIARY_SPACE = 4 // Used for history heuristic, killer moves, etc.
const CAPTURE_OFFSET = 10
const BLIND_CAPTURE_SCORE = 6 // Used for if the low value victim is undefended

var mvv_lvaTable = [20]int8{
	6, 9, 12, 16, // PAWN AGGRESSOR
	2, 6, 11, 15, // MINOR_PIECE AGGRESSOR
	1, 4, 6, 14, // ROOK AGGRESSOR
	0, 3, 5, 6, // QUEEN AGGRESSOR
	7, 8, 10, 13, // KING AGGRESSOR
} //P  M  R  Q
// 	   VICTIMS

func init() {
	for i := range mvv_lvaTable {
		mvv_lvaTable[i] += CAPTURE_OFFSET
		aggressor_val := i / 4
		victim_val := i % 4
		if aggressor_val <= victim_val {
			mvv_lvaTable[i] += INTERMEDIARY_SPACE + BLIND_CAPTURE_SCORE
		}
	}
}

// Basically just combines K + B into a single piece type
func pieceTypeToMVVTableType(pieceType int) int {
	if pieceType > KNIGHT {
		return pieceType - 1
	}
	return pieceType
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
func (board *Board) moveordering(PVMove Move, TTMove Move, moveList []Move) {
	for i := range moveList {
		if moveList[i].enc == PVMove.enc {
			moveList[i].priority = 127
			continue
		}
		if moveList[i].enc == TTMove.enc {
			moveList[i].priority = 126
			continue
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

		mvv_lva_capturingPiece := pieceTypeToMVVTableType(board.PieceInfoArr[getStartingPosition(moveList[i])].pieceTYPE)
		mvv_lva_takenPiece := pieceTypeToMVVTableType(takenPiece)

		// Access MVV_LVA list
		moveList[i].priority = mvv_lvaTable[(mvv_lva_capturingPiece*4)+mvv_lva_takenPiece]
	}

	slices.SortFunc(moveList, func(a, b Move) int {
		return cmp.Compare(b.priority, a.priority)
	})
}

func (board *Board) quiescence_moveordering(moveList []Move) {
	for i := range moveList {
		var takenPiece int

		switch GetFlag(moveList[i]) {
		case captureFlag:
			takenPiece = board.PieceInfoArr[getTargetPosition(moveList[i])].pieceTYPE
		case epCaptureFlag:
			takenPiece = PAWN
		case knightPromoCaptureFlag, bishopPromoCaptureFlag, rookPromoCaptureFlag:
			moveList[i].priority = 100 // Promotions are always good
			continue
		case queenPromoCaptureFlag:
			moveList[i].priority = 101 // Normally queen promo = best promo
			continue
		default:
			panic("Quiessence move ordering should only be used for captures, but this flag is not a capture!")
		}

		aggressor := pieceTypeToMVVTableType(board.PieceInfoArr[getStartingPosition(moveList[i])].pieceTYPE)
		victim := pieceTypeToMVVTableType(takenPiece)

		// Access MVV_LVA list
		moveList[i].priority = board.MVV_LVA_score(aggressor, victim, getTargetPosition(moveList[i]))
	}

	slices.SortFunc(moveList, func(a, b Move) int {
		return cmp.Compare(b.priority, a.priority)
	})
}

func (board *Board) MVV_LVA_score(aggressor, victim int, capturePosition Position) (priority int8) {
	/* BLIND (Better or Lesser If Not Defended) CAPTURES, if the victim is undefended, then the priority is increased */
	if aggressor > victim && !board.isAttacked(capturePosition, board.GetTopState().TurnColor^1) {
		priority += BLIND_CAPTURE_SCORE
	}

	priority += mvv_lvaTable[(aggressor*4)+victim]
	return priority
}
