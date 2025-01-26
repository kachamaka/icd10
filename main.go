package main

import (
	"ICD-10/handlers"
	"fmt"
	"log"
	"net/http"
)

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func server() {
	http.Handle("/", middleware(http.HandlerFunc(handlers.Home)))
	http.Handle("/search-by-description", middleware(http.HandlerFunc(handlers.SearchHandler)))

	fmt.Println("Server running on port http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	// dir := "./icd10_data_symptoms"
	// indexICD10Data(dir)
	server()

}
