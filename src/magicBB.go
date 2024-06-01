package chessengine

import (
	"context"
	"encoding/binary"
	"log"
	"math/rand"
	"os"
	"time"
)

var bishopCalculatedMoves [64]BitBoard // pre-calculated moves for a bishop on an empty 8x8 board
var ROOK_MAGIC_SLICE [64][]BitBoard
var BISHOP_MAGIC_SLICE [64][]BitBoard
var magicRookNumbers, magicRookShifts, magicBishopNumbers, magicBishopShifts [64]uint64

type ValidMoveGenerationFunc func(Position, BitBoard) BitBoard
type GetPossiblesFunc func(Position) BitBoard

func init() {
	for position := 0; position < 64; position++ {
		bishopMoves := uint64(0)
		row := (position & 0b111000) >> 3
		col := position & 0b000111

		// Diagonals
		for i, j := row, col; i < 8 && j < 8; i, j = i+1, j+1 {
			if i*8+j != position {
				bishopMoves |= 1 << (i*8 + j)
			}
		}
		for i, j := row, col; i < 8 && j >= 0; i, j = i+1, j-1 {
			if i*8+j != position {
				bishopMoves |= 1 << (i*8 + j)
			}
		}
		for i, j := row, col; i >= 0 && j < 8; i, j = i-1, j+1 {
			if i*8+j != position {
				bishopMoves |= 1 << (i*8 + j)
			}
		}
		for i, j := row, col; i >= 0 && j >= 0; i, j = i-1, j-1 {
			if i*8+j != position {
				bishopMoves |= 1 << (i*8 + j)
			}
		}

		bishopCalculatedMoves[position] = bishopMoves
	}
}

// returns bitboard of all possible rook positions (excluding current position) using pre-calculated consts
func RookMask(position Position) (retval BitBoard) {
	row := (position & 0b111000) >> 3
	col := position & 0b000111

	retval |= (Col1Full<<col | Row1Full<<(row*8)) & ^(uint64(1) << position)

	return SlidingMask(retval, position)
}

// returns bitboard of all possible bishop positions (excluding current position) using pre-calculated array
func BishopMask(position Position) (retval BitBoard) {
	return SlidingMask(bishopCalculatedMoves[position], position)
}

func SlidingMask(bitboard BitBoard, position Position) BitBoard {
	row := (position & 0b111000) >> 3
	col := position & 0b000111
	if row != 0 {
		bitboard &= ^Row1Full
	}
	if row != 7 {
		bitboard &= ^Row8Full
	}
	if col != 0 {
		bitboard &= ^Col1Full
	}
	if col != 7 {
		bitboard &= ^Col8Full
	}
	return bitboard
}

func generateBlockersBitBoard(moveMask BitBoard) []BitBoard {
	indexCount := moveMask
	retvalSize := 1
	for ind := PopLSB(&indexCount); ind != INVALID_POSITION; ind = PopLSB(&indexCount) {
		retvalSize <<= 1
	}

	retval := make([]BitBoard, retvalSize)

	for patternIndex := 0; patternIndex < retvalSize; patternIndex++ {
		temp := moveMask
		i := 0
		for bitIndex := PopLSB(&temp); bitIndex != INVALID_POSITION; i++ {
			bit := (patternIndex >> i) & 1
			retval[patternIndex] |= uint64(bit) << bitIndex
			bitIndex = PopLSB(&temp)
		}
	}

	return retval
}

