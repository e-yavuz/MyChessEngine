package main

import (
	testgames "chessengine/src/testgames"
	uci "chessengine/src/uci"
	"fmt"
)

func main() {
	// runTest("./bots/v12d", "./bots/v12c", "src/testgames/testgames_highquality.txt", 0, 50)
	makeUCI()
}

func runTest(engine1, engine2, testgamesPath string, startindex, numgames int) {
	fmt.Println(testgames.StartGameFile(engine1, engine2, testgamesPath, startindex, numgames))
}

func makeUCI() {
	uci.UCI()
}
