package chessengine

func Perft(board *Board, ply int, rootLevel bool) (retval uint64, rootNodes map[string]uint64) {
	if ply == 0 {
		return 1, nil
	}
	if rootLevel {
		rootNodes = make(map[string]uint64)
	}

	// Reset this entry in the moveList pool back to having 0 entries
	moveList := board.GenerateMoves(ALL, searchMovePool[ply][:0])
	if ply == 1 && !rootLevel {
		return uint64(len(moveList)), nil
	}

	for _, move := range moveList {
		board.MakeMove(move)
		leafCount, _ := Perft(board, ply-1, false)
		retval += leafCount
		if rootLevel {
			rootNodes[MoveToString(move)] += leafCount
		}
		board.UnMakeMove()
	}
	return retval, rootNodes
}