func manualValidRookMovesBitBoard(position Position, blockerPattern BitBoard) (retval BitBoard) {
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

func manualValidBishopMovesBitBoard(position Position, blockerPattern BitBoard) (retval BitBoard) {
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

func magicRookBlockers() (retval [64][]BitBoard) {
	for i := Position(0); i < 64; i++ {
		moveMask := RookMask(i)
		retval[i] = generateBlockersBitBoard(moveMask)
	}
	return retval
}

func magicBishopBlockers() (retval [64][]BitBoard) {
	for i := Position(0); i < 64; i++ {
		moveMask := BishopMask(i)
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
						log.Printf("[%d] Improved\n\tNumber: %d\n\tShift: %d\n\tSize: %d\n\tMemory: %fKB\n",
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
	numbers, shifts, sizes, err := readMagicNumberFiles(filePath)
	if err != nil {
		panic(err)
	} else {
		magicNumberGeneration(seed, ctx, filePath, numbers, shifts, sizes, blockers)
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
// for magic file, output order is Number, Shift, and Size (1, 2, and 3 return values respectively)
func readMagicNumberFiles(filename string) ([64]uint64, [64]uint64, [64]uint64, error) {
	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		return [64]uint64{}, [64]uint64{}, [64]uint64{}, err
	}
	defer file.Close()

	// Read the number of arrays
	var numArrays uint64
	err = binary.Read(file, binary.LittleEndian, &numArrays)
	if err != nil {
		return [64]uint64{}, [64]uint64{}, [64]uint64{}, err
	}

	// Read each array
	var data [3][64]uint64
	for i := uint64(0); i < numArrays; i++ {
		// Read the array data directly into the fixed-size array
		err = binary.Read(file, binary.LittleEndian, &data[i])
		if err != nil {
			return [64]uint64{}, [64]uint64{}, [64]uint64{}, err
		}
	}

	return data[0], data[1], data[2], nil
}

func InitMagicBitBoardTable(rookFile string, bishopFile string) {
	rookMagicNumbers, rookMagicShift, rookMagicSizes, err := readMagicNumberFiles(rookFile)
	if err != nil {
		panic(err)
	}
	bishopMagicNumbers, bishopMagicShift, bishopMagicSizes, err := readMagicNumberFiles(bishopFile)
	if err != nil {
		panic(err)
	}
	// Sets up the global magic Arrays
	createMagicArray(&rookMagicNumbers, &rookMagicShift, &rookMagicSizes, &ROOK_MAGIC_SLICE, RookMask, manualValidRookMovesBitBoard)
	createMagicArray(&bishopMagicNumbers, &bishopMagicShift, &bishopMagicSizes, &BISHOP_MAGIC_SLICE, BishopMask, manualValidBishopMovesBitBoard)
	// Moves local variables to global, protected variables for use in Get functions
	magicRookNumbers = rookMagicNumbers
	magicRookShifts = rookMagicShift
	magicBishopNumbers = bishopMagicNumbers
	magicBishopShifts = bishopMagicShift
}

func createMagicArray(magicNumbers, magicShifts, magicSizes *[64]uint64, globalMagicArray *[64][]BitBoard, getMoves GetPossiblesFunc, possibleMoves ValidMoveGenerationFunc) {
	for i := Position(0); i < 64; i++ {
		moveMask := getMoves(i)
		blockerPatterns := generateBlockersBitBoard(moveMask)
		(*globalMagicArray)[i] = make([]uint64, (*magicSizes)[i]+1)
		for _, blockerPattern := range blockerPatterns {
			legalMoveBitBoard := possibleMoves(i, blockerPattern)
			(*globalMagicArray)[i][(blockerPattern*(*magicNumbers)[i])>>(*magicShifts)[i]] = legalMoveBitBoard
		}
	}
}

// Multiplies blocker bit board by corresponding magic Number, shifts with magic shift, then returns legal move BitBoard
// Assumes input bitBoard and position are valid, breaks otherwise
func GetRookMoves(position Position, blockerBitBoard BitBoard) BitBoard {
	if position == INVALID_POSITION {
		panic("wtf")
	}
	shiftedIndex := (blockerBitBoard * magicRookNumbers[position]) >> magicRookShifts[position]
	return ROOK_MAGIC_SLICE[position][shiftedIndex]
}

// Multiplies blocker bit board by corresponding magic Number, shifts with magic shift, then returns legal move BitBoard
// Assumes input bitBoard and position are valid, breaks otherwise
func GetBishopMoves(position Position, blockerBitBoard BitBoard) BitBoard {
	shiftedIndex := (blockerBitBoard * magicBishopNumbers[position]) >> magicBishopShifts[position]
	return BISHOP_MAGIC_SLICE[position][shiftedIndex]
}
