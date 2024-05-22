package board

import (
	"fmt"
	"strings"
	"testing"
)

func reverseString(s string) string {
	// Convert the string to a slice of runes to handle multi-byte characters.
	runes := []rune(s)
	length := len(runes)

	// Swap the characters from both ends of the slice.
	for i := 0; i < length/2; i++ {
		runes[i], runes[length-1-i] = runes[length-1-i], runes[i]
	}

	// Convert the slice of runes back to a string and return it.
	return string(runes)
}

// printBitboard prints a 64-bit uint64 number as an 8x8 bit grid.
func printBitboard(bb uint64) {
	println()
	for i := 0; i < 8; i++ {
		row := (bb >> (8 * (7 - i))) & 0xFF
		rowBits := fmt.Sprintf("%08b", row)
		rowBitsSpaced := strings.Join(strings.Split(rowBits, ""), " ")
		fmt.Println(reverseString(rowBitsSpaced))
	}
	println()
}

func coordTouint64(row int, col int) uint64 {
	return 1 << coordToPosition(row, col)
}

func coordToPosition(row int, col int) int {
	return (row * 8) + col
}

// Helper function to compare bitboards
func compareBitboards(a, b uint64) bool {
	return a == b
}

func TestCalculateWPawnMoves(t *testing.T) {
	var expected, result uint64
	var pos int

	pos = coordToPosition(0, 0)
	result = calculateWPawnMoves(pos)
	expected = coordTouint64(1, 0)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(0, 7)
	result = calculateWPawnMoves(pos)
	expected = coordTouint64(1, 7)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(6, 7)
	result = calculateWPawnMoves(pos)
	expected = coordTouint64(7, 7)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(7, 7)
	result = calculateWPawnMoves(pos)
	expected = 0
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %d; want %d", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}
}

func TestCalculateBPawnMoves(t *testing.T) {
	var expected, result uint64
	var pos int

	pos = coordToPosition(0, 0)
	result = calculateBPawnMoves(pos)
	expected = 0
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(0, 6)
	result = calculateBPawnMoves(pos)
	expected = 0
	if !compareBitboards(result, expected) {
		t.Errorf("position:(%d) = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(1, 5)
	result = calculateBPawnMoves(pos)
	expected = coordTouint64(0, 5)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(7, 7)
	result = calculateBPawnMoves(pos)
	expected = coordTouint64(6, 7)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %d; want %d", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}
}

func TestCalculateKingMoves(t *testing.T) {
	var expected, result uint64
	var pos int

	pos = coordToPosition(0, 0)
	result = calculateKingMoves(pos)
	expected = coordTouint64(1, 0) | coordTouint64(1, 1) | coordTouint64(0, 1)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(1, 0)
	result = calculateKingMoves(pos)
	expected = coordTouint64(2, 0) | coordTouint64(2, 1) | coordTouint64(1, 1) | coordTouint64(0, 0) | coordTouint64(0, 1)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(1, 1)
	result = calculateKingMoves(pos)
	expected = coordTouint64(2, 1) | coordTouint64(2, 2) | coordTouint64(1, 2) | coordTouint64(0, 2) | coordTouint64(0, 1) | coordTouint64(0, 0) | coordTouint64(1, 0) | coordTouint64(2, 0)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}
}

func TestCalculateQueenMoves(t *testing.T) {
	var expected, result uint64
	var pos int

	pos = coordToPosition(0, 0)
	result = calculateQueenMoves(pos)
	expected = calculateRookMoves(pos) | calculateBishopMoves(pos)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(3, 3)
	result = calculateQueenMoves(pos)
	expected = calculateRookMoves(pos) | calculateBishopMoves(pos)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(7, 7)
	result = calculateQueenMoves(pos)
	expected = calculateRookMoves(pos) | calculateBishopMoves(pos)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}
}

func TestCalculateRookMoves(t *testing.T) {
	var expected, result uint64
	var pos int

	pos = coordToPosition(0, 0)
	result = calculateRookMoves(pos)
	for i := 1; i < 8; i++ {
		expected |= coordTouint64(0, i)
		expected |= coordTouint64(i, 0)
	}
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(3, 3)
	result = calculateRookMoves(pos)
	expected = 0
	for i := 0; i < 8; i++ {
		if i != 3 {
			expected |= coordTouint64(3, i)
			expected |= coordTouint64(i, 3)
		}
	}
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(7, 7)
	result = calculateRookMoves(pos)
	expected = 0
	for i := 0; i < 7; i++ {
		expected |= coordTouint64(7, i)
		expected |= coordTouint64(i, 7)
	}
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}
}

func TestCalculateBishopMoves(t *testing.T) {
	var expected, result uint64
	var pos int

	pos = coordToPosition(0, 0)
	result = calculateBishopMoves(pos)
	for i := 1; i < 8; i++ {
		expected |= coordTouint64(i, i)
	}
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(3, 3)
	result = calculateBishopMoves(pos)
	expected = 0
	for i := 1; i < 4; i++ {
		expected |= coordTouint64(3+i, 3+i)
		expected |= coordTouint64(3-i, 3+i)
		expected |= coordTouint64(3+i, 3-i)
		expected |= coordTouint64(3-i, 3-i)
	}
	expected |= coordTouint64(7, 7)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(7, 7)
	result = calculateBishopMoves(pos)
	expected = 0
	for i := 1; i < 8; i++ {
		expected |= coordTouint64(7-i, 7-i)
	}
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}
}

func TestCalculateKnightMoves(t *testing.T) {
	var expected, result uint64
	var pos int

	pos = coordToPosition(0, 0)
	result = calculateKnightMoves(pos)
	expected = coordTouint64(2, 1) | coordTouint64(1, 2)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(3, 3)
	result = calculateKnightMoves(pos)
	expected = coordTouint64(5, 4) | coordTouint64(5, 2) | coordTouint64(1, 4) | coordTouint64(1, 2) | coordTouint64(4, 5) | coordTouint64(4, 1) | coordTouint64(2, 5) | coordTouint64(2, 1)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}

	pos = coordToPosition(7, 7)
	result = calculateKnightMoves(pos)
	expected = coordTouint64(5, 6) | coordTouint64(6, 5)
	if !compareBitboards(result, expected) {
		t.Errorf("position:%d = %064b; want %064b", pos, result, expected)
		printBitboard(result)
		printBitboard(expected)
	}
}
