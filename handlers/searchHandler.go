package handlers

import (
	"ICD-10-project/elastic"
	"ICD-10-project/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req models.ICD10SearchQuery
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Error parsing JSON: ", err)
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	icd10Codes, err := elastic.Search(req.Query)
	if err != nil {
		log.Println("Error searching: ", err)
		http.Error(w, "Error searching", http.StatusInternalServerError)
		return
	}

	res := models.SearchResponse{
		ICD10Codes: icd10Codes,
	}
	// send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
