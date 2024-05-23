package main

import (
	chessengine "chessengine/src"
	"fmt"
)

func main() {
	b := chessengine.InitStartBoard()
	fmt.Print(b.DisplayBoard())
}
