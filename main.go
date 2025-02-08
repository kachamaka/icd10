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
	port := 8080

	http.Handle("/", middleware(http.HandlerFunc(handlers.Home)))
	http.Handle("/api/search-by-description", middleware(http.HandlerFunc(handlers.SearchByDescriptionHandler)))
	http.Handle("/api/search-by-symptoms", middleware(http.HandlerFunc(handlers.SearchBySymptomsHandler)))

	fmt.Printf("Server running on port https://localhost:%v\n", port)

	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	// err := http.ListenAndServeTLS(fmt.Sprintf(":%v", port), "./cert/cert.pem", "./cert/key.pem", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func main() {
	// dir := "./icd10_data_symptoms"
	// util.IndexICD10Data(dir)
	server()

	// symptomText := "I've been feeling really tired for the past few days, and today I woke up with a headache and a sore throat. Yesterday, I had a mild fever that lasted for a few hours, but it went away in the evening. I also have a persistent cough and some difficulty breathing, especially at night. I’ve noticed a bit of swelling around my eyes and my legs feel a little weak. My stomach feels upset, and I’ve had some nausea along with a bit of diarrhea. I haven’t been able to sleep well because of these issues, and I feel dehydrated with a dry mouth."

	// processedWords, err := util.ProcessTextWithPython(symptomText)
	// if err != nil {
	// 	log.Fatalf("Error processing text with Python: %s", err)
	// }

	// // fmt.Println(processedWords)

	// elastic.SearchBySymptoms(processedWords)
}
