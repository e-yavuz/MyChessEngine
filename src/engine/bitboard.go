package chessengine

import "fmt"

// Bitmask for all possible movement for each piece on a 8x8 board
var (
	Full             BitBoard = 0xFFFFFFFFFFFFFFFF
	Row1Full         BitBoard = 0xFF
	Row2Full         BitBoard = 0xFF00
	Row7Full         BitBoard = 0xFF000000000000
	Row8Full         BitBoard = 0xFF00000000000000
	Col1Full         BitBoard = 0x101010101010101
	Col8Full         BitBoard = 0x8080808080808080
	PromotionFull    BitBoard = Row1Full | Row8Full
	NonPromotionFull BitBoard = ^PromotionFull
	InteriorFull     BitBoard = ^(Row1Full | Row8Full | Col1Full | Col8Full)
)

const (
	INVALID_POSITION Position = 0xFF
)

type BitBoard = uint64

func PlaceOnBitBoard(bitboard *BitBoard, position Position) {
	*bitboard |= 1 << position
}

func RemoveFromBitBoard(bitboard *BitBoard, position Position) {
	*bitboard &= ^(uint64(1) << position)
}

func PopLSB(bitboard *BitBoard) Position {
	if *bitboard == 0 {
		return INVALID_POSITION
	}
	var index Position = 0
	var mask BitBoard = 1
	for (*bitboard & mask) == 0 {
		mask <<= 1
		index++
	}
	*bitboard &= ^mask
	return index
}

func Shift(bitboard BitBoard, dir Direction) (retval BitBoard) {
	if dir > 0 {
		retval = bitboard << uint64(dir)
	} else {
		retval = bitboard >> uint64(dir*-1)
	}
	return retval
}

func BitBoardToString(bitboard BitBoard) (retval string) {
	for i := 0; i < 8; i++ {
		retval += reverseStringHelper(fmt.Sprintf("%08b", (bitboard>>(56-8*i))&0xFF)) + "\n"
	}
	return retval
}

func reverseStringHelper(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

/**
 * Flip a bitboard vertically about the centre ranks.
 * Rank 1 is mapped to rank 8 and vice versa.
 * @param x any bitboard
 * @return bitboard x flipped vertically
 */
func flipVertical(x BitBoard) BitBoard {
	const k1 BitBoard = 0x00FF00FF00FF00FF
	const k2 BitBoard = 0x0000FFFF0000FFFF
	x = ((x >> 8) & k1) | ((x & k1) << 8)
	x = ((x >> 16) & k2) | ((x & k2) << 16)
	x = (x >> 32) | (x << 32)
	return x
}

/**
 * Mirror a bitboard horizontally about the center files.
 * File a is mapped to file h and vice versa.
 * @param x any bitboard
 * @return bitboard x mirrored horizontally
 */
func mirrorHorizontal(x BitBoard) BitBoard {
	const k1 BitBoard = 0x5555555555555555
	const k2 BitBoard = 0x3333333333333333
	const k4 BitBoard = 0x0f0f0f0f0f0f0f0f
	x = ((x >> 1) & k1) + 2*(x&k1)
	x = ((x >> 2) & k2) + 4*(x&k2)
	x = ((x >> 4) & k4) + 16*(x&k4)
	return x
}
