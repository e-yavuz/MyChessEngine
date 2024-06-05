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

var gameBoard *Board
var options Options = Options{HashSize: 16, Time_ms: 1000}
var uciDebug bool = false

type Options struct {
	HashSize int   // in MB, default 16
	Time_ms  int64 // Allowed time for the engine to think in milliseconds
}

// UCI is the main function to start the UCI loop
func UCI() {
	InitMagicBitBoardTable("magic_rook", "magic_bishop")
	InitZobristTable()
	initPeSTO()
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "uci" {
			go UCICommand()
		} else if text == "isready" {
			IsReadyCommand()
		} else if strings.HasPrefix(text, "position ") {
			PositionCommand(text)
		} else if strings.HasPrefix(text, "go") {
			go GoCommand(text)
		} else if strings.HasPrefix(text, "ucinewgame") {
			StartAI()
		} else if text == "quit" {
			break
		} else if text == "debug" {
			uciDebug = !uciDebug
			fmt.Println("Debug mode: ", uciDebug)
		} else if strings.HasPrefix(text, "setoption") {
			SetOptions(text)
		} else if text == "aimove" && uciDebug {
			if gameBoard == nil {
				fmt.Println("No board to make move on")
			} else {
				gameBoard.MakeMove(GoCommand("go"))
				fmt.Println(gameBoard.DisplayBoard())
			}
		} else if strings.HasPrefix(text, "possiblemoves") && uciDebug {
			// Prints possible moves
			if gameBoard == nil {
				fmt.Println("No board to make move on")
			} else {
				possibleMoves := append(*gameBoard.generateMoves(CAPTURE), *gameBoard.generateMoves(QUIET)...)
				for _, move := range possibleMoves {
					fmt.Print(MoveToString(move) + " ")
				}
				fmt.Println()
			}
		} else if strings.HasPrefix(text, "makemove") && uciDebug { // custom command to change the board state for playing
			MakeMoveCommand(text)
		} else if text == "humangame" && uciDebug {
			for {
				text, _ := reader.ReadString('\n')
				text = strings.TrimSpace(text)
				if text == "quit" {
					break
				} else if strings.HasPrefix(text, "makemove") {
					MakeMoveCommand(text)
					gameBoard.MakeMove(GoCommand("go"))
					fmt.Println(gameBoard.DisplayBoard())
				}
			}
		} else if text == "aionlygame" && uciDebug {
			AIOnlygame()
		} else if text == "eval" && uciDebug {
			fmt.Println(gameBoard.Evaluate())
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
	fmt.Println("readyok")
}

// PositionCommand is the response to the position command
func PositionCommand(text string) {
	text = strings.TrimPrefix(text, "position ")
	if strings.HasPrefix(text, "startpos") {
		gameBoard = InitStartBoard()
		text = strings.TrimPrefix(text, "startpos ")
	} else if strings.HasPrefix(text, "fen") {
		text = strings.TrimPrefix(text, "fen ")
		gameBoard = InitFENBoard(text)
	}

	if strings.HasPrefix(text, "moves") {
		text = strings.TrimPrefix(text, "moves ")
		moves := strings.Split(text, " ")
		for _, moveUCI := range moves {
			move, ok := gameBoard.TryMoveUCI(moveUCI)
			if !ok {
				gameBoard = nil
				fmt.Println("Invalid move: " + moveUCI)
			} else {
				gameBoard.MakeMove(move)
			}
		}
	}
	fmt.Println(gameBoard.DisplayBoard())
}

func StartAI() {
	gameBoard = nil
}

// GoCommand is the response to the go command, it will do a search using iterative deepening
func GoCommand(text string) Move {
	if gameBoard == nil {
		fmt.Println("Error: No board to make move on")
		return NULL_MOVE
	}

	if bookMove := gameBoard.getOpeningBookMove(); bookMove != NULL_MOVE {
		fmt.Printf("bestmove %s\nPulled from opening Book\n", MoveToString(bookMove))
		return bookMove
	}

	cancelChannel := make(chan int)

	// Start a timer which send a true value to the cancelChannel after the time is up
	go func() {
		time.Sleep(time.Duration(options.Time_ms) * time.Millisecond)
		close(cancelChannel)
	}()

	move, eval, depth := gameBoard.StartSearch(cancelChannel)

	fmt.Printf("bestmove %s\nscore %d, searched depth %d\n", MoveToString(move), eval, depth)

	if uciDebug {
		fmt.Printf("TT occupancy: %f\n", float32(TableSize)/TableCapacity)
	}

	return bestMove
}

// MakeMoveCommand is the response to the makemove command, it will make a move on the board
func MakeMoveCommand(text string) {
	text = strings.TrimPrefix(text, "makemove ")
	move, ok := gameBoard.TryMoveUCI(text)
	if ok {
		gameBoard.MakeMove(move)
	} else {
		fmt.Println("Invalid move: " + text)
	}
	fmt.Println(gameBoard.DisplayBoard())
}

func AIOnlygame() {
	for {
		bestMove := GoCommand("go")
		if bestMove == NULL_MOVE {
			fmt.Println("Game over!")
			break
		}
		gameBoard.MakeMove(bestMove)
		fmt.Println(gameBoard.DisplayBoard())
	}
}
