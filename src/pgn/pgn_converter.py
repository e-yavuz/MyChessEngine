from typing import List
import chess.pgn

# This program converts a list of PGN games to a list of UCI moves

# Loads a list of PGN files and returns a list of games
def load_pgn_files(file_paths: List[str]):
    games = []
    for file_path in file_paths:
        with open(file_path) as file:
            while True:
                game = chess.pgn.read_game(file)
                if game is None:
                    break
                games.append(game)
    return games


# Converts a list of PGN games to a list of UCI moves
def convert_to_uci_moves(games: List[chess.pgn.Game]):
    uci_moves_list = []
    for game in games:
        uci_moves = []
        board = game.board()
        for move in game.mainline_moves():
            uci_moves.append(move.uci())
            board.push(move)
        uci_moves_list.append(uci_moves)
    return uci_moves_list

# Appends list of UCI moves to a file
def append_to_file(file_path: str, uci_moves_list: List[List[str]]):
    with open(file_path, 'a') as file:
        for uci_moves in uci_moves_list:
            file.write(' '.join(uci_moves) + '\n')