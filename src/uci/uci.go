package chessengine

// Path: src/uci.go
// Basic UCI commands for a chessengine

import (
	"bufio"
	engine "chessengine/src/engine"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	name = "ChessEngineEmre v6 (King Pawn Shield in evaluation)"
)

var options Options = Options{HashSize: 16, Time_ms: 10000}
var uciDebug bool = false
var gameBoard *engine.Board

type Options struct {
	HashSize int   // in MB, default 16
	Time_ms  int64 // Allowed time for the engine to think in milliseconds
}

// UCI is the main function to start the UCI loop
func UCI() {
	engine.InitMagicBitBoardTable("magic_rook", "magic_bishop")
	engine.InitZobristTable()
	engine.InitPeSTO()
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "uci" {
			UCICommand()
		} else if text == "isready" {
			IsReadyCommand()
		} else if strings.HasPrefix(text, "position ") {
			PositionCommand(text)
		} else if strings.HasPrefix(text, "go") {
			GoCommand(text)
		} else if strings.HasPrefix(text, "ucinewgame") {
			gameBoard = nil
		} else if strings.HasPrefix(text, "debug") {
			DebugCommand(text)
		} else if strings.HasPrefix(text, "setoption") {
			SetOptions(text)
		} else if strings.HasPrefix(text, "possiblemoves") && uciDebug {
			PossibleMoves()
		} else if text == "testgame" && uciDebug {
			TestGame()
		} else if text == "eval" && uciDebug {
			fmt.Println(gameBoard.Evaluate())
		} else if text == "quit" {
			break
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
	fmt.Println("id name", name)
	fmt.Println("id author Emre")
	fmt.Println("uciok")
}

// IsReadyCommand is the response to the isready command
func IsReadyCommand() {
	fmt.Println("readyok")
}

// PositionCommand is the response to the position command
func PositionCommand(text string) {
	newBoard := new(engine.Board)
	text = strings.TrimPrefix(text, "position ")
	if strings.HasPrefix(text, "startpos") {
		newBoard = engine.InitStartBoard()
		text = strings.TrimPrefix(text, "startpos ")
	} else if strings.HasPrefix(text, "fen") {
		text = strings.TrimPrefix(text, "fen ")
		newBoard = engine.InitFENBoard(text)
	}

	// Finds index of substring "moves" if it exists in text and
	// gets an Arr of all moves after "moves"
	if strings.Contains(text, "moves") {
		index := strings.Index(text, "moves")
		moves := strings.Split(text[index+len("moves")+1:], " ")
		for _, moveUCI := range moves {
			move, ok := newBoard.TryMoveUCI(moveUCI)
			if !ok {
				fmt.Println("Invalid move: " + moveUCI)
				return
			} else {
				newBoard.MakeMove(move)
			}
		}
	}
	gameBoard = newBoard
}

// GoCommand is the response to the go command, it will do a search using iterative deepening
func GoCommand(text string) engine.Move {
	if gameBoard == nil {
		fmt.Println("Error: No board to make move on")
		return engine.NULL_MOVE
	}

	if bookMove := gameBoard.GetOpeningBookMove(); bookMove != engine.NULL_MOVE {
		fmt.Printf("bestmove %s\n", engine.MoveToString(bookMove))
		if uciDebug {
			fmt.Println("Opening book move")
		}
		return bookMove
	}

	cancelChannel := make(chan int)

	// Start a timer which send a true value to the cancelChannel after the time is up
	go func() {
		time.Sleep(time.Duration(options.Time_ms) * time.Millisecond)
		close(cancelChannel)
	}()

	move, eval, moveChain := gameBoard.StartSearch(cancelChannel)

	fmt.Printf("bestmove %s\n", engine.MoveToString(move))

	if uciDebug {
		// Return best move chain and evaluation
		fmt.Print("Move chain: ")
		for _, move := range moveChain {
			fmt.Print(engine.MoveToString(move) + " ")
		}
		fmt.Printf("\nEvaluation: %d, Searched depth: %d\n", eval, len(moveChain))
	}

	if uciDebug {
		fmt.Printf("TT occupancy: %f\n", float32(engine.TableSize)/engine.TableCapacity)
	}

	return move
}

// Sets the debug mode to true or false
func DebugCommand(text string) {
	text = strings.TrimPrefix(text, "debug ")
	if text == "on" {
		uciDebug = true
		engine.DebugMode = true
	} else if text == "off" {
		uciDebug = false
		engine.DebugMode = false
	}
}

// If debug on, this returns all possible moves from a position
func PossibleMoves() {
	if gameBoard == nil {
		fmt.Println("No board to make move on")
	} else {
		moveList := make([]engine.Move, 0, engine.MAX_MOVE_COUNT)
		moveList = gameBoard.GenerateMoves(engine.ALL, moveList)
		for _, move := range moveList {
			fmt.Print(engine.MoveToString(move) + " ")
		}
		fmt.Println()
	}
}

// Starts a CLI game against the bot
func TestGame() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "quit" {
			break
		} else if strings.HasPrefix(text, "move") {
			moveUCI := strings.TrimPrefix(text, "move ")
			move, ok := gameBoard.TryMoveUCI(moveUCI)
			if !ok {
				continue
			}
			gameBoard.MakeMove(move)
			gameBoard.MakeMove(GoCommand("go"))

			fmt.Println(gameBoard.DisplayBoard())
		} else if strings.HasPrefix(text, "ai") {
			gameBoard.MakeMove(GoCommand("go"))
			fmt.Println(gameBoard.DisplayBoard())
		} else if strings.HasPrefix(text, "undomove") {
			gameBoard.UnMakeMove()
			fmt.Println(gameBoard.DisplayBoard())
		} else {
			fmt.Println("Unknown command: " + text)
		}
	}
}
