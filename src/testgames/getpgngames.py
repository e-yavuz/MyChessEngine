import chess.pgn  # Thanks to: github.com/niklasf/python-chess/
import random
import math

pgn = open("lichess_db_standard_rated_2014-10.pgn")
positions = []

numGames = 1000

with chess.engine.SimpleEngine.popen_uci("./stockfish") as engine:
    opening_set = {""}
    for i in range(numGames):
        if i % (numGames/10) == 0:
            print(i/numGames * 100, "%")
        game = chess.pgn.read_game(pgn)
        if game == None:
            break
        moves = game.mainline_moves()
        board = game.board()
        if game.headers["Opening"] in opening_set:
            continue
        opening_set.add(game.headers["Opening"])
        
        
        
        # Random (even) move count
        plyToPlay = math.floor(8 + (2 * random.random())) & ~1
        # Play moves on board
        numPlyPlayed = 0
        thisGamePGN = ""
        node = None
        newGame = chess.pgn.Game()
        for move in moves:
            board.push(move)
            if node is None:
                node = newGame.add_variation(move)
            else:
                node = node.add_variation(move)
            numPlyPlayed += 1
            if numPlyPlayed == plyToPlay:
                # Record the pgn string of the board
                # print("check")
                newGame.setup(board)
                newGame.headers = game.headers
                newGame.headers["Result"] = "*"
                thisGamePGN = str(newGame)
                break
        
        # Get length of moves
        moves = list(moves)
        
        info = engine.analyse(board, chess.engine.Limit(time=0.1))
        score = info["score"].relative.score(mate_score=10000)
        if len(moves) >= plyToPlay * 2 and score < 25:
            positions.append(thisGamePGN)

# Write to file
with open("pgn_output.txt", "w") as file:
    for string in positions:
        file.write(string + "\n\n")
