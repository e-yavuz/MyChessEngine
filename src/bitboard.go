package chessengine

// Bitmask for all possible movement for each piece on a 8x8 board
var (
	BPawnMoves [64]uint64
	WPawnMoves [64]uint64

	KnightMoves [64]uint64

	BishopMoves [64]uint64

	RookMoves [64]uint64

	QueenMoves [64]uint64

	KingMoves [64]uint64
)

type BitBoard struct {
	Encoding uint64
}

func init() {
	for i := 0; i < 64; i++ {
		WPawnMoves[i] = calculateWPawnMoves(i)
		BPawnMoves[i] = calculateBPawnMoves(i)
		KnightMoves[i] = calculateKnightMoves(i)
		BishopMoves[i] = calculateBishopMoves(i)
		QueenMoves[i] = calculateQueenMoves(i)
		RookMoves[i] = calculateRookMoves(i)
		KingMoves[i] = calculateKingMoves(i)
	}
}

func (bitboard *BitBoard) PlaceOnBitBoard(position byte) {
	bitboard.Encoding |= 1 << position
}

func (bitboard *BitBoard) RemoveFromBitBoard(position byte) {
	bitboard.Encoding &= (1 << position) - 1
}

func (bitboard *BitBoard) LSBpositions() []byte {
	var value uint64 = bitboard.Encoding
	var retval []byte
	position := byte(0)
	for position < 64 {
		if value&1 == 1 {
			retval = append(retval, position)
		}
		value >>= 1
		position++
	}
	return retval
}

func calculateWPawnMoves(position int) (retval uint64) {
	row := position / 8

	// Example for white pawn moving forward one square
	if row < 7 {
		retval |= 1 << (position + 8)
	}
	// Example for white pawn initial double move
	if row == 1 {
		retval |= 1 << (position + 16)
	}

	return retval
}

func calculateBPawnMoves(position int) (retval uint64) {
	row := position / 8

	// Example for black pawn moving forward one square
	if row > 0 {
		retval |= 1 << (position - 8)
	}
	// Example for black pawn initial double move
	if row == 6 {
		retval |= 1 << (position - 16)
	}

	return retval
}

func calculateKingMoves(position int) (retval uint64) {
	row := position / 8
	col := position % 8

	for _, delta := range []int{-1, 0, 1} {
		for _, epsilon := range []int{-1, 0, 1} {
			if delta == 0 && epsilon == 0 {
				continue
			}
			newRow, newCol := int(row)+delta, int(col)+epsilon
			if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
				retval |= 1 << (uint64(newRow*8 + newCol))
			}
		}
	}
	return retval
}

func calculateQueenMoves(position int) (retval uint64) {
	return calculateRookMoves(position) | calculateBishopMoves(position)
}

func calculateRookMoves(position int) (retval uint64) {
	row := position / 8
	col := position % 8

	// Moves along the row
	for i := 0; i < 8; i++ {
		if i != col {
			retval |= 1 << (row*8 + i)
		}
	}

	// Moves along the column
	for i := 0; i < 8; i++ {
		if i != row {
			retval |= 1 << (i*8 + col)
		}
	}

	return retval
}

func calculateBishopMoves(position int) (retval uint64) {
	row := position / 8
	col := position % 8

	// Diagonals
	for i, j := row, col; i < 8 && j < 8; i, j = i+1, j+1 {
		if i*8+j != position {
			retval |= 1 << (i*8 + j)
		}
	}
	for i, j := row, col; i < 8 && j >= 0; i, j = i+1, j-1 {
		if i*8+j != position {
			retval |= 1 << (i*8 + j)
		}
	}
	for i, j := row, col; i >= 0 && j < 8; i, j = i-1, j+1 {
		if i*8+j != position {
			retval |= 1 << (i*8 + j)
		}
	}
	for i, j := row, col; i >= 0 && j >= 0; i, j = i-1, j-1 {
		if i*8+j != position {
			retval |= 1 << (i*8 + j)
		}
	}

	return retval
}

func calculateKnightMoves(position int) (retval uint64) {
	row := position / 8
	col := position % 8

	deltas := []struct{ dr, dc int }{
		{2, 1}, {2, -1}, {-2, 1}, {-2, -1},
		{1, 2}, {1, -2}, {-1, 2}, {-1, -2},
	}

	for _, delta := range deltas {
		newRow, newCol := row+delta.dr, col+delta.dc
		if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
			retval |= 1 << (newRow*8 + newCol)
		}
	}

	return retval
}
