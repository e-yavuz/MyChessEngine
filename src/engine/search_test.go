package chessengine

import (
	"testing"
	"time"
)

func Test_SearchStartPosition_20s(t *testing.T) {
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	InitZobristTable()
	InitPeSTO()

	test := InitStartBoard()
	TTReset(test, DefaultTTMBSize)
	DebugMode = true
	cancelChannel := make(chan struct{})

	startTime := time.Now()
	go func() {
		time.Sleep(time.Duration(20000) * time.Millisecond)
		close(cancelChannel)
	}()
	test.StartSearch(startTime, cancelChannel)
	DebugMode = false
}

func Test_SearchPosition1_20s(t *testing.T) {
	InitMagicBitBoardTable("../../magic_rook", "../../magic_bishop")
	InitZobristTable()
	InitPeSTO()

	test := InitFENBoard("qrb5/rk1p1K2/p2P4/Pp6/1N2n3/6p1/5nB1/6b1 w - - 0 1")
	TTReset(test, DefaultTTMBSize)
	DebugMode = true
	cancelChannel := make(chan struct{})

	startTime := time.Now()
	go func() {
		time.Sleep(time.Duration(20000) * time.Millisecond)
		close(cancelChannel)
	}()
	test.StartSearch(startTime, cancelChannel)
	DebugMode = false
}
