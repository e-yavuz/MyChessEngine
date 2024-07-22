package chessengine

const (
	PAWN = iota
	KNIGHT
	BISHOP
	ROOK
	QUEEN
	KING
	NULL_PIECE
)

type PieceInfo struct {
	thisBitBoard *BitBoard
	color        int8
	pieceTYPE    int
}

type CheckerInfo struct {
	pieceInfo       PieceInfo
	position        Position
	intermediaryRay BitBoard
}

type PinnedPieceInfo struct {
	checkerInfo   CheckerInfo
	pieceInfo     PieceInfo
	position      Position
	possibleQuiet BitBoard
}

func (pi *PieceInfo) Equal(other *PieceInfo) bool {
	return *pi.thisBitBoard == *other.thisBitBoard && pi.color == other.color
}

// Constructor for new piece with empty state
func NewPiece() *PieceInfo {
	return &PieceInfo{}
}
