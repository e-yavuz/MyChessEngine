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
