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
	name = "ChessEngineEmre v7b (FULL UCI SUPPORT)"
)

var options Options = Options{HashSize: 16, OwnBook: false}
var uciDebug bool = false
var gameBoard *engine.Board
var searchCancelChannel chan int

type Options struct {
	OwnBook  bool  // Set engine [on/off] to pull from openingbook.txt
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
		ok, err := readCommand(reader)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if !ok {
			break
		}
	}
}

func readCommand(reader *bufio.Reader) (bool, error) {
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "uci" {
		return true, commandUCI()
	} else if text == "isready" {
		return true, commandIsReady()
	} else if strings.HasPrefix(text, "position ") {
		return true, commandPosition(text)
	} else if strings.HasPrefix(text, "go") {
		if gameBoard == nil {
			return true, fmt.Errorf("invalid board")
		}
		go commandGo(text)
		return true, nil
	} else if strings.HasPrefix(text, "stop") {
		close(searchCancelChannel)
		return true, nil
	} else if strings.HasPrefix(text, "ucinewgame") {
		commandUCINewGame()
		return true, nil
	} else if strings.HasPrefix(text, "debug") {
		return true, commandDebug(text)
	} else if strings.HasPrefix(text, "setoption") {
		return true, commandOptions(text)
	} else if strings.HasPrefix(text, "possiblemoves") {
		return true, commandPossibleMoves()
	} else if strings.HasPrefix(text, "move") {
		return true, commandTestGameMakeMove(text)
	} else if strings.HasPrefix(text, "aimove") {
		return true, commandTestGameAI()
	} else if strings.HasPrefix(text, "undomove") {
		return true, commandTestGameUndoMove()
	} else if strings.HasPrefix(text, "eval") {
		return true, commandEval(text)
	} else if text == "help" {
		return true, commandHelp()
	} else if text == "quit" {
		return false, nil
	} else {
		return true, fmt.Errorf("unknown command: %s", text)
	}
}

// commandOptions sets either the hash size or the time for the engine to think
func commandOptions(text string) error {
	text = strings.TrimPrefix(text, "setoption ")
	if strings.HasPrefix(text, "name Hash") {
		text = strings.TrimPrefix(text, "name Hash ")
		hashSize, _ := strconv.Atoi(text)
		options.HashSize = hashSize
	} else if strings.HasPrefix(text, "name OwnBook") {
		text = strings.TrimPrefix(text, "name OwnBook ")
		switch text {
		case "true":
			options.OwnBook = true
		case "false":
			options.OwnBook = false
		default:
			return fmt.Errorf("unvalid OwnBook option, wanted: [true/false], got: %s", text)
		}
	} else {
		return fmt.Errorf("invalid option: %s", text)
	}
	return nil
}

// commandUCI is the response to the UCI command
func commandUCI() error {
	fmt.Println("id name", name)
	fmt.Println("id author Emre")
	fmt.Println("uciok")
	return nil
}

// commandIsReady is the response to the isready command
func commandIsReady() error {
	fmt.Println("readyok")
	return nil
}

func commandUCINewGame() {
	gameBoard = nil
	engine.TTClear()
}

// commandPosition is the response to the position command
func commandPosition(text string) error {
	newBoard := new(engine.Board)
	text = strings.TrimPrefix(text, "position ")
	if strings.HasPrefix(text, "startpos") {
		newBoard = engine.InitStartBoard()
		text = strings.TrimPrefix(text, "startpos ")
	} else if strings.HasPrefix(text, "fen") {
		text = strings.TrimPrefix(text, "fen ")
		newBoard = engine.InitFENBoard(text)
	} else {
		return fmt.Errorf("invalid position: %s", text)
	}

	// Finds index of substring "moves" if it exists in text and
	// gets an Arr of all moves after "moves"
	if strings.Contains(text, "moves") {
		index := strings.Index(text, "moves")
		moves := strings.Split(text[index+len("moves")+1:], " ")
		for _, moveUCI := range moves {
			move, ok := newBoard.TryMoveUCI(moveUCI)
			if !ok {
				return fmt.Errorf("invalid move: %s", moveUCI)
			} else {
				newBoard.MakeMove(move)
			}
		}
	}
	gameBoard = newBoard
	return nil
}

