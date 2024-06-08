package chessengine

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

var openingMap map[uint64]map[Move]bool

const openingBookFileName = "openingbook.txt"

// Opens a folder with games in the following format...
/*
e2e4 e7e5 f2f4 f8c5 g1f3 d7d6 c2c3 c8g4 f1e2 g4f3 e2f3 b8c6 b2b4 c5b6 b4b5 c6e7 d2d4 e5d4 c3d4 g8f6 e1g1 e8g8 c1b2 d6d5 e4e5 f6d7 b1c3 c7c6 d1d3 e7g6 g2g3 f7f5 g1h1 d8e7 c3d5 c6d5 f3d5 g8h8 b2a3 e7d8 d5b7 a8b8 a3f8 d7f8 b7c6 d8d4 d3f5 g6e7 f5c2 b8c8 a1d1 d4b4 d1b1 b4a5 f1f3 e7c6 f3c3 c8d8 b5c6 a5d5 c2g2 d5d1 g2f1 d1d5 f1f3 d5a2 b1d1 d8d1 f3d1 f8e6 d1f3 a2b1 h1g2 g7g6 f3d3 b1g1 g2h3 h7h5 d3c4 g1d1 c3c1 d1g4 h3g2 h5h4 c4f1 g6g5 f1d1 e6f4 g2h1
e2e4 e7e5 g1f3 b8c6 d2d4 e5d4 f3d4 g8f6 b1c3 f8b4 d4c6 b7c6 d1d4 d8e7 f2f3 d7d5 c1g5 e8g8 e1c1 b4c5 g5f6 g7f6 d4a4 c5e3 c1b1 d5d4 c3e2 c6c5 e2c1 c8e6 f1c4 f8b8 c1d3 b8b6
*/
// Reads in each line, and per move calculates a Zobrist key
func openGames(filepath string, max_opening_depth int) {
	openingMap = make(map[uint64]map[uint16]bool)
	InitMagicBitBoardTable("magic_rook", "magic_bishop")
	InitZobristTable()
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Read in each line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parseMoves(line, max_opening_depth)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func parseMoves(line string, max_opening_depth int) {
	board := InitStartBoard()
	moves := strings.Split(line, " ")
	for i := 0; i < max_opening_depth && i < len(moves); i++ {
		move, ok := board.TryMoveUCI(moves[i])

		if ok {
			if _, ok := openingMap[board.GetTopState().ZobristKey]; !ok {
				openingMap[board.GetTopState().ZobristKey] = make(map[Move]bool)
			}
			openingMap[board.GetTopState().ZobristKey][move] = true
			board.MakeMove(move)
		} else {
			return
		}
	}
}

func writeMapToFile() {
	file, err := os.Create(openingBookFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Get a sorted list of keys
	keys := make([]uint64, 0, len(openingMap))
	for k := range openingMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	// Write the keys and moves to the file
	for _, k := range keys {
		// Convert the uint64 key to a string and write it to the file
		file.WriteString(fmt.Sprintf("%d ", k))
		for move := range openingMap[k] {
			file.WriteString(fmt.Sprintf("%d ", move))
		}
		file.WriteString("\n")
	}
	// Print out size of file
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Opening book file size: %d, Memory: %.1f KB\n", len(openingMap), float32(fileInfo.Size())/1024)
	openingMap = nil
}

func CreateOpeningBook(inputFilePath, outputFilePath string, maxDepth int) {
	openGames(inputFilePath, maxDepth)
	writeMapToFile()
}

func (board *Board) GetOpeningBookMove() Move {
	if !board.GetTopState().useOpeningBook {
		return NULL_MOVE
	}

	zobristKey := board.GetTopState().ZobristKey

	file, err := os.Open(openingBookFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Split the line into a key and a list of moves
		parts := strings.Split(line, " ")
		var key uint64
		_, err := fmt.Sscanf(parts[0], "%d", &key)
		if err != nil {
			log.Fatal(err)
		}
		if zobristKey == key {
			// Get the list of moves
			moves := parts[1:]
			// Pick a random move from the list
			move, _ := strconv.Atoi(moves[rand.Intn(len(moves))])
			return Move(move)
		}
	}
	// Could not find opening book move, switch to search
	board.GetTopState().useOpeningBook = false
	return NULL_MOVE
}
