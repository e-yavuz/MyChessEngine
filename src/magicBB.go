package chessengine

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	ROOK = iota
	BISHOP
)

var RookLookupTable [64]map[BitBoard]BitBoard
var BishopLookupTable [64]map[BitBoard]BitBoard

func calculateRookMoves(position int) (retval BitBoard) {
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

func calculateBishopMoves(position int) (retval BitBoard) {
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

func getMoveMask(bitboard BitBoard, position int) BitBoard {
	if position/8 != 0 {
		bitboard &= ^Row1Full
	}
	if position/8 != 7 {
		bitboard &= ^Row8Full
	}
	if position%8 != 0 {
		bitboard &= ^Col1Full
	}
	if position%8 != 7 {
		bitboard &= ^Col8Full
	}
	return bitboard
}

func generateBlockersBitBoard(moveMask BitBoard) []BitBoard {
	indexCount := moveMask
	retvalSize := 1
	for ind := PopLSB(&indexCount); ind != NULL_POSITION; ind = PopLSB(&indexCount) {
		retvalSize <<= 1
	}

	retval := make([]BitBoard, retvalSize)

	for patternIndex := 0; patternIndex < retvalSize; patternIndex++ {
		temp := moveMask
		i := 0
		for bitIndex := PopLSB(&temp); bitIndex != NULL_POSITION; i++ {
			bit := (patternIndex >> i) & 1
			retval[patternIndex] |= uint64(bit) << bitIndex
			bitIndex = PopLSB(&temp)
		}
	}

	return retval
}

func possibleRookMovesBitBoard(position int, blockerPattern BitBoard) (retval BitBoard) {
	dirArr := []Direction{N, E, S, W}
	for _, dir := range dirArr { // 4 directions
		prevMoveMarker := uint64(1) << position
		currentMoveMarker := Shift(prevMoveMarker, dir)
		for out := currentMoveMarker & blockerPattern; currentMoveMarker != 0; out = currentMoveMarker & blockerPattern {
			if (prevMoveMarker&Col1Full != 0 && currentMoveMarker&Col8Full != 0) ||
				(prevMoveMarker&Col8Full != 0 && currentMoveMarker&Col1Full != 0) {
				break
			}
			retval |= currentMoveMarker
			prevMoveMarker = currentMoveMarker
			currentMoveMarker = Shift(currentMoveMarker, dir)
			if out != 0 {
				break
			}
		}
	}

	return retval
}

func possibleBishopMovesBitBoard(position int, blockerPattern BitBoard) (retval BitBoard) {
	dirArr := []Direction{N + E, E + S, S + W, W + N}
	for _, dir := range dirArr { // 4 directions
		prevMoveMarker := uint64(1) << position
		currentMoveMarker := Shift(prevMoveMarker, dir)
		for out := currentMoveMarker & blockerPattern; currentMoveMarker != 0; out = currentMoveMarker & blockerPattern {
			if (prevMoveMarker&Col1Full != 0 && currentMoveMarker&Col8Full != 0) ||
				(prevMoveMarker&Col8Full != 0 && currentMoveMarker&Col1Full != 0) {
				break
			}
			retval |= currentMoveMarker
			prevMoveMarker = currentMoveMarker
			currentMoveMarker = Shift(currentMoveMarker, dir)
			if out != 0 {
				break
			}
		}
	}

	return retval
}

func CreateLookupTables() {
	createRookTable()
	createBishopTable()
}

func createRookTable() {
	for i := 0; i < 64; i++ {
		moveMask := getMoveMask(calculateRookMoves(i), i)
		blockerPatterns := generateBlockersBitBoard(moveMask)
		RookLookupTable[i] = make(map[BitBoard]BitBoard, len(blockerPatterns))
		for _, blockerPattern := range blockerPatterns {
			legalMoveBitBoard := possibleRookMovesBitBoard(i, blockerPattern)
			RookLookupTable[i][blockerPattern] = legalMoveBitBoard
		}
	}
}

func createBishopTable() {
	for i := 0; i < 64; i++ {
		moveMask := getMoveMask(calculateBishopMoves(i), i)
		blockerPatterns := generateBlockersBitBoard(moveMask)
		BishopLookupTable[i] = make(map[BitBoard]BitBoard, len(blockerPatterns))
		for _, blockerPattern := range blockerPatterns {
			legalMoveBitBoard := possibleBishopMovesBitBoard(i, blockerPattern)
			BishopLookupTable[i][blockerPattern] = legalMoveBitBoard
		}
	}
}

func magicRookBlockers() (retval [64][]BitBoard) {
	for i := 0; i < 64; i++ {
		moveMask := getMoveMask(calculateRookMoves(i), i)
		retval[i] = generateBlockersBitBoard(moveMask)
	}
	return retval
}

func magicBishopBlockers() (retval [64][]BitBoard) {
	for i := 0; i < 64; i++ {
		moveMask := getMoveMask(calculateBishopMoves(i), i)
		retval[i] = generateBlockersBitBoard(moveMask)
	}
	return retval
}

func magicNumberGeneration(seed int64, ctx context.Context, filePath string, bestNumbers, bestShifts, bestSizes [64]uint64, blockers [64][]uint64) {
	r := rand.New(rand.NewSource(seed))
	hm := make(map[BitBoard]uint64, 2048)
	for {
		select {
		case <-ctx.Done():
			writeMagicNumberFiles(filePath,
				bestNumbers,
				bestShifts,
				bestSizes)
			return
		default:
			currentNumber := r.Uint64()
			for i := uint64(0); i < 64; i++ {
				checkNumber := currentNumber * (i + 1)
				shift := bestShifts[i]
				for {
					checkNumber++
					stopFlag := false
					maxSize := uint64(0) // Only use max_size, don't need min_index since reducing max_size -> min_index being 0
					for _, blocker := range blockers[i] {
						index := (blocker * currentNumber) >> shift
						if hm[index] == checkNumber { // Found a collision
							stopFlag = true
							break
						}
						hm[index] = checkNumber
						if index > maxSize {
							maxSize = index
						}
					}

					if stopFlag {
						break
					}

					if bestSizes[i] > maxSize || bestSizes[i] == 0 {
						bestSizes[i] = maxSize
						bestShifts[i] = shift
						bestNumbers[i] = currentNumber
						fmt.Printf("[%d] Improved\n\tNumber: %d\n\tShift: %d\n\tSize: %d\n\tMemory: %fKB\n",
							i, currentNumber, shift, maxSize, float64(maxSize)*64/1024)
					}
					shift += 1
				}
			}
		}
	}
}

func magicNumberInitGeneration(seed int64, ctx context.Context, filePath string, blockers [64][]uint64) {
	var empty [64]uint64
	magicNumberGeneration(seed, ctx, filePath, empty, empty, empty, blockers)
}

func magicNumberFileGeneration(seed int64, ctx context.Context, filePath string, blockers [64][]uint64) {
	out, err := readMagicNumberFiles(filePath)
	if err != nil {
		panic(err)
	} else {
		magicNumberGeneration(seed, ctx, filePath, out[0], out[1], out[2], blockers)
	}
}

func InitMagicNumber(seed int64, seconds int, filePath string) {
	ctx, cancel := context.WithCancel(context.Background())

	// Start the task
	go magicNumberInitGeneration(seed, ctx, filePath+"_rook", magicRookBlockers())
	go magicNumberInitGeneration(seed, ctx, filePath+"_bishop", magicBishopBlockers())

	// Start a timer to cancel the task after 3 seconds
	startTimer(cancel, time.Duration(seconds)*time.Second)

	// Wait to allow the task and timer to complete (usually not needed in real applications)
	time.Sleep(time.Duration(seconds+1) * time.Second)
}

func ImproveMagicNumber(seed int64, seconds int, rookPath string, bishopPath string) {
	ctx, cancel := context.WithCancel(context.Background())

	// Start the task
	go magicNumberFileGeneration(seed, ctx, rookPath, magicRookBlockers())
	go magicNumberFileGeneration(seed, ctx, bishopPath, magicBishopBlockers())

	// Start a timer to cancel the task after 3 seconds
	startTimer(cancel, time.Duration(seconds)*time.Second)

	// Wait to allow the task and timer to complete (usually not needed in real applications)
	time.Sleep(time.Duration(seconds+1) * time.Second)
}

// Timer function that cancels the context after a duration
func startTimer(cancel context.CancelFunc, duration time.Duration) {
	go func() {
		time.Sleep(duration)
		cancel()
	}()
}

// Function to write multiple slices of uint64 to a binary file
func writeMagicNumberFiles(filename string, data ...[64]uint64) error {
	// Open the file for writing
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the number of arrays
	err = binary.Write(file, binary.LittleEndian, uint64(len(data)))
	if err != nil {
		return err
	}

	// Write each array to the file
	for _, arr := range data {
		// Write the array data
		err = binary.Write(file, binary.LittleEndian, arr[:])
		if err != nil {
			return err
		}
	}

	return nil
}

// Function to read multiple slices of uint64 from a binary file
func readMagicNumberFiles(filename string) ([][64]uint64, error) {
	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the number of arrays
	var numArrays uint64
	err = binary.Read(file, binary.LittleEndian, &numArrays)
	if err != nil {
		return nil, err
	}

	// Read each array
	data := make([][64]uint64, numArrays)
	for i := uint64(0); i < numArrays; i++ {
		// Read the array data directly into the fixed-size array
		err = binary.Read(file, binary.LittleEndian, &data[i])
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// func InitMagicBitBoardTable(rookFile string, bishopFile string) {

// }
