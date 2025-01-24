package main

import (
	"ICD-10-project/elastic"
	"ICD-10-project/handlers"
	"ICD-10-project/models"
	"ICD-10-project/util"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

// func home(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Welcome to the ICD-10 search engine!\n")
// }

func server() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	// http.HandleFunc("/", home)
	// http.HandleFunc("/createIndex", createIndexHandler)
	http.HandleFunc("/search", handlers.SearchHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running on port http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func writeDataToCSV() {
	csvFile, err := os.Create("./result.csv")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer csvFile.Close()

	csvFile.WriteString("type,typeCode,title\n")

	files, _ := os.ReadDir("./icd10_data")
	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}
		// open file
		f, err := os.Open(fmt.Sprintf("./icd10_data/%s", fileName))
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer f.Close()

		// decode json
		var data map[string]interface{}
		err = json.NewDecoder(f).Decode(&data)
		if err != nil {
			log.Fatalf("Error decoding JSON: %v", err)
		}

		// write data values with keys the headers
		typeName := data["type"].(string)
		categoryCode := data["categoryCode"].(string)
		subcategory := data["subcategory"].(string)
		title := data["title"].(string)
		title = fmt.Sprintf("\"%v\"", title)

		if subcategory == "" {
			subcategory = categoryCode
		}

		csvFile.WriteString(fmt.Sprintf("%s,%s,%s\n", typeName, subcategory, title))
	}
}
func indexICD10Data() {
	dir := "./icd10_data_symptoms"
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

func queryElastic(query string) []models.ICD10SearchResponse {
	res, err := elastic.Search(query)
	if err != nil {
		log.Fatalf("Error executing search: %v", err)
	}

	return res
}

func addSymptomsData() {
	data, err := util.LoadICD10Data()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(len(data))

	file, err := os.Open("symptoms.csv")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}
	records = records[1:]
	// DISEASES,SYMPTOM2,SYMPTOM3,SYMPTOM4,SYMPTOM5,SYMPTOM6,SYMPTOM7,SYMPTOM8,SYMPTOM9,SYMPTOM10,SYMPTOM11,SYMPTOM12,SYMPTOM13,SYMPTOM14,SYMPTOM15,SYMPTOM16,SYMPTOM17

	for _, record := range records {
		// fmt.Println(strings.Trim())
		title := strings.Trim(record[0], " ")
		var symptoms []string
		for i := 1; i < len(record); i++ {
			symptom := strings.Trim(record[i], " ")
			if symptom != "" {
				symptoms = append(symptoms, symptom)
			}
		}
		fmt.Println("Title:", title)
		fmt.Println("Symptoms:", symptoms)
		icd10res, err := elastic.Search(title)
		if err != nil {
			log.Printf("Error searching ES: %v", err)
		}
		if len(icd10res) == 0 {
			continue
		}

		topResult := icd10res[0]
		// fmt.Println("Top result:", topResult.ICD10Code, "Title:", topResult.Title)

		icd10Entry := data[topResult.ICD10Code]
		icd10Entry.Symptoms = symptoms
		data[topResult.ICD10Code] = icd10Entry
		// fmt.Println(data[topResult.ICD10Code], "MODIFIED DATA")
	}

	for _, value := range data {
		fileName := fmt.Sprintf("./icd10_data_symptoms/icd10_%s_%s_%s", value.ChapterCode, value.BlockCode, value.CategoryCode)
		if value.Subcategory != "" {
			fileName += fmt.Sprintf("_%s", value.Subcategory)
		}
		fileName += ".json"

		file, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("Error creating file: %v", err)
		}
		defer file.Close()

		prettyJSON, err := json.MarshalIndent(value, "", "    ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}

		_, err = file.Write(prettyJSON)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	}
}

func prettyPrint() {
	files, err := os.ReadDir("./icd10_data_symptoms")
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}

		f, err := os.Open("./icd10_data_symptoms/" + fileName)
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer f.Close()

		var data models.ICD10IndexRequest
		err = json.NewDecoder(f).Decode(&data)
		if err != nil {
			log.Fatalf("Error decoding JSON: %v", err)
		}

		data.ICD10Code = data.Subcategory
		if data.Subcategory == "" {
			data.ICD10Code = data.CategoryCode
		}

		for i, symptom := range data.Symptoms {
			data.Symptoms[i] = strings.ReplaceAll(symptom, "_", " ")
		}

		prettyJSON, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}

		err = os.WriteFile("./icd10_data_symptoms/"+fileName, prettyJSON, 0644)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	}
}

func main() {
	// go server()

	// indexICD10Data()
	server()

	// prettyPrint()

	// time.Sleep(time.Hour)
}
