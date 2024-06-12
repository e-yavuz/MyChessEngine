package main

import (
	testgames "chessengine/src/testgames"
	uci "chessengine/src/uci"
	"fmt"
)

func main() {
	// runTest("./bots/v5", "./bots/v6", "src/testgames/testgames_highquality.txt", 150, 200)
	makeUCI()
}

func runTest(engine1, engine2, testgamesPath string, startindex, numgames int) {
	fmt.Println(testgames.StartGameFile(engine1, engine2, testgamesPath, startindex, numgames))
}

func makeUCI() {
	uci.UCI()
}
