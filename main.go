package main

import (
	"chessengine/src/board"
	"fmt"
)

func main() {
	b := board.InitStartBoard()
	b.DisplayBoard()
	fmt.Printf("Castling rights: BK %t, BQ %t, WK %t, WQ %t\nEnPassant Position: %d\nWhite Turn: %t, Turn Count: %d, Draw Count: %d",
		b.CastleBKing, b.CastleBQueen, b.CastleWKing, b.CastleWQueen, b.EnPassantPosition, b.IsWhiteTurn, b.TurnCounter, b.DrawCounter)
}
