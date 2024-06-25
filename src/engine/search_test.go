package chessengine

import (
	"testing"
	"time"
)

func Test_SearchStartPosition_3s(t *testing.T) {
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	InitZobristTable()
	InitPeSTO()

	test := InitStartBoard()
	cancelChannel := make(chan int)

	startTime := time.Now()
	go func() {
		time.Sleep(time.Duration(3000) * time.Millisecond)
		close(cancelChannel)
	}()
	test.StartSearch(startTime, cancelChannel)
}
