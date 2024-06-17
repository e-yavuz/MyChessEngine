package chessengine

import (
	"bufio"
	engine "chessengine/src/engine"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var gameBoard *engine.Board

func StartGameFile(engine1 string, engine2 string, fenFile string, startIndex, numGames int) (int, int, int, int) {
	// Open the file
	file, err := os.Open(fenFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return -1, -1, -1, -1
	}
	defer file.Close()

	// Create a new Scanner for the file
	scanner := bufio.NewScanner(file)

	// Read the file
	fenBoards := []string{}
	i := 0
	for scanner.Scan() && i/2 < (startIndex+numGames) {
		// Read each line of the file
		line := scanner.Text()

		// Add every even line to fenBoards (odd lines are evaluations)
		if i%2 == 0 && (i/2) >= startIndex {
			fenBoards = append(fenBoards, line)
		}

		i++
	}
	return StartGame(engine1, engine2, fenBoards)
}

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
	stderr1, _ := e1.StderrPipe()
	stdout2, _ := e2.StdoutPipe()
	stdin2, _ := e2.StdinPipe()
	stderr2, _ := e2.StderrPipe()

	// Start the engines
	err = e1.Start()
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr1)
		}
		fmt.Println("Error starting engine 1:", err)
		return -1, -1, -1, -1
	}
	err = e2.Start()
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr2)
		}
		fmt.Println("Error starting engine 2:", err)
		return -1, -1, -1, -1
	}

	// Wait for the engines to start
	_, err = stdin1.Write([]byte("uci\n"))
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr1)
		}
		fmt.Println("Error writing to engine 1:", err)
		return -1, -1, -1, -1
	}

	n, err := stdout1.Read(buf)
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr1)
		}
		fmt.Println("Error reading from engine 1:", err)
		return -1, -1, -1, -1
	}
	fmt.Println("Engine 1:", string(buf[:n]))

	_, err = stdin2.Write([]byte("uci\n"))
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr2)
		}
		fmt.Println("Error writing to engine 2:", err)
		return -1, -1, -1, -1
	}

	n, err = stdout2.Read(buf)
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr2)
		}
		fmt.Println("Error reading from engine 2:", err)
		return -1, -1, -1, -1
	}
	fmt.Println("Engine2:", string(buf[:n]))

	// e1 is white
	for _, fen := range fenBoards {
		// Start the game
		_, err = stdin1.Write([]byte("ucinewgame\n"))
		if err != nil {
			if err == io.EOF {
				handleEOFError(stderr1)
			}
			fmt.Println("Error writing to engine 1:", err)
			return -1, -1, -1, -1
		}
		_, err = stdin2.Write([]byte("ucinewgame\n"))
		if err != nil {
			if err == io.EOF {
				handleEOFError(stderr2)
			}
			fmt.Println("Error writing to engine 2:", err)
			return -1, -1, -1, -1
		}
		gameBoard = engine.InitFENBoard(fen)
		moveList = []string{}
		for {
			// Make a turn for engine1
			result := makeTurn(1, stdin1, stdout1, stderr1, fen, &moveList)
			if result != engine.InProgress {
				if engine.IsDraw(result) {
					draws++
					fmt.Println("Game over by:", engine.GameResultToString(result))
				}
				if engine.IsWhiteWin(result) {
					engine1Wins++
					fmt.Println("Game over by: Engine 1 Win")
				}
				if result == engine.Error {
					errors++
					fmt.Printf("Game over by: Error\nFen: %s\nMoves: %v\n", fen, moveList)
				}
				break
			}

			// Make a turn for engine2
			result = makeTurn(2, stdin2, stdout2, stderr2, fen, &moveList)
			if result != engine.InProgress {
				if engine.IsDraw(result) {
					draws++
					fmt.Println("Game over by:", engine.GameResultToString(result))
				}
				if engine.IsBlackWin(result) {
					engine2Wins++
					fmt.Println("Game over by: Engine 2 Win")
				}
				if result == engine.Error {
					errors++
					fmt.Printf("Game over by: Error\nFen: %s\nMoves: %v\n", fen, moveList)
				}
				break
			}
		}

		// Start the game
		_, err = stdin1.Write([]byte("ucinewgame\n"))
		if err != nil {
			if err == io.EOF {
				handleEOFError(stderr1)
			}
			fmt.Println("Error writing to engine 1:", err)
			return -1, -1, -1, -1
		}
		_, err = stdin2.Write([]byte("ucinewgame\n"))
		if err != nil {
			if err == io.EOF {
				handleEOFError(stderr2)
			}
			fmt.Println("Error writing to engine 2:", err)
			return -1, -1, -1, -1
		}

		// same board, but switched sides
		gameBoard = engine.InitFENBoard(fen)
		moveList = []string{}
		for {
			// Make a turn for engine2
			result := makeTurn(2, stdin2, stdout2, stderr2, fen, &moveList)
			if result != engine.InProgress {
				if engine.IsDraw(result) {
					draws++
					fmt.Println("Game over by:", engine.GameResultToString(result))
				}
				if engine.IsWhiteWin(result) {
					engine2Wins++
					fmt.Println("Game over by: Engine 2 Win")
				}
				if result == engine.Error {
					errors++
					fmt.Printf("Game over by: Error\nFen: %s\nMoves: %v\n", fen, moveList)
				}
				break
			}

			// Make a turn for engine1
			result = makeTurn(1, stdin1, stdout1, stderr1, fen, &moveList)
			if result != engine.InProgress {
				if engine.IsDraw(result) {
					draws++
					fmt.Println("Game over by:", engine.GameResultToString(result))
				}
				if engine.IsBlackWin(result) {
					engine1Wins++
					fmt.Println("Game over by: Engine 1 Win")
				}
				if result == engine.Error {
					errors++
					fmt.Printf("Game over by: Error\nFen: %s\nMoves: %v\n", fen, moveList)
				}
				break
			}
		}
	}

	return engine1Wins, engine2Wins, draws, errors
}

