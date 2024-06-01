package chessengine

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestInitTable(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	InitMagicNumber(1, 2, "temptest")
	InitMagicBitBoardTable("temptest_rook", "temptest_bishop")

	if RookMask(4) != 0x1010101010106E {
		t.Fatalf("Failed possible move generation\n\tGot:\n%s\n\tWanted:\n%s",
			BitBoardToString(RookMask(4)), BitBoardToString(0x1010101010106E))
	}

	if GetRookMoves(4, 32) != 1157442765409226799 {
		t.Fatalf("Failed magic array pull\n\tGot:\n%s\n\tWanted:\n%s",
			BitBoardToString(GetRookMoves(4, 32)), BitBoardToString(1157442765409226799))
	}

	if BishopMask(12) != 0x244280000 {
		t.Fatalf("Failed magic array pull\n\tGot:\n%s\n\tWanted:\n%s",
			BitBoardToString(BishopMask(12)), BitBoardToString(0x244280000))
	}

	if GetBishopMoves(12, 0x280028) != 550832177192 {
		t.Fatalf("Failed magic array pull\n\tGot:\n%s\n\tWanted:\n%s",
			BitBoardToString(GetBishopMoves(12, 0x280028)), BitBoardToString(1157442765409226799))
	}

	// Remove temporary test files
	os.Remove("temptest_rook")
	os.Remove("temptest_bishop")
}
