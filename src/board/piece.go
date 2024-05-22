package board

type Piece struct {
	ThisBitBoard *BitBoard
	IsWhite      bool
}

// Constructor for new piece with empty state
func NewPiece() *Piece {
	return &Piece{}
}
