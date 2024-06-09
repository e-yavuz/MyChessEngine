package main

import (
	testgames "chessengine/src/testgames"
	uci "chessengine/src/uci"
	"fmt"
)

func main() {
	// runTest("./bots/v1", "./bots/v4", "src/testgames/testgames_2.txt", 0, 10)
	makeUCI()
}

func runTest(engine1, engine2, testgamesPath string, startindex, numgames int) {
	fmt.Println(testgames.StartGameFile(engine1, engine2, testgamesPath, startindex, numgames))
}

func makeUCI() {
	uci.UCI()
}
