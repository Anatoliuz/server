// main.go

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Client struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
}

func main() {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	))
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()

	// Configure and start the router
	router := mux.NewRouter()
	router.HandleFunc("/register", logRequest(registerClient)).Methods("POST")
	router.HandleFunc("/clients", logRequest(listClients)).Methods("GET")
	router.HandleFunc("/client/{id}/exec", logRequest(execCommand)).Methods("POST")

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Middleware function for logging incoming requests
func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.RequestURI)
		handler.ServeHTTP(w, r)
		log.Printf("Completed %s in %v", r.RequestURI, time.Since(start))
	}
}

func registerClient(w http.ResponseWriter, r *http.Request) {
	clientAddr := r.RemoteAddr
	log.Printf("Registering client from address: %s", clientAddr)

	_, err := db.Exec("INSERT INTO clients (address) VALUES ($1)", clientAddr)
	if err != nil {
		log.Printf("Error registering client: %v", err)
		http.Error(w, "Error registering client", http.StatusInternalServerError)
		return
	}

	log.Printf("Client registered successfully from address: %s", clientAddr)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Client registered"))
}

func listClients(w http.ResponseWriter, r *http.Request) {
	log.Println("Fetching list of clients...")

	rows, err := db.Query("SELECT id, address FROM clients")
	if err != nil {
		log.Printf("Error fetching clients: %v", err)
		http.Error(w, "Error fetching clients", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	clients := []Client{}
	for rows.Next() {
		var client Client
		if err := rows.Scan(&client.ID, &client.Address); err != nil {
			log.Printf("Error scanning client data: %v", err)
			http.Error(w, "Error processing clients", http.StatusInternalServerError)
			return
		}
		clients = append(clients, client)
	}

	log.Printf("Found %d clients", len(clients))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func execCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["id"]

	// Parse JSON request body to get the command
	var requestBody struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Fetch the client's address from the database
	var address string
	err := db.QueryRow("SELECT address FROM clients WHERE id = $1", clientID).Scan(&address)
	if err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Prepare the JSON payload for the remote server
	payload := map[string]string{"command": requestBody.Command}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to create command payload", http.StatusInternalServerError)
		return
	}

	// Send the command to the client's address
	clientURL := fmt.Sprintf("http://%s/execute", address)
	req, err := http.NewRequest("POST", clientURL, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		http.Error(w, "Failed to create request to client", http.StatusInternalServerError)
		return
	}

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send command to client", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Client command execution failed", resp.StatusCode)
		return
	}

	fmt.Fprintf(w, "Command sent successfully to client at %s\n", address)
}