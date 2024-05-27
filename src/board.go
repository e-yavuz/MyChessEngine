package chessengine

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

type Board struct {
	W Pieces // White Pieces
	B Pieces // Black Pieces

	stateInfoArr []*StateInfo

	PieceInfoArr [64]*PieceInfo
}

func (board *Board) Equal(other *Board) bool {
	bitBoardsCompare := board.B.Pawn == other.B.Pawn &&
		board.B.Knight == other.B.Knight &&
		board.B.Rook == other.B.Rook &&
		board.B.Bishop == other.B.Bishop &&
		board.B.Queen == other.B.Queen &&
		board.B.King == other.B.King &&
		board.W.Pawn == other.W.Pawn &&
		board.W.Knight == other.W.Knight &&
		board.W.Rook == other.W.Rook &&
		board.W.Bishop == other.W.Bishop &&
		board.W.Queen == other.W.Queen &&
		board.W.King == other.W.King

	stCompare := board.GetTopState().Equal(other.GetTopState())

	PieceInfoArrCompare := true

	for i := 0; i < 64; i++ {
		if board.PieceInfoArr[i] != nil && other.PieceInfoArr[i] != nil {
			PieceInfoArrCompare = board.PieceInfoArr[i].Equal(other.PieceInfoArr[i]) &&
				PieceInfoArrCompare
		} else {
			PieceInfoArrCompare = board.PieceInfoArr[i] == nil &&
				other.PieceInfoArr[i] == nil && PieceInfoArrCompare
		}

		if !PieceInfoArrCompare {
			return false
		}

	}

	return bitBoardsCompare && stCompare
}

func (board *Board) GetTopState() *StateInfo {
	return board.stateInfoArr[len(board.stateInfoArr)-1]
}

func (board *Board) PushNewState(newState *StateInfo) {
	board.stateInfoArr = append(board.stateInfoArr, newState)
}

func (board *Board) PopTopState() *StateInfo {
	if len(board.stateInfoArr) == 0 {
		panic("State Info array is already empty!")
	}
	retval := board.GetTopState()
	// Pop the top
	board.stateInfoArr = board.stateInfoArr[:len(board.stateInfoArr)-1]
	return retval
}

// Starts with FEN "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
func InitStartBoard() *Board {
	return InitFENBoard(StartingFen)
}

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
		stateInfoArr: []*StateInfo{{}},
	}

	var position Position = 56

	for _, r := range piecePositions {
		if r == '/' {
			position -= 16
		} else if unicode.IsLetter(r) {
			retval.placeFENonBoard(r, position)
			position += 1
		} else if unicode.IsDigit(r) {
			position += Position(r - '0')
		}
	}

	for _, r := range castlingRights {
		switch r {
		case 'K':
			retval.GetTopState().CastleWKing = true
		case 'Q':
			retval.GetTopState().CastleWQueen = true
		case 'k':
			retval.GetTopState().CastleBKing = true
		case 'q':
			retval.GetTopState().CastleBQueen = true
		}

	}

	if enPassantSquare != "-" {
		var col byte = enPassantSquare[0] - 'a'
		var row byte = enPassantSquare[1] - '1'
		retval.GetTopState().EnPassantPosition = Position(row*8 + col)
	} else {
		retval.GetTopState().EnPassantPosition = NULL_POSITION
	}

	retval.GetTopState().IsWhiteTurn = turnColor == "w"
	retval.GetTopState().DrawCounter, _ = strconv.Atoi(drawCount)
	retval.GetTopState().TurnCounter, _ = strconv.Atoi(turnCount)

	return retval
}

