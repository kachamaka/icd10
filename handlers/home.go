package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Welcome to the Home Page"}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error occurred while marshaling response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
