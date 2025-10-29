package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// GameState represents the state of the game, now with variable board size.
type GameState struct {
	Board     []string `json:"board"`
	BoardSize int      `json:"boardSize"`
	Winner    string   `json:"winner"`
}

// Player and AI constants
const (
	PLAYER_X = "X"
	AI_O     = "O"
	EMPTY    = ""
)

func main() {
	mux := http.NewServeMux()

	// The /play handler runs all the new game logic.
	mux.HandleFunc("/play", playHandler)

	// Note: This server does not serve an HTML file.
	// It's *only* a backend API.
	// You'll need a separate frontend to send requests to it.

	log.Println("Starting Tic-Tac-Toe AI server on http://localhost:8080")
	log.Println("Send POST requests to /play")
	
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}

// playHandler handles the game logic for each move.
func playHandler(w http.ResponseWriter, r *http.Request) {
	// --- CORS Handling ---
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// --- Decode Request ---
	var currentState GameState
	if err := json.NewDecoder(r.Body).Decode(&currentState); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if currentState.BoardSize == 0 || len(currentState.Board) != (currentState.BoardSize*currentState.BoardSize) {
		http.Error(w, "Board size and board length do not match", http.StatusBadRequest)
		return
	}

	// --- Game Logic ---
	
	// 1. Check if Player (X) has already won
	if checkWinner(currentState.Board, PLAYER_X, currentState.BoardSize) {
		currentState.Winner = PLAYER_X
	} else if isBoardFull(currentState.Board) {
		currentState.Winner = "draw"
	} else {
		// 2. It's AI's turn. Find the best move.
		aiMove(&currentState)

		// 3. Check if AI (O) won
		if checkWinner(currentState.Board, AI_O, currentState.BoardSize) {
			currentState.Winner = AI_O
		} else if isBoardFull(currentState.Board) {
			currentState.Winner = "draw"
		}
	}

	// --- Encode Response ---
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&currentState); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// aiMove routes to the correct AI logic based on board size.
func aiMove(state *GameState) {
	var bestMove int

	if state.BoardSize == 3 {
		// Use "perfect" Minimax for 3x3
		bestMove = findBestMoveMinimax(state.Board)
	} else {
		// Use "strong" Heuristic for 4x4, 5x5, etc.
		bestMove = findBestMoveHeuristic(state.Board, state.BoardSize)
	}

	// Make the move
	if bestMove != -1 && state.Board[bestMove] == EMPTY {
		state.Board[bestMove] = AI_O
	}
}

// --- Dynamic N x N Win Checker ---
func checkWinner(board []string, player string, n int) bool {
	// Check rows
	for r := 0; r < n; r++ {
		match := true
		for c := 0; c < n; c++ {
			if board[r*n+c] != player {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	// Check columns
	for c := 0; c < n; c++ {
		match := true
		for r := 0; r < n; r++ {
			if board[r*n+c] != player {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	// Check diagonal (top-left to bottom-right)
	matchDiag1 := true
	for i := 0; i < n; i++ {
		if board[i*n+i] != player {
			matchDiag1 = false
			break
		}
	}
	if matchDiag1 {
		return true
	}

	// Check diagonal (top-right to bottom-left)
	matchDiag2 := true
	for i := 0; i < n; i++ {
		if board[i*n+(n-1-i)] != player {
			matchDiag2 = false
			break
		}
	}
	if matchDiag2 {
		return true
	}

	return false
}

// isBoardFull checks if there are any empty cells left.
func isBoardFull(board []string) bool {
	for _, cell := range board {
		if cell == EMPTY {
			return false
		}
	}
	return true
}

func getEmptySpots(board []string) []int {
	var emptySpots []int
	for i, cell := range board {
		if cell == EMPTY {
			emptySpots = append(emptySpots, i)
		}
	}
	return emptySpots
}

// --- AI: 4x4 & 5x5 Heuristic Logic ---
func findBestMoveHeuristic(board []string, n int) int {
	emptySpots := getEmptySpots(board)
	if len(emptySpots) == 0 {
		return -1
	}

	// 1. Check for AI win
	for _, move := range emptySpots {
		boardCopy := make([]string, len(board))
		copy(boardCopy, board)
		boardCopy[move] = AI_O
		if checkWinner(boardCopy, AI_O, n) {
			return move
		}
	}

	// 2. Check for player win (and block)
	for _, move := range emptySpots {
		boardCopy := make([]string, len(board))
		copy(boardCopy, board)
		boardCopy[move] = PLAYER_X
		if checkWinner(boardCopy, PLAYER_X, n) {
			return move
		}
	}

	// 3. Try to take the center (or near-center)
	centerIdx := (n * n) / 2
	if board[centerIdx] == EMPTY {
		return centerIdx
	}

	// 4. Try to take corners
	corners := []int{0, n - 1, n * (n - 1), n*n - 1}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(corners), func(i, j int) { corners[i], corners[j] = corners[j], corners[i] })
	for _, corner := range corners {
		if board[corner] == EMPTY {
			return corner
		}
	}

	// 5. Take any random available spot
	return emptySpots[rand.Intn(len(emptySpots))]
}

// --- AI: 3x3 Minimax (Unbeatable) Logic ---

// (This win checker is simplified and hardcoded for 3x3 for the Minimax)
func minimaxWinChecker(b []string, p string) bool {
	return (b[0] == p && b[1] == p && b[2] == p) ||
		(b[3] == p && b[4] == p && b[5] == p) ||
		(b[6] == p && b[7] == p && b[8] == p) ||
		(b[0] == p && b[3] == p && b[6] == p) ||
		(b[1] == p && b[4] == p && b[7] == p) ||
		(b[2] == p && b[5] == p && b[8] == p) ||
		(b[0] == p && b[4] == p && b[8] == p) ||
		(b[2] == p && b[4] == p && b[6] == p)
}

// findBestMoveMinimax is the entry point for the 3x3 AI
func findBestMoveMinimax(board []string) int {
	bestVal := -int(math.Inf(1))
	bestMove := -1

	for i := 0; i < 9; i++ {
		if board[i] == EMPTY {
			board[i] = AI_O // Make the move
			moveVal := minimax(board, 0, false)
			board[i] = EMPTY // Undo the move

			if moveVal > bestVal {
				bestMove = i
				bestVal = moveVal
			}
		}
	}
	return bestMove
}

// minimax is the core recursive function
func minimax(board []string, depth int, isMaximizing bool) int {
	// Check for terminal states
	if minimaxWinChecker(board, AI_O) {
		return 10 - depth
	}
	if minimaxWinChecker(board, PLAYER_X) {
		return depth - 10
	}
	if isBoardFull(board) {
		return 0
	}

	if isMaximizing {
		// AI's turn (maximize score)
		best := -int(math.Inf(1))
		for i := 0; i < 9; i++ {
			if board[i] == EMPTY {
				board[i] = AI_O
				best = max(best, minimax(board, depth+1, false))
				board[i] = EMPTY
			}
		}
		return best
	} 
	
	// Player's turn (minimize score)
	best := int(math.Inf(1))
	for i := 0; i < 9; i++ {
		if board[i] == EMPTY {
			board[i] = PLAYER_X
			best = min(best, minimax(board, depth+1, true))
			board[i] = EMPTY
		}
	}
	return best
}

// Helper functions for minimax
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