/*
go specifications:
start calculating on the current position set up with the "position" command.
There are a number of commands that can follow this command, all will be sent in the same string.
If one command is not sent its value should be interpreted as it would not influence the search.
  - searchmoves <move1> .... <movei>
    restrict search to this moves only
    Example: After "position startpos" and "go infinite searchmoves e2e4 d2d4"
    the engine should only search the two moves e2e4 and d2d4 in the initial position.
  - ponder
    start searching in pondering mode.
    Do not exit the search in ponder mode, even if it's mate!
    This means that the last move sent in in the position string is the ponder move.
    The engine can do what it wants to do, but after a "ponderhit" command
    it should execute the suggested move to ponder on. This means that the ponder move sent by
    the GUI can be interpreted as a recommendation about which move to ponder. However, if the
    engine decides to ponder on a different move, it should not display any mainlines as they are
    likely to be misinterpreted by the GUI because the GUI expects the engine to ponder
    on the suggested move.
  - wtime <x>
    white has x msec left on the clock
  - btime <x>
    black has x msec left on the clock
  - winc <x>
    white increment per move in mseconds if x > 0
  - binc <x>
    black increment per move in mseconds if x > 0
  - movestogo <x>
    there are x moves to the next time control,
    this will only be sent if x > 0,
    if you don't get this and get the wtime and btime it's sudden death
  - depth <x>
    search x plies only.
  - nodes <x>
    search x nodes only,
  - mate <x>
    search for a mate in x moves
  - movetime <x>
    search exactly x mseconds
  - infinite
    search until the "stop" command. Do not exit the search without being told so in this mode!
*/
func commandGo(text string) (engine.Move, error) {
	if gameBoard == nil {
		return engine.NULL_MOVE, fmt.Errorf("invalid board")
	}
	result := engine.GetGameState(gameBoard)
	if result != engine.InProgress {
		if engine.IsDraw(result) {
			fmt.Println("Game over by:", engine.GameResultToString(result))
		}
		if engine.IsBlackWin(result) {
			fmt.Println("Black Wins!")
		}
		if engine.IsWhiteWin(result) {
			fmt.Println("White Wins!")
		}
		if result == engine.Error {
			fmt.Printf("Game over by: Error")
		}
		return engine.NULL_MOVE, nil
	}

	if options.OwnBook {
		if bookMove := gameBoard.GetOpeningBookMove(); bookMove != engine.NULL_MOVE {
			if uciDebug {
				fmt.Println("Using opening book...")
			}
			fmt.Printf("bestmove %s\n", engine.MoveToString(bookMove))
			return bookMove, nil
		}
	}

	// Reset the search cancel channel to open
	searchCancelChannel = make(chan int)

	var timeInMilliseconds int64 = -1
	var remainingMoves int64 = 1

	// Get the time to search for
	text = strings.TrimPrefix(text, "go ")
	// Split the string into an array of strings
	textArr := strings.Split(text, " ")
	// Loop through the array of strings
	for i := 0; i < len(textArr); i++ {
		switch textArr[i] {
		case "wtime":
			timeInMilliseconds, _ = strconv.ParseInt(textArr[i+1], 10, 64)
		case "btime":
			timeInMilliseconds, _ = strconv.ParseInt(textArr[i+1], 10, 64)
		case "movetime":
			timeInMilliseconds, _ = strconv.ParseInt(textArr[i+1], 10, 64)
			remainingMoves = 1
		case "movestogo":
			remainingMoves, _ = strconv.ParseInt(textArr[i+1], 10, 64)
		}
	}

	timeInMilliseconds = int64(float64(timeInMilliseconds) / float64(remainingMoves))

	if timeInMilliseconds != -1 {
		// Start a timer to close channel and end search at the end
		go func() {
			time.Sleep(time.Duration(timeInMilliseconds-1) * time.Millisecond)
			close(searchCancelChannel)
		}()
	}

	move := gameBoard.StartSearch(time.Now(), searchCancelChannel)

	if uciDebug {
		fmt.Printf("TT occupancy: %0.2f%%, Collisions: %d, NewEntries: %d\n", 100*float32(engine.DebugTableSize)/float32(engine.TableCapacity), engine.DebugCollisions, engine.DebugNewEntries)
	}

	fmt.Printf("bestmove %s\n", engine.MoveToString(move))

	return move, nil
}

// Sets the debug mode to true or false
func commandDebug(text string) error {
	text = strings.TrimPrefix(text, "debug ")
	if text == "on" {
		uciDebug = true
		engine.DebugMode = true
	} else if text == "off" {
		uciDebug = false
		engine.DebugMode = false
	} else {
		return fmt.Errorf("invalid Debug suffix, wanted: [on/off], got: %s", text)
	}
	return nil
}

// If debug on, this returns all possible moves from a position
func commandPossibleMoves() error {
	if !uciDebug {
		return fmt.Errorf("possiblemoves command only available during with debug [on]")
	}
	if gameBoard == nil {
		return fmt.Errorf("invalid board")
	} else {
		moveList := make([]engine.Move, 0, engine.MAX_MOVE_COUNT)
		moveList = gameBoard.GenerateMoves(engine.ALL, moveList)
		for _, move := range moveList {
			fmt.Print(engine.MoveToString(move) + " ")
		}
		fmt.Println()
	}
	return nil
}

