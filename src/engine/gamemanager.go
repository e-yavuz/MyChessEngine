package chessengine

import "math/bits"

const (
	WhiteIsMated byte = iota
	BlackIsMated
	WhiteTimeout
	BlackTimeout
	DrawByArbiter
	FiftyMoveRule
	Repetition
	Stalemate
	InsufficientMaterial
	InProgress
	Error
)

const LightSquares BitBoard = 0x55AA55AA55AA55AA

// // Path: src/gamemanager.go

func isInsufficientMaterial(board *Board) bool {
	// Can't have insufficient material with pawns on the board
	if board.W.Pawn|board.B.Pawn > 0 {
		return false
	}

	// Can't have insufficient material with queens/rooks on the board
	if board.W.Queen|board.B.Queen|board.W.Rook|board.B.Rook > 0 {
		return false
	}

	// If no pawns, queens, or rooks on the board, then consider knight and bishop cases
	whiteBishops := board.W.Bishop
	blackBishops := board.B.Bishop
	whiteKnights := board.W.Knight
	blackKnighs := board.B.Knight
	numWhiteMinors := whiteBishops | whiteKnights
	numBlackMinors := blackBishops | blackKnighs

	numMinors := numWhiteMinors | numBlackMinors

	// Lone kings or King vs King + single minor: is insuffient
	if bits.OnesCount64(numMinors) <= 1 {
		return true
	}

	// Bishop vs bishop: is insufficient when bishops are same colour complex
	if bits.OnesCount64(numMinors) == 2 && bits.OnesCount64(whiteBishops) == 1 && bits.OnesCount64(blackBishops) == 1 {
		whiteBishopIsLightSquare := (whiteBishops & LightSquares) != 0
		blackBishopIsLightSquare := (blackBishops & LightSquares) != 0
		return whiteBishopIsLightSquare == blackBishopIsLightSquare
	}

	return false
}

func GetGameState(board *Board) byte {
	moveList := append(*board.GenerateMoves(CAPTURE), *board.GenerateMoves(QUIET)...)

	// Look for mate/stalemate
	if len(moveList) == 0 {
		if board.InCheck() {
			if board.GetTopState().IsWhiteTurn {
				return WhiteIsMated
			}
			return BlackIsMated
		}
		return Stalemate
	}

	// Fifty move rule
	if board.GetTopState().HalfMoveClock >= 100 {
		return FiftyMoveRule
	}

	// Threefold repetition
	if board.RepetitionPositionHistory[board.GetTopState().ZobristKey] == 3 {
		return Repetition
	}

	// Look for insufficient material
	if isInsufficientMaterial(board) {
		return InsufficientMaterial
	}

	return InProgress
}

func IsWhiteWin(gameResult byte) bool {
	return gameResult == BlackIsMated || gameResult == BlackTimeout
}

func IsBlackWin(gameResult byte) bool {
	return gameResult == WhiteIsMated || gameResult == WhiteTimeout
}

func IsDraw(gameResult byte) bool {
	return gameResult == DrawByArbiter || gameResult == FiftyMoveRule || gameResult == Repetition || gameResult == Stalemate || gameResult == InsufficientMaterial
}
