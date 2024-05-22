package board

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

/*
	Contains
		all 12 pieace BitBoards and pieceInfoHashmap
		function to generate all psuedo legal moves
		function to generate all legal moves
		function make a move
		function to unmake a move
		function to convert FEN string to a board
*/

const StartingFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
const NullPosition = byte(255)

type Board struct {
	Bpawn   BitBoard
	Bknight BitBoard
	Brook   BitBoard
	Bbishop BitBoard
	Bqueen  BitBoard
	Bking   BitBoard

	Wpawn   BitBoard
	Wknight BitBoard
	Wrook   BitBoard
	Wbishop BitBoard
	Wqueen  BitBoard
	Wking   BitBoard

	EnPassantPosition byte
	IsWhiteTurn       bool

	CastleWKing  bool
	CastleBKing  bool
	CastleWQueen bool
	CastleBQueen bool

	DrawCounter int
	TurnCounter int

	PieceInfoMap map[byte]*Piece
}

// TODO
func GeneratePsuedoLegalMoves() {
	return
}

// TODO
func GenerateLegalMoves() {
	return
}

// Invariant: Assumes move is possible to make!
func (board *Board) MakeMove(move *Move) {
	piece := board.PieceInfoMap[move.GetStartingPosition()]

	//Start by removing this piece from the bitBoard it was originally, it is essentially in limbo now
	piece.ThisBitBoard.removeFromBitBoard(move.GetStartingPosition())

	//Now need to check flags to see what is happening!
	switch move.GetFlag() {
	case quietFlag: //Default case
		break
	case doublePawnPushFlag:
		board.EnPassantPosition = (move.GetStartingPosition() + move.GetTargetPosition()) / 2
	case kingCastleFlag:
		if piece.IsWhite {
			board.CastleWKing = false
			board.CastleWQueen = false
			board.swapPiecePositions(&board.Wrook, move.GetTargetPosition()+1, move.GetTargetPosition()-1)
		} else {
			board.CastleBKing = false
			board.CastleBQueen = false
			board.swapPiecePositions(&board.Brook, move.GetTargetPosition()+1, move.GetTargetPosition()-1)
		}
	case queenCastleFlag:
		if piece.IsWhite {
			board.CastleWKing = false
			board.CastleWQueen = false
			board.swapPiecePositions(&board.Wrook, move.GetTargetPosition()-1, move.GetTargetPosition()+1)
		} else {
			board.CastleBKing = false
			board.CastleBQueen = false
			board.swapPiecePositions(&board.Brook, move.GetTargetPosition()-1, move.GetTargetPosition()+1)
		}
	case captureFlag:
		

	}

}

func (board *Board) swapPiecePositions(bitboard *BitBoard, startPosition, targetPosition byte) {
	bitboard.removeFromBitBoard(startPosition)
	bitboard.placeOnBitBoard(targetPosition)
	board.PieceInfoMap[targetPosition] = board.PieceInfoMap[startPosition]
	delete(board.PieceInfoMap, startPosition)
}

// TODO
func (board *Board) UnMakeMove(move *Move) {
	return
}

// Starts with FEN "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
func InitStartBoard() *Board {
	return InitFENBoard(StartingFen)
}

// TODO add in handling for enpassant and castling status
// Sets up board with all empty state given a FEN string
func InitFENBoard(FEN string) *Board {
	FEN_Arr := strings.Split(FEN, " ")
	piecePositions := FEN_Arr[0]
	turnColor := FEN_Arr[1]
	castlingRights := FEN_Arr[2]
	enPassantSquare := FEN_Arr[3]
	drawCount := FEN_Arr[4]
	turnCount := FEN_Arr[5]

	retval := &Board{
		PieceInfoMap: make(map[byte]*Piece),
		IsWhiteTurn:  turnColor == "w",
	}

	position := byte(56)

	for _, r := range piecePositions {
		if r == '/' {
			position -= 16
		} else if unicode.IsLetter(r) {
			retval.placeFENonBoard(r, position)
			position += 1
		} else if unicode.IsDigit(r) {
			position += byte(r - '0')
		}
	}

	for _, r := range castlingRights {
		switch r {
		case 'K':
			retval.CastleWKing = true
		case 'Q':
			retval.CastleWQueen = true
		case 'k':
			retval.CastleBKing = true
		case 'q':
			retval.CastleBQueen = true
		}

	}

	if enPassantSquare != "-" {
		var row byte = enPassantSquare[0] - 'a'
		var col byte = enPassantSquare[1] - '0'
		retval.EnPassantPosition = (row*8 + col)
	} else {
		retval.EnPassantPosition = NullPosition
	}

	retval.DrawCounter, _ = strconv.Atoi(drawCount)
	retval.TurnCounter, _ = strconv.Atoi(turnCount)

	return retval
}