func commandTestGameMakeMove(text string) error {
	if !uciDebug {
		return fmt.Errorf("move command only available during with debug [on]")
	}
	if gameBoard == nil {
		return fmt.Errorf("invalid board")
	}

	moveUCI := strings.TrimPrefix(text, "move ")
	move, ok := gameBoard.TryMoveUCI(moveUCI)

	if !ok {
		return fmt.Errorf("invalid move: " + moveUCI)
	} else {
		gameBoard.MakeMove(move)
	}
	result := engine.GetGameState(gameBoard)
	if result != engine.InProgress {
		if engine.IsDraw(result) {
			fmt.Println("Game over by:", engine.GameResultToString(result))
		}
		if engine.IsBlackWin(result) {
			fmt.Println("Black Wins!")
		}
		if engine.IsWhiteWin(result) {
			fmt.Println("White Wins!")
		}
		if result == engine.Error {
			fmt.Printf("Game over by: Error")
		}
		return nil
	}

	aimove, err := commandGo("go")
	if err != nil {
		return err
	}

	gameBoard.MakeMove(aimove)
	fmt.Println(gameBoard.DisplayBoard())
	result = engine.GetGameState(gameBoard)

	if result != engine.InProgress {
		if engine.IsDraw(result) {
			fmt.Println("Game over by:", engine.GameResultToString(result))
		}
		if engine.IsBlackWin(result) {
			fmt.Println("Black Wins!")
		}
		if engine.IsWhiteWin(result) {
			fmt.Println("White Wins!")
		}
		if result == engine.Error {
			fmt.Printf("Game over by: Error")
		}
	}
	return nil
}

func commandTestGameAI() error {
	if !uciDebug {
		return fmt.Errorf("aimove command only available during with debug [on]")
	}
	if gameBoard == nil {
		return fmt.Errorf("invalid board")
	}

	aimove, err := commandGo("go")
	if err != nil {
		return err
	}
	gameBoard.MakeMove(aimove)
	fmt.Println(gameBoard.DisplayBoard())
	result := engine.GetGameState(gameBoard)
	if result != engine.InProgress {
		if engine.IsDraw(result) {
			fmt.Println("Game over by:", engine.GameResultToString(result))
		}
		if engine.IsBlackWin(result) {
			fmt.Println("Black Wins!")
		}
		if engine.IsWhiteWin(result) {
			fmt.Println("White Wins!")
		}
		if result == engine.Error {
			fmt.Printf("Game over by: Error")
		}
	}
	return nil
}

func commandTestGameUndoMove() error {
	if !uciDebug {
		return fmt.Errorf("undomove command only available during with debug [on]")
	}
	if gameBoard == nil {
		return fmt.Errorf("invalid board")
	}

	fmt.Println("Taking back move: " + engine.MoveToString(gameBoard.GetTopState().PrecedentMove))
	gameBoard.UnMakeMove()
	fmt.Println(gameBoard.DisplayBoard())
	return nil
}

func commandEval(text string) error {
	if gameBoard == nil {
		return fmt.Errorf("invalid board")
	}
	//TODOlow make eval search to provided depth and return evaluation fo search
	eval := gameBoard.Evaluate()
	sign := "+"
	if eval < 0 {
		sign = ""
	}
	fmt.Printf("Evaluation: %s%0.2f\n", sign, float32(gameBoard.Evaluate())/100)
	return nil
}

func commandHelp() error {
	fmt.Println("\tuci - Initialize the UCI protocol")
	fmt.Println("\tisready - Check if the engine is ready")
	fmt.Println("\tposition [startpos/fen <fen_string>] [moves <move_list>] - Set up the board position")
	fmt.Println("\tgo - Start searching for the best move")
	fmt.Println("\tucinewgame - Clear the board and reset the game")
	fmt.Println("\tdebug [on/off] - Enable or disable debug mode")
	fmt.Println("\tsetoption")
	fmt.Println("\t\tname Hash <hash_size> - Set the hash table size")
	fmt.Println("\t\tname Time <time_ms> - Set the allowed thinking time")
	fmt.Println("\t\tname OwnBook [on/off] - Sets if engine can use saved book moves")
	fmt.Println("\tpossiblemoves - Display all possible moves from the current position (debug mode only)")
	fmt.Println("\tmove <move_uci> - Make a custom move, followed by the engine's move, on the current board (debug mode only)")
	fmt.Println("\taimove - Tell engine to make best discovered move on the current board (debug mode only)")
	fmt.Println("\tundomove - Undo the last move on the current board (debug mode only)")
	fmt.Println("\teval - Evaluate the current position (debug mode only)")
	fmt.Println("\thelp - Display this help message")
	fmt.Println("\tquit - Exit the program")
	return nil
}
