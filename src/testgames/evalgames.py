import chess
import chess.engine

def get_fen_list():
    # Initialize a set
    opening_set = {""}
    fen_list = []
    
    with open("output.txt", "r") as fen_file:
        while True:
            opening = fen_file.readline()
            if not opening:
                break
            # Check if the opening is in the set already
            if opening in opening_set:
                continue
            fen = fen_file.readline()
            fen_list.append(fen)
    return fen_list

def evaluate_fen(fen_list: list[str], stockfish_path: str = "./stockfish") -> tuple[list[str], list[float]]:
    fen_retval = []
    eval_retval = []
    with chess.engine.SimpleEngine.popen_uci(stockfish_path) as engine:
        # Open a file of fen strings
        i = 0
        for fen in fen_list:
            board = chess.Board(fen)
            # Set up the board with the FEN
            info = engine.analyse(board, chess.engine.Limit(time=1))
            score = info["score"].relative.score(mate_score=10000)
            # See if the score is roughly equal
            if score is not None and abs(score) < 20:
                fen_retval.append(fen)
                eval_retval.append(score/100.0)
            if (i % int(len(fen_list)/100)) == 0:
                print(int(i/len(fen_list) * 100), "%")
            i += 1
                
                
    return fen_retval, eval_retval


stockfish_path = "./stockfish"  # Update with the correct path to your Stockfish binary

fen_list, eval_list = evaluate_fen(get_fen_list(), stockfish_path)
# write fen_list and eval_list to a file
with open("testgames_3.txt", "w") as file:
    for fen, eval in zip(fen_list, eval_list):
        file.write(fen)
        file.write(str(eval) + "\n")
