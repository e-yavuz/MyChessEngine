package chessengine

type StateInfo struct {
	ZobristKey           uint64
	Capture              *PieceInfo
	PrePromotionBitBoard *BitBoard
	PrecedentMove        Move // The move that created the current state, used by UnMakeMove()

	EnPassantPosition Position
	CastleState       byte

	HalfMoveClock byte
	TurnCounter   byte

	useOpeningBook bool

	inCheck   bool
	TurnColor int8
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
		si.TurnColor == other.TurnColor &&
		si.CastleState == other.CastleState &&
		si.ZobristKey == other.ZobristKey &&
		si.HalfMoveClock == other.HalfMoveClock &&
		si.TurnCounter == other.TurnCounter &&
		si.inCheck == other.inCheck

	return captureCompare && promotionCompare && valueCompare
}

func (si *StateInfo) setCastleWKing(val bool) {
	if val {
		si.CastleState |= 0b0001
	} else {
		si.CastleState &= 0b1110
	}
}

func (si *StateInfo) setCastleBKing(val bool) {
	if val {
		si.CastleState |= 0b0010
	} else {
		si.CastleState &= 0b1101
	}
}

func (si *StateInfo) setCastleWQueen(val bool) {
	if val {
		si.CastleState |= 0b0100
	} else {
		si.CastleState &= 0b1011
	}
}

func (si *StateInfo) setCastleBQueen(val bool) {
	if val {
		si.CastleState |= 0b1000
	} else {
		si.CastleState &= 0b0111
	}
}

func (si *StateInfo) getCastleWKing() bool {
	return si.CastleState&0b0001 != 0
}

func (si *StateInfo) getCastleBKing() bool {
	return si.CastleState&0b0010 != 0
}

func (si *StateInfo) getCastleWQueen() bool {
	return si.CastleState&0b0100 != 0
}

func (si *StateInfo) getCastleBQueen() bool {
	return si.CastleState&0b1000 != 0
}
