import chess
import chess.pgn

def pgn_to_long_algebraic(fen, pgn_move):
    # Initialize the board with the FEN string
    board = chess.Board(fen)
    
    # Parse the PGN move to a chess.Move object
    move = board.parse_san(pgn_move)
    
    # Get the long algebraic notation
    long_algebraic_move = move.uci()
    
    return long_algebraic_move

# 1k1r4/pp1b1R2/3q2pp/4p3/2B5/4Q3/PPP2B2/2K5 b - - bm Qd1+; id "BK.01";
def readBKFile(file_name):
    retval = []
    with open(file_name) as file:
        newline = file.readline()
        while newline != "":
            fen = newline[:newline.index("bm")]
            pgnMoves = newline[newline.index("bm")+3:newline.index(";")]
            pgnMoves = pgnMoves.split(" ")
            bestMoves = []
            for move in pgnMoves:
                bestMoves.append(pgn_to_long_algebraic(fen, move))
            id = newline[newline.index("\"")+1:newline.index("\"", newline.index("\"")+1)]
            moveListstring = " ".join(bestMoves)
            retval.append(";".join([fen, moveListstring, id]))
            newline = file.readline()
    return retval

def writeBKMoves(file_name, bk_read_out):
    with open(file_name, 'w') as file:
        for test_info in bk_read_out:
            file.write(test_info + '\n')
        
        

out = readBKFile("/Applications/VSCODE/Projects/Personal/WorkInProgress/ChessEngine/MyEngine/src/testpositions/testpositions_in.txt")
writeBKMoves("/Applications/VSCODE/Projects/Personal/WorkInProgress/ChessEngine/MyEngine/src/testpositions/testpositions_out.txt", out)