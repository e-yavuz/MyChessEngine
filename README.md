# Chess Engine Project

   |\_
  /  .\_
 |   ___)
 |  _ \
 /_____\
[_______]

## Overview
This project is a personal endeavor inspired by Sebastian Lague's chess engine video series. It aims to create a fully functional chess engine with various advanced features to enhance its playing strength and efficiency. The project is written in Go and includes several state-of-the-art techniques used in modern chess engines.

## Features

### 1. Quiescence Search
Quiescence Search is an extension of the minimax algorithm that focuses on evaluating "quiet" positions, where no immediate tactical threats exist. This helps to avoid the horizon effect, where the engine might miss important tactical sequences just beyond its search depth.

### 2. Null Move Pruning
Null Move Pruning is a technique used to quickly eliminate branches in the search tree that are unlikely to provide a better move. By making a "null move" (a pass move) and seeing if the opponent can significantly improve their position, the engine can decide to prune that branch early.

### 3. Late Move Reduction (LMR)
Late Move Reduction reduces the search depth for less promising moves later in the move list. This allows the engine to allocate more time to analyzing the most promising moves, improving overall search efficiency.

### 4. Principal Variation Search (PVS)
Principal Variation Search is an optimization of the alpha-beta pruning algorithm. It assumes that the first move considered in a position is the best move and searches it with a full window. Subsequent moves are searched with a null window, which can significantly reduce the number of nodes evaluated.

### 5. Move Ordering
Efficient move ordering is crucial for the performance of alpha-beta pruning. This engine uses several heuristics for move ordering, including:
- Most Valuable Victim - Least Valuable Aggressor (MVV-LVA)
- Killer Move Heuristic
- History Heuristic

### 6. Transposition Table
A transposition table is a cache that stores previously evaluated positions to avoid redundant calculations. This helps in speeding up the search process by reusing results from previously evaluated positions.

### 7. Magic Bitboards
Magic Bitboards are a technique for fast computation of sliding piece attacks (rooks, bishops, queens). They use precomputed tables to quickly determine the attack squares for these pieces.

### 8. Opening Book
The engine includes an opening book to play well-known opening moves quickly and efficiently. This helps in reaching the middle game with a strong position without spending much time on the clock.

### 9. Evaluation Function
The evaluation function assesses the strength of a position based on various factors such as material balance, piece activity, king safety, pawn structure, and more. This function is crucial for the engine to make informed decisions during the search.

## Project Structure
## Getting Started

### Prerequisites

- Go 1.16 or later

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/chessengine.git
    cd chessengine
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Build the project:
    ```sh
    go build -o chessengine main.go
    ```

## Running the Engine

To start the engine in UCI mode:
```sh
./chessengine
```

## Getting Started

### Prerequisites

- Go 1.16 or later
- A modern C++ compiler (for compiling magic bitboards)
- Visual Studio Code (recommended for development)

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/chessengine.git
    cd chessengine
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Build the project:
    ```sh
    go build -o chessengine main.go
    ```

## Running the Engine

To start the engine in UCI mode:
```sh
./chessengine
```

## Usage

### UCI Commands

The engine supports the Universal Chess Interface (UCI) protocol. Here are some common commands:

- [`uci`]: Initialize the engine and print engine info.
- [`isready`]: Check if the engine is ready.
- [`position [fen | startpos] moves ...`]: Set up the board position.
- [`go [depth | movetime | wtime | btime | movestogo]`]: Start calculating the best move.
- [`stop`]: Stop calculating.
- [`quit`]: Exit the engine.

### Example

To play a game against the engine using UCI commands:
```sh
uci
isready
position startpos
go depth 10
```

## Contributing

Contributions are welcome! Please fork the repository and create a pull request with your changes. Ensure that your code follows the existing style and includes tests for new features or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Acknowledgments

- Sebastian Lague for his inspiring chess engine video series.
- The open-source chess community for various resources and ideas.