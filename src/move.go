package chessengine

/*
	Contains info for move + all possible flags
*/

type Move struct {
	Encoding uint16
}

func (move *Move) GetStartingPosition() byte {
	return byte(move.Encoding & 0b111111)
}

func (move *Move) GetTargetPosition() byte {
	return byte((move.Encoding >> 6) & 0b111111)
}

/*
code	promotion	capture	special 1	special 0	kind of move
0	0	0	0	0	quiet moves
1	0	0	0	1	double pawn push
2	0	0	1	0	king castle
3	0	0	1	1	queen castle
4	0	1	0	0	captures
5	0	1	0	1	ep-capture
8	1	0	0	0	knight-promotion
9	1	0	0	1	queen-promotion
11	1	1	0	0	knight-promo capture
13	1	1	0	1	queen-promo capture
*/

const (
	quietFlag              = 0b0000
	doublePawnPushFlag     = 0b0001
	kingCastleFlag         = 0b0010
	queenCastleFlag        = 0b0011
	captureFlag            = 0b0100
	epCaptureFlag          = 0b0101
	knightPromotionFlag    = 0b1000
	queenPromotionFlag     = 0b1011
	knightPromoCaptureFlag = 0b1100
	queenPromoCaptureFlag  = 0b1111
)

func (move *Move) GetFlag() byte {
	return byte(move.Encoding >> 12)
}
