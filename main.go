package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// GameState represents the state of the tic-tac-toe game.
type GameState struct {
	Board  []string `json:"board"`
	Winner string   `json:"winner"`
}

func main() {
	mux := http.NewServeMux()

	// --- NEW FIX: Directly serve the HTML file ---
	// This handler is simpler and avoids potential version issues with 'embed'.
	// It requires the server to be run from the project's root directory.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If a user requests anything other than the main page, show a "not found" error.
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// Explicitly serve the index.html file from its known path.
		http.ServeFile(w, r, "api/frontend/index.html")
	})

	// This handler runs the game logic and remains unchanged.
	mux.HandleFunc("/play", playHandler)

	log.Println("Server starting on http://localhost:8080")
	log.Println("IMPORTANT: Run this from your project's main folder (the one containing the 'api' directory).")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}


// playHandler handles incoming requests from the frontend.
func playHandler(w http.ResponseWriter, r *http.Request) {
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

	var currentState GameState
	if err := json.NewDecoder(r.Body).Decode(&currentState); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if checkWinner(currentState.Board, "X") {
		currentState.Winner = "X"
	} else if isBoardFull(currentState.Board) {
		currentState.Winner = "draw"
	} else {
		aiMove(&currentState.Board)
		if checkWinner(currentState.Board, "O") {
			currentState.Winner = "O"
		} else if isBoardFull(currentState.Board) {
			currentState.Winner = "draw"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&currentState); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// aiMove finds the best possible move for the AI ("O").
func aiMove(board *[]string) {
	// 1. Check for a winning move for "O"
	for i := 0; i < 9; i++ {
		if (*board)[i] == "" {
			(*board)[i] = "O"
			if checkWinner(*board, "O") {
				return
			}
			(*board)[i] = ""
		}
	}

	// 2. Check for a blocking move against "X"
	for i := 0; i < 9; i++ {
		if (*board)[i] == "" {
			(*board)[i] = "X"
			if checkWinner(*board, "X") {
				(*board)[i] = "O"
				return
			}
			(*board)[i] = ""
		}
	}

	// 3. Take center if available
	if (*board)[4] == "" {
		(*board)[4] = "O"
		return
	}

	// 4. Take a random corner if available
	corners := []int{0, 2, 6, 8}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(corners), func(i, j int) { corners[i], corners[j] = corners[j], corners[i] })
	for _, corner := range corners {
		if (*board)[corner] == "" {
			(*board)[corner] = "O"
			return
		}
	}

	// 5. Take any remaining empty cell
	for i := 0; i < 9; i++ {
		if (*board)[i] == "" {
			(*board)[i] = "O"
			return
		}
	}
}

// checkWinner determines if a player has won.
func checkWinner(board []string, player string) bool {
	winConditions := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // Rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // Columns
		{0, 4, 8}, {2, 4, 6}, // Diagonals
	}

	for _, condition := range winConditions {
		if board[condition[0]] == player && board[condition[1]] == player && board[condition[2]] == player {
			return true
		}
	}
	return false
}

// isBoardFull checks if there are any empty cells left.
func isBoardFull(board []string) bool {
	for _, cell := range board {
		if cell == "" {
			return false
		}
	}
	return true
}

