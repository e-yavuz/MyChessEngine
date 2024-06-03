package chessengine

const (
	zobristSeed = 2361912
)

var zobristPieceArr [6][2][64]uint64
var zobristCastleArr [16]uint64
var zobristEnPassantArr [8]uint64
var zobristWhiteSideToMove uint64

// InitZobristTable initializes the Zobrist table with random values
func InitZobristTable() {
	var x ranctx
	raninit(&x, zobristSeed)

	for i := 0; i < 6; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 64; k++ {
				zobristPieceArr[i][j][k] = ranval(&x)
			}
		}
	}

	for i := 0; i < 16; i++ {
		zobristCastleArr[i] = ranval(&x)
	}

	for i := 0; i < 8; i++ {
		zobristEnPassantArr[i] = ranval(&x)
	}

	zobristWhiteSideToMove = ranval(&x)
}
