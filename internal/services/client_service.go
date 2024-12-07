package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"

	"server/internal/db"
)

type RegisterRequest struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

type CommandRequest struct {
	Command string `json:"command"`
}

// RegisterClient registers a new client in the database.
func RegisterClient(request RegisterRequest) error {
	if request.IP == "" || request.Port == "" {
		return errors.New("IP and port fields are required")
	}

	clientAddr := fmt.Sprintf("%s:%s", request.IP, request.Port)

	// Create a new client using GORM
	client := db.Client{
		Address: clientAddr,
	}
	if err := db.DB.Create(&client).Error; err != nil {
		return fmt.Errorf("failed to register client: %w", err)
	}

	return nil
}

// GetClients retrieves all clients from the database.
func GetClients() ([]db.Client, error) {
	var clients []db.Client

	// Fetch all clients using GORM
	if err := db.DB.Find(&clients).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve clients: %w", err)
	}

	return clients, nil
}

func ExecCommand(w http.ResponseWriter, r *http.Request) {
	// Extract client ID from the URL
	vars := mux.Vars(r)
	clientID, err := strconv.Atoi(vars["id"]) // Convert client ID to integer
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Parse JSON request body to get the command
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
	clientURL := fmt.Sprintf("http://%s/execute", client.Address)
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

func EnqueueCommand(clientID uint, command string) (uint, error) {
	if command == "" {
		return 0, errors.New("command cannot be empty")
	}

	newCommand := db.Command{
		ClientID: clientID,
		Command:  command,
		Status:   "queued", // Команда по умолчанию сохраняется в статусе "queued"
	}

	// Сохраняем команду в базе данных
	result := db.DB.Create(&newCommand)
	if result.Error != nil {
		return 0, result.Error
	}

	return newCommand.ID, nil
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
	commandID, err := EnqueueCommand(uint(clientID), req.Command)
	if err != nil {
		http.Error(w, "Failed to save command", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ с ID команды
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Command queued with ID: %d for client ID: %d", commandID, clientID)
}
