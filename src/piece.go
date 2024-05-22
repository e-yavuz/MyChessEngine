package chessengine

type PieceInfo struct {
	ThisBitBoard *BitBoard
	IsWhite      bool
}

// Constructor for new piece with empty state
func NewPiece() *PieceInfo {
	return &PieceInfo{}
}
