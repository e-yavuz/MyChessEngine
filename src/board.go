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

func (board *Board) DeepCopy() (retval Board) {
	retval = *board
	for i, stateInfo := range board.stateInfoArr {
		temp := *stateInfo
		retval.stateInfoArr[i] = &temp
	}
	for i, pieceInfo := range board.PieceInfoArr {
		if pieceInfo != nil {
			temp := *pieceInfo
			retval.PieceInfoArr[i] = &temp
		}
	}

	return retval
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
		retval.GetTopState().EnPassantPosition = INVALID_POSITION
	}

	retval.GetTopState().IsWhiteTurn = turnColor == "w"
	retval.GetTopState().HalfMoveClock, _ = strconv.Atoi(drawCount)
	retval.GetTopState().TurnCounter, _ = strconv.Atoi(turnCount)

	return retval
}

// Places single piece denoted by char (i.e. 'p' for black pawn) onto position with empty state
func (board *Board) placeFENonBoard(r rune, position Position) {
	thisPiece := NewPiece()
	switch r {
	case 'p':
		thisPiece.thisBitBoard = &board.B.Pawn
		thisPiece.isWhite = false
		thisPiece.pieceTYPE = PAWN
	case 'r':
		thisPiece.thisBitBoard = &board.B.Rook
		thisPiece.isWhite = false
		thisPiece.pieceTYPE = ROOK
	case 'n':
		thisPiece.thisBitBoard = &board.B.Knight
		thisPiece.isWhite = false
		thisPiece.pieceTYPE = KNIGHT
	case 'b':
		thisPiece.thisBitBoard = &board.B.Bishop
		thisPiece.isWhite = false
		thisPiece.pieceTYPE = BISHOP
	case 'q':
		thisPiece.thisBitBoard = &board.B.Queen
		thisPiece.isWhite = false
		thisPiece.pieceTYPE = QUEEN
	case 'k':
		thisPiece.thisBitBoard = &board.B.King
		thisPiece.isWhite = false
		thisPiece.pieceTYPE = KING
	case 'P':
		thisPiece.thisBitBoard = &board.W.Pawn
		thisPiece.isWhite = true
		thisPiece.pieceTYPE = PAWN
	case 'R':
		thisPiece.thisBitBoard = &board.W.Rook
		thisPiece.isWhite = true
		thisPiece.pieceTYPE = ROOK
	case 'N':
		thisPiece.thisBitBoard = &board.W.Knight
		thisPiece.isWhite = true
		thisPiece.pieceTYPE = KNIGHT
	case 'B':
		thisPiece.thisBitBoard = &board.W.Bishop
		thisPiece.isWhite = true
		thisPiece.pieceTYPE = BISHOP
	case 'Q':
		thisPiece.thisBitBoard = &board.W.Queen
		thisPiece.isWhite = true
		thisPiece.pieceTYPE = QUEEN
	case 'K':
		thisPiece.thisBitBoard = &board.W.King
		thisPiece.isWhite = true
		thisPiece.pieceTYPE = KING
	}
	PlaceOnBitBoard(thisPiece.thisBitBoard, position)
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

	for position := PopLSB(&Bpawn); position != INVALID_POSITION; position = PopLSB(&Bpawn) {
		boardRep[position/8][position%8] = 'p'
	}
	for position := PopLSB(&Brook); position != INVALID_POSITION; position = PopLSB(&Brook) {
		boardRep[position/8][position%8] = 'r'
	}
	for position := PopLSB(&Bknight); position != INVALID_POSITION; position = PopLSB(&Bknight) {
		boardRep[position/8][position%8] = 'n'
	}
	for position := PopLSB(&Bbishop); position != INVALID_POSITION; position = PopLSB(&Bbishop) {
		boardRep[position/8][position%8] = 'b'
	}
	for position := PopLSB(&Bqueen); position != INVALID_POSITION; position = PopLSB(&Bqueen) {
		boardRep[position/8][position%8] = 'q'
	}
	for position := PopLSB(&Bking); position != INVALID_POSITION; position = PopLSB(&Bking) {
		boardRep[position/8][position%8] = 'k'
	}
	for position := PopLSB(&Wpawn); position != INVALID_POSITION; position = PopLSB(&Wpawn) {
		boardRep[position/8][position%8] = 'P'
	}
	for position := PopLSB(&Wrook); position != INVALID_POSITION; position = PopLSB(&Wrook) {
		boardRep[position/8][position%8] = 'R'
	}
	for position := PopLSB(&Wknight); position != INVALID_POSITION; position = PopLSB(&Wknight) {
		boardRep[position/8][position%8] = 'N'
	}
	for position := PopLSB(&Wbishop); position != INVALID_POSITION; position = PopLSB(&Wbishop) {
		boardRep[position/8][position%8] = 'B'
	}
	for position := PopLSB(&Wqueen); position != INVALID_POSITION; position = PopLSB(&Wqueen) {
		boardRep[position/8][position%8] = 'Q'
	}
	for position := PopLSB(&Wking); position != INVALID_POSITION; position = PopLSB(&Wking) {
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
