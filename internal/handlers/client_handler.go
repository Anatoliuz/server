package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"net"
	"net/http"
	"server/internal/db"
	"strconv"

	"github.com/gorilla/mux"
	"server/internal/services"
)

type CommandRequest struct {
	Command string `json:"command"`
}

func RegisterClient(w http.ResponseWriter, r *http.Request) {
	var request services.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := services.RegisterClient(request); err != nil {
		http.Error(w, "Error registering client", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Client registered"))
}

func ListClients(w http.ResponseWriter, r *http.Request) {
	clients, err := services.GetClients()
	if err != nil {
		http.Error(w, "Error fetching clients", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func ExecCommand(w http.ResponseWriter, r *http.Request) {
	// Extract client ID from the URL
	vars := mux.Vars(r)
	clientIDStr := vars["id"]

	// Convert clientID from string to uint
	clientID, err := strconv.ParseUint(clientIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Parse JSON request body
	var requestBody struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Fetch the client's address from the database using GORM
	var client db.Client
	if err := db.DB.First(&client, uint(clientID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Client not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch client", http.StatusInternalServerError)
		}
		return
	}

	// Prepare the JSON payload for the remote server
	payload := map[string]string{"command": requestBody.Command}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to create command payload", http.StatusInternalServerError)
		return
	}

	// Use the client's address to construct the client URL
	clientURL := fmt.Sprintf("http://%s/e", client.Address)
	req, err := http.NewRequest("POST", clientURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, "Failed to create request to client", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
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

	// Respond with success
	fmt.Fprintf(w, "Command sent successfully to client at %s\n", client.Address)
}

func SaveCommand(w http.ResponseWriter, r *http.Request) {
	// Получаем ID клиента из пути запроса
	vars := mux.Vars(r)
	clientID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Сохраняем команду через сервис
	commandID, err := services.EnqueueCommand(uint(clientID), req.Command)
	if err != nil {
		http.Error(w, "Failed to save command", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ с ID команды
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Command queued with ID: %d for client ID: %d", commandID, clientID)
}

// ListCommands fetches all commands for the client based on their IP address
func ListCommands(w http.ResponseWriter, r *http.Request) {
	// Extract the client's IP address from the request
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Unable to determine client IP", http.StatusBadRequest)
		return
	}

	// Fetch the client from the database by IP
	var client db.Client
	if err := db.DB.Where("address LIKE ?", clientIP+"%").First(&client).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Client not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch client", http.StatusInternalServerError)
		}
		return
	}

	// Fetch commands associated with the client
	var commands []db.Command
	if err := db.DB.Where("client_id = ?", client.ID).Find(&commands).Error; err != nil {
		http.Error(w, "Failed to fetch commands", http.StatusInternalServerError)
		return
	}

	// Respond with the list of commands
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}