// Places single piece denoted by char (i.e. 'p' for black pawn) onto position with empty state
func (board *Board) placeFENonBoard(r rune, position Position) {
	thisPiece := NewPiece()
	switch r {
	case 'p':
		thisPiece.ThisBitBoard = &board.B.Pawn
		thisPiece.IsWhite = false
	case 'r':
		thisPiece.ThisBitBoard = &board.B.Rook
		thisPiece.IsWhite = false
	case 'n':
		thisPiece.ThisBitBoard = &board.B.Knight
		thisPiece.IsWhite = false
	case 'b':
		thisPiece.ThisBitBoard = &board.B.Bishop
		thisPiece.IsWhite = false
	case 'q':
		thisPiece.ThisBitBoard = &board.B.Queen
		thisPiece.IsWhite = false
	case 'k':
		thisPiece.ThisBitBoard = &board.B.King
		thisPiece.IsWhite = false
	case 'P':
		thisPiece.ThisBitBoard = &board.W.Pawn
		thisPiece.IsWhite = true
	case 'R':
		thisPiece.ThisBitBoard = &board.W.Rook
		thisPiece.IsWhite = true
	case 'N':
		thisPiece.ThisBitBoard = &board.W.Knight
		thisPiece.IsWhite = true
	case 'B':
		thisPiece.ThisBitBoard = &board.W.Bishop
		thisPiece.IsWhite = true
	case 'Q':
		thisPiece.ThisBitBoard = &board.W.Queen
		thisPiece.IsWhite = true
	case 'K':
		thisPiece.ThisBitBoard = &board.W.King
		thisPiece.IsWhite = true
	}
	PlaceOnBitBoard(thisPiece.ThisBitBoard, position)
	board.PieceInfoArr[position] = thisPiece
}

// TODOlow InitPGNBoard
func InitPGNBoard(PGN string) *Board {
	return &Board{}
}

// Displays 8x8 board in cmdline
func (board *Board) DisplayBoard() (retval string) {
	var boardRep [8][8]rune
	Bpawn := board.B.Pawn
	Brook := board.B.Rook
	Bknight := board.B.Knight
	Bbishop := board.B.Bishop
	Bqueen := board.B.Queen
	Bking := board.B.King
	Wpawn := board.W.Pawn
	Wrook := board.W.Rook
	Wknight := board.W.Knight
	Wbishop := board.W.Bishop
	Wqueen := board.W.Queen
	Wking := board.W.King

	for position := PopLSB(&Bpawn); position != NULL_POSITION; position = PopLSB(&Bpawn) {
		boardRep[position/8][position%8] = 'p'
	}
	for position := PopLSB(&Brook); position != NULL_POSITION; position = PopLSB(&Brook) {
		boardRep[position/8][position%8] = 'r'
	}
	for position := PopLSB(&Bknight); position != NULL_POSITION; position = PopLSB(&Bknight) {
		boardRep[position/8][position%8] = 'n'
	}
	for position := PopLSB(&Bbishop); position != NULL_POSITION; position = PopLSB(&Bbishop) {
		boardRep[position/8][position%8] = 'b'
	}
	for position := PopLSB(&Bqueen); position != NULL_POSITION; position = PopLSB(&Bqueen) {
		boardRep[position/8][position%8] = 'q'
	}
	for position := PopLSB(&Bking); position != NULL_POSITION; position = PopLSB(&Bking) {
		boardRep[position/8][position%8] = 'k'
	}
	for position := PopLSB(&Wpawn); position != NULL_POSITION; position = PopLSB(&Wpawn) {
		boardRep[position/8][position%8] = 'P'
	}
	for position := PopLSB(&Wrook); position != NULL_POSITION; position = PopLSB(&Wrook) {
		boardRep[position/8][position%8] = 'R'
	}
	for position := PopLSB(&Wknight); position != NULL_POSITION; position = PopLSB(&Wknight) {
		boardRep[position/8][position%8] = 'N'
	}
	for position := PopLSB(&Wbishop); position != NULL_POSITION; position = PopLSB(&Wbishop) {
		boardRep[position/8][position%8] = 'B'
	}
	for position := PopLSB(&Wqueen); position != NULL_POSITION; position = PopLSB(&Wqueen) {
		boardRep[position/8][position%8] = 'Q'
	}
	for position := PopLSB(&Wking); position != NULL_POSITION; position = PopLSB(&Wking) {
		boardRep[position/8][position%8] = 'K'
	}
	retval += "-----------------" + "\n"
	for row := 7; row >= 0; row-- {
		retval += "|"
		str := ""
		for col := 0; col < 8; col++ {
			if boardRep[row][col] != 0 {
				str += string(boardRep[row][col])
			} else {
				str += "#"
			}
		}
		retval += fmt.Sprintf("%s|\n", strings.Join(strings.Split(str, ""), " "))
	}
	retval += "-----------------"

	return retval
}
