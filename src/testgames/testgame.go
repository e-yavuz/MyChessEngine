package chessengine

import (
	engine "chessengine/src/engine"
	"fmt"
	"io"
	"os/exec"
)

var gameBoard *engine.Board

// Start a game for
func StartGame(engine1 string, engine2 string, fenBoards []string) (int, int, int, int) {
	engine.InitMagicBitBoardTable("magic_rook", "magic_bishop")
	engine.InitZobristTable()
	engine.InitPeSTO()
	var engine1Wins, engine2Wins, draws, errors int
	var err error
	moveList := []string{}
	buf := make([]byte, 1024)
	// Command to execute
	e1 := exec.Command(fmt.Sprintf("./%s", engine1))
	e2 := exec.Command(fmt.Sprintf("./%s", engine2))

	// Create pipes for standard input and output
	stdout1, _ := e1.StdoutPipe()
	stdin1, _ := e1.StdinPipe()
	stdout2, _ := e2.StdoutPipe()
	stdin2, _ := e2.StdinPipe()

	// Start the engines
	err = e1.Start()
	if err != nil {
		fmt.Println("Error starting engine1:", err)
		return -1, -1, -1, -1
	}
	err = e2.Start()
	if err != nil {
		fmt.Println("Error starting engine2:", err)
		return -1, -1, -1, -1
	}

	// Wait for the engines to start
	_, err = stdin1.Write([]byte("uci\n"))
	if err != nil {
		fmt.Println("Error writing to engine1:", err)
		return -1, -1, -1, -1
	}

	n, err := stdout1.Read(buf)
	if err != nil {
		fmt.Println("Error reading from engine1:", err)
		return -1, -1, -1, -1
	}
	fmt.Println("Engine1:", string(buf[:n]))

	_, err = stdin2.Write([]byte("uci\n"))
	if err != nil {
		fmt.Println("Error writing to engine2:", err)
		return -1, -1, -1, -1
	}

	n, err = stdout2.Read(buf)
	if err != nil {
		fmt.Println("Error reading from engine2:", err)
		return -1, -1, -1, -1
	}
	fmt.Println("Engine2:", string(buf[:n]))

	// Start the game
	_, err = stdin1.Write([]byte("ucinewgame\n"))
	if err != nil {
		fmt.Println("Error writing to engine1:", err)
		return -1, -1, -1, -1
	}
	_, err = stdin2.Write([]byte("ucinewgame\n"))
	if err != nil {
		fmt.Println("Error writing to engine2:", err)
		return -1, -1, -1, -1
	}

	// e1 is white
	for _, fen := range fenBoards {
		gameBoard = engine.InitFENBoard(fen)
		moveList = []string{}
		for {
			// Make a turn for engine1
			result := makeTurn(stdin1, stdout1, fen, &moveList)
			if result != engine.InProgress {
				if engine.IsDraw(result) {
					draws++
				}
				if engine.IsWhiteWin(result) {
					engine1Wins++
				}
				if result == engine.Error {
					errors++
				}
				break
			}

			// Make a turn for engine2
			result = makeTurn(stdin2, stdout2, fen, &moveList)
			if result != engine.InProgress {
				fmt.Println("Game over:", result)
				if engine.IsDraw(result) {
					draws++
				}
				if engine.IsBlackWin(result) {
					engine2Wins++
				}
				if result == engine.Error {
					errors++
				}
				break
			}
		}

		// same board, but switched sides
		gameBoard = engine.InitFENBoard(fen)
		moveList = []string{}
		for {
			// Make a turn for engine2
			result := makeTurn(stdin2, stdout2, fen, &moveList)
			if result != engine.InProgress {
				if engine.IsDraw(result) {
					draws++
				}
				if engine.IsWhiteWin(result) {
					engine2Wins++
				}
				if result == engine.Error {
					errors++
				}
				break
			}

			// Make a turn for engine1
			result = makeTurn(stdin1, stdout1, fen, &moveList)
			if result != engine.InProgress {
				if engine.IsDraw(result) {
					draws++
				}
				if engine.IsBlackWin(result) {
					engine1Wins++
				}
				if result == engine.Error {
					errors++
				}
				break
			}
		}
	}

	return engine1Wins, engine2Wins, draws, errors
}

// makeTurn makes a turn for the engine and gets the gameresult after the turn
func makeTurn(stdin io.WriteCloser, stdout io.ReadCloser, fenString string, moveList *[]string) byte {
	var err error

	// Check the engine isready
	_, err = stdin.Write([]byte("isready\n"))
	if err != nil {
		fmt.Println("Error writing to engine:", err)
		return engine.Error
	}

	// Read the output
	buf := make([]byte, 1024)
	_, err = stdout.Read(buf)
	if err != nil {
		fmt.Println("Error reading from engine:", err)
		return engine.Error
	}

	positionString := "position fen " + fenString
	if len(*moveList) != 0 {
		positionString += " moves"
	}
	for _, move := range *moveList {
		positionString += " " + move
	}
	positionString += "\n"

	// Write the current state of the board to the engine
	_, err = stdin.Write([]byte(positionString))
	if err != nil {
		fmt.Println("Error writing to engine:", err)
		return engine.Error
	}

	// Get the move from the engine
	_, err = stdin.Write([]byte("go\n"))
	if err != nil {
		fmt.Println("Error writing to engine:", err)
		return engine.Error
	}

	// Read the output
	n, err := stdout.Read(buf)
	if err != nil {
		fmt.Println("Error reading from engine:", err)
		return engine.Error
	}
	bestmove := string(buf[9 : n-1])

	// Strip "bestmove " from the beginning of the string
	gameMove, ok := gameBoard.TryMoveUCI(bestmove)
	if !ok {
		fmt.Println("Invalid move from engine:", bestmove)
		return engine.Error
	} else {
		gameBoard.MakeMove(gameMove)
	}
	*moveList = append(*moveList, bestmove)

	// Return status of new board
	return engine.GetGameState(gameBoard)
}
