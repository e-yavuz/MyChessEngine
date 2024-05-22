package main

import (
	chessengine "chessengine/src"
)

func main() {
	b := chessengine.InitStartBoard()
	b.DisplayBoard()
}
