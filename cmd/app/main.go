package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"server/internal/db"
	"server/internal/handlers"
)

func main() {
	// Initialize the database
	db.InitDB()
	defer func() {
		sqlDB, err := db.DB.DB()
		if err != nil {
			log.Printf("Failed to close database connection: %v", err)
			return
		}
		sqlDB.Close()
	}()

	// Configure and start the router
	router := mux.NewRouter()
	router.HandleFunc("/register", handlers.LogRequest(handlers.RegisterClient)).Methods("POST")
	router.HandleFunc("/clients", handlers.LogRequest(handlers.ListClients)).Methods("GET")
	router.HandleFunc("/client/{id}/execute", handlers.LogRequest(handlers.ExecCommand)).Methods("POST")
	router.HandleFunc("/client/{id}/command", handlers.LogRequest(handlers.SaveCommand)).Methods("POST")
	router.HandleFunc("/commands", handlers.LogRequest(handlers.ListCommands)).Methods("GET")

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
