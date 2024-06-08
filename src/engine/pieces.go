package chessengine

type Pieces struct {
	Pawn   BitBoard
	Knight BitBoard
	Rook   BitBoard
	Bishop BitBoard
	Queen  BitBoard
	King   BitBoard
}

func (pieces *Pieces) OccupancyBitBoard() BitBoard {
	return pieces.Pawn | pieces.Knight | pieces.Rook | pieces.Bishop | pieces.Queen | pieces.King
}
