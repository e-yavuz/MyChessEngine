package chessengine

type StateInfo struct {
	EnPassantPosition Position
	IsWhiteTurn       bool

	CastleWKing  bool
	CastleBKing  bool
	CastleWQueen bool
	CastleBQueen bool

	DrawCounter int
	TurnCounter int

	PrecedentMove        Move // The move that created the current state, used by UnMakeMove()
	Capture              *PieceInfo
	PrePromotionBitBoard *BitBoard
}

func (si *StateInfo) Equal(other *StateInfo) bool {
	var captureCompare, promotionCompare, valueCompare bool
	if si.Capture != nil && other.Capture != nil {
		captureCompare = si.Capture.Equal(other.Capture)
	} else {
		captureCompare = si.Capture == other.Capture
	}

	if si.PrePromotionBitBoard != nil && other.PrePromotionBitBoard != nil {
		promotionCompare = *si.PrePromotionBitBoard == *other.PrePromotionBitBoard
	} else {
		promotionCompare = si.PrePromotionBitBoard == other.PrePromotionBitBoard
	}

	valueCompare = si.EnPassantPosition == other.EnPassantPosition &&
		si.IsWhiteTurn == other.IsWhiteTurn &&
		si.CastleWKing == other.CastleWKing &&
		si.CastleBKing == other.CastleBKing &&
		si.CastleWQueen == other.CastleWQueen &&
		si.CastleBQueen == other.CastleBQueen &&
		si.DrawCounter == other.DrawCounter &&
		si.TurnCounter == other.TurnCounter

	return captureCompare && promotionCompare && valueCompare
}
