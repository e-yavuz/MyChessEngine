package chessengine

type PieceInfo struct {
	ThisBitBoard *BitBoard
	IsWhite      bool
}

func (pi *PieceInfo) Equal(other *PieceInfo) bool {
	return pi.ThisBitBoard.Equal(other.ThisBitBoard) && pi.IsWhite == other.IsWhite
}

// Constructor for new piece with empty state
func NewPiece() *PieceInfo {
	return &PieceInfo{}
}