// makeTurn makes a turn for the engine and gets the gameresult after the turn
func makeTurn(engineID int, stdin io.WriteCloser, stdout, stderr io.ReadCloser, fenString string, moveList *[]string) byte {
	var err error

	// Check the engine isready
	_, err = stdin.Write([]byte("isready\n"))
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr)
		}
		fmt.Printf("Error writing to engine %d: %s\n", engineID, err)
		return engine.Error
	}

	// Read the output
	buf := make([]byte, 1024)
	_, err = stdout.Read(buf)
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr)
		}
		fmt.Printf("Error reading from engine %d: %s\n", engineID, err)
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
		fmt.Printf("Error writing to engine %d: %s\n", engineID, err)
		return engine.Error
	}

	// Get the move from the engine
	_, err = stdin.Write([]byte("go\n"))
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr)
		}
		fmt.Printf("Error writing to engine %d: %s\n", engineID, err)
		return engine.Error
	}

	// Read the output
	n, err := stdout.Read(buf)
	if err != nil {
		if err == io.EOF {
			handleEOFError(stderr)
		}
		fmt.Printf("Error reading from engine %d: %s\n", engineID, err)
		return engine.Error
	}
	bestmove := string(buf[9 : n-1])

	// Strip "bestmove " from the beginning of the string
	gameMove, ok := gameBoard.TryMoveUCI(bestmove)
	if !ok {
		fmt.Printf("Invalid move from engine %d: %s\n", engineID, bestmove)
		return engine.Error
	} else {
		gameBoard.MakeMove(gameMove)
	}
	*moveList = append(*moveList, bestmove)

	// Return status of new board
	return engine.GetGameState(gameBoard)
}

func handleEOFError(stdErr io.ReadCloser) {
	buf := make([]byte, 1024)
	n, err := stdErr.Read(buf)
	if err != nil {
		fmt.Println("Error reading from engine:", err)
	}
	fmt.Println("Error:", string(buf[:n]))
}