// Places single piece denoted by char (i.e. 'p' for black pawn) onto position with empty state
func (board *Board) placeFENonBoard(r rune, position byte) {
	thisPiece := NewPiece()
	switch r {
	case 'p':
		thisPiece.ThisBitBoard = &board.Bpawn
		thisPiece.IsWhite = false
	case 'r':
		thisPiece.ThisBitBoard = &board.Brook
		thisPiece.IsWhite = false
	case 'n':
		thisPiece.ThisBitBoard = &board.Bknight
		thisPiece.IsWhite = false
	case 'b':
		thisPiece.ThisBitBoard = &board.Bbishop
		thisPiece.IsWhite = false
	case 'q':
		thisPiece.ThisBitBoard = &board.Bqueen
		thisPiece.IsWhite = false
	case 'k':
		thisPiece.ThisBitBoard = &board.Bking
		thisPiece.IsWhite = false
	case 'P':
		thisPiece.ThisBitBoard = &board.Wpawn
		thisPiece.IsWhite = true
	case 'R':
		thisPiece.ThisBitBoard = &board.Wrook
		thisPiece.IsWhite = true
	case 'N':
		thisPiece.ThisBitBoard = &board.Wknight
		thisPiece.IsWhite = true
	case 'B':
		thisPiece.ThisBitBoard = &board.Wbishop
		thisPiece.IsWhite = true
	case 'Q':
		thisPiece.ThisBitBoard = &board.Wqueen
		thisPiece.IsWhite = true
	case 'K':
		thisPiece.ThisBitBoard = &board.Wking
		thisPiece.IsWhite = true
	}
	thisPiece.ThisBitBoard.placeOnBitBoard(position)
	board.PieceInfoMap[position] = thisPiece
}

// TODO
func InitPGNBoard(PGN string) *Board {
	return &Board{}
}

// Displays 8x8 board in cmdline
func (board *Board) DisplayBoard() {
	var boardRep [8][8]rune
	for _, position := range board.Bpawn.LSBpositions() {
		boardRep[position/8][position%8] = 'p'
	}
	for _, position := range board.Brook.LSBpositions() {
		boardRep[position/8][position%8] = 'r'
	}
	for _, position := range board.Bknight.LSBpositions() {
		boardRep[position/8][position%8] = 'n'
	}
	for _, position := range board.Bbishop.LSBpositions() {
		boardRep[position/8][position%8] = 'b'
	}
	for _, position := range board.Bqueen.LSBpositions() {
		boardRep[position/8][position%8] = 'q'
	}
	for _, position := range board.Bking.LSBpositions() {
		boardRep[position/8][position%8] = 'k'
	}
	for _, position := range board.Wpawn.LSBpositions() {
		boardRep[position/8][position%8] = 'P'
	}
	for _, position := range board.Wrook.LSBpositions() {
		boardRep[position/8][position%8] = 'R'
	}
	for _, position := range board.Wknight.LSBpositions() {
		boardRep[position/8][position%8] = 'N'
	}
	for _, position := range board.Wbishop.LSBpositions() {
		boardRep[position/8][position%8] = 'B'
	}
	for _, position := range board.Wqueen.LSBpositions() {
		boardRep[position/8][position%8] = 'Q'
	}
	for _, position := range board.Wking.LSBpositions() {
		boardRep[position/8][position%8] = 'K'
	}
	fmt.Println("-----------------")
	for row := 7; row >= 0; row-- {
		fmt.Print("|")
		str := ""
		for col := 0; col < 8; col++ {
			if boardRep[row][col] != 0 {
				str += string(boardRep[row][col])
			} else {
				str += "#"
			}
		}
		fmt.Printf("%s|\n", strings.Join(strings.Split(str, ""), " "))
	}
	fmt.Println("-----------------")
}
