package main

import (
	"ICD-10-project/elastic"
	"ICD-10-project/handlers"
	"ICD-10-project/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/search-by-description", middleware(http.HandlerFunc(handlers.SearchHandler)))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running on port http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func indexICD10Data(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	var wg sync.WaitGroup
	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}

		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			f, err := os.Open(dir + "/" + fileName)
			if err != nil {
				log.Printf("Error opening file: %v", err)
				return
			}
			defer f.Close()

			var icd10IndexRequest models.ICD10IndexRequest
			err = json.NewDecoder(f).Decode(&icd10IndexRequest)
			if err != nil {
				log.Printf("Error decoding JSON: %v", err)
				return
			}

			icd10IndexRequest.ICD10Code = icd10IndexRequest.Subcategory
			if icd10IndexRequest.Subcategory == "" {
				icd10IndexRequest.ICD10Code = icd10IndexRequest.CategoryCode
			}

			elastic.IndexateICD10Code(icd10IndexRequest)
		}(fileName)
	}
	wg.Wait()
}

func main() {
	// dir := "./icd10_data_symptoms"
	// indexICD10Data(dir)
	server()

}
