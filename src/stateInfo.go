package chessengine

type StateInfo struct {
	EnPassantPosition byte
	IsWhiteTurn       bool

	CastleWKing  bool
	CastleBKing  bool
	CastleWQueen bool
	CastleBQueen bool

	DrawCounter int
	TurnCounter int

	Capture              *PieceInfo
	PrePromotionBitBoard *BitBoard
}
