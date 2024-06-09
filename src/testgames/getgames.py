import chess.pgn  # Thanks to: github.com/niklasf/python-chess/
import random
import math

pgn = open("lichess_db_standard_rated_2014-10.pgn")
positions = []

for i in range(15000):
    game = chess.pgn.read_game(pgn)
    if game == None:
        break
    moves = game.mainline_moves()
    board = game.board()
    
    # Random (even) move count
    plyToPlay = math.floor(16 + 20 * random.random()) & ~1
    # Play moves on board
    numPlyPlayed = 0
    fen = ""
    for move in moves:
        board.push(move)
        numPlyPlayed += 1
        if numPlyPlayed == plyToPlay:
            fen = board.fen()
            break
        
    # Get length of moves
    moves = list(moves)
    
    
    # Record position (if game continued for 20+ moves, and 10+ pieces remain)
    numPiecesInPos = sum(fen.lower().count(char) for char in "rnbq")
    if len(moves) >= plyToPlay * 2 and numPiecesInPos >= 10:
        positions.append(game.headers['Opening'])
        positions.append(fen)
    
    if i % 1500 == 0:
        print(i/15000.0 * 100, "%")

# Write to file
with open("output.txt", "w") as file:
    for string in positions:
        file.write(string + "\n")
