package chessengine

// Path: src/uci.go
// Basic UCI commands for a chessengine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var AIBoard *Board
var options Options = Options{HashSize: 16, Time_ms: 1000}

type Options struct {
	HashSize int   // in MB, default 16
	Time_ms  int64 // Allowed time for the engine to think in milliseconds
}

// UCI is the main function to start the UCI loop
func UCI() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "uci" {
			UCICommand()
		} else if text == "isready" {
			IsReadyCommand()
		} else if strings.HasPrefix(text, "position") {
			PositionCommand(text)
		} else if strings.HasPrefix(text, "go") {
			GoCommand(text)
		} else if text == "aimove" {
			if AIBoard == nil {
				fmt.Println("No board to make move on")
			} else {
				AIBoard.MakeMove(GoCommand("go"))
				fmt.Println(AIBoard.DisplayBoard())
			}
		} else if strings.HasPrefix(text, "ucinewgame") {
			StartAI()
		} else if strings.HasPrefix(text, "possiblemoves") {
			// Prints possible moves
			if AIBoard == nil {
				fmt.Println("No board to make move on")
			} else {
				possibleMoves := append(*AIBoard.generateMoves(CAPTURE), *AIBoard.generateMoves(QUIET)...)
				for _, move := range possibleMoves {
					fmt.Print(MoveToString(move) + " ")
				}
				fmt.Println()
			}
		} else if strings.HasPrefix(text, "setoption") {
			SetOptions(text)
		} else if strings.HasPrefix(text, "makemove") { // custom command to change the board state for playing
			MakeMoveCommand(text)
			fmt.Println(AIBoard.DisplayBoard())
		} else if text == "quit" {
			break
		} else if text == "aigame" {
			go func() {
				for {
					bestMove := GoCommand("go")
					if bestMove == NULL_MOVE {
						break
					}
					AIBoard.MakeMove(bestMove)
					fmt.Println(AIBoard.DisplayBoard())
				}
			}()
		} else {
			fmt.Println("Unknown command: " + text)
		}
	}
}

// SetOptions sets either the hash size or the time for the engine to think
func SetOptions(text string) {
	text = strings.TrimPrefix(text, "setoption ")
	if strings.HasPrefix(text, "name Hash") {
		text = strings.TrimPrefix(text, "name Hash ")
		hashSize, _ := strconv.Atoi(text)
		options.HashSize = hashSize
	} else if strings.HasPrefix(text, "name Time") {
		text = strings.TrimPrefix(text, "name Time ")
		time_ms, _ := strconv.Atoi(text)
		options.Time_ms = int64(time_ms)
	}
}

// UCICommand is the response to the UCI command
func UCICommand() {
	fmt.Println("id name ChessEngine")
	fmt.Println("id author Emre")
	fmt.Println("uciok")
}

// IsReadyCommand is the response to the isready command
func IsReadyCommand() {
	InitMagicBitBoardTable("magic_rook", "magic_bishop")
	InitZobristTable()
	initPeSTO()
	fmt.Println("readyok")
}

// PositionCommand is the response to the position command
func PositionCommand(text string) {
	text = strings.TrimPrefix(text, "position ")
	if strings.HasPrefix(text, "startpos") {
		AIBoard = InitStartBoard()
		text = strings.TrimPrefix(text, "startpos ")
	} else if strings.HasPrefix(text, "fen") {
		text = strings.TrimPrefix(text, "fen ")
		AIBoard = InitFENBoard(text)
	}

	if strings.HasPrefix(text, "moves") {
		text = strings.TrimPrefix(text, "moves ")
		moves := strings.Split(text, " ")
		for _, moveUCI := range moves {
			if !AIBoard.TryMoveUCI(moveUCI) {
				AIBoard = nil
				fmt.Println("Invalid move: " + moveUCI)
			}
		}
	}
}

func StartAI() {
	AIBoard = nil
}

// GoCommand is the response to the go command, it will do a search using iterative deepening
func GoCommand(text string) Move {
	cancelChannel := make(chan int)

	// Start a timer which send a true value to the cancelChannel after the time is up
	go func() {
		time.Sleep(time.Duration(options.Time_ms) * time.Millisecond)
		fmt.Println("time up")
		close(cancelChannel)
	}()

	var depth byte = 1
	bestMove := NULL_MOVE
	bestScore := -MIN_VALUE

	for {
		foundMove, foundScore := AIBoard.Search(depth, 0, MIN_VALUE, MAX_VALUE, bestMove, cancelChannel)
		if foundMove != NULL_MOVE {
			bestMove = foundMove
			bestScore = foundScore
		} else {
			break
		}
		depth++
	}

	fmt.Printf("bestmove %s\nscore %d, searched depth %d\n", MoveToString(bestMove), bestScore, depth)

	return bestMove
}

// MakeMoveCommand is the response to the makemove command, it will make a move on the board
func MakeMoveCommand(text string) {
	text = strings.TrimPrefix(text, "makemove ")
	if !AIBoard.TryMoveUCI(text) {
		fmt.Println("Invalid move: " + text)
	}
}
