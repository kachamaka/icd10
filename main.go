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

func queryElastic(query string) []models.ICD10SearchResponse {
	res, err := elastic.Search(query)
	if err != nil {
		log.Fatalf("Error executing search: %v", err)
	}

	return res
}

func extractSymptoms() {
	file, err := os.Open("./data/symptomsData.csv")
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

	symptoms := make(map[string][]string)

	for _, record := range records {
		disease := strings.Trim(record[0], " ")
		for i := 1; i < len(record); i++ {
			symptom := strings.Trim(record[i], " ")
			if symptom != "" {
				symptom = strings.ReplaceAll(symptom, "_", " ")
				if !util.Contains(symptoms[disease], symptom) {
					symptoms[disease] = append(symptoms[disease], symptom)
				}
			}
		}
	}

	csvFile, err := os.Create("./data/diseasesSymptoms.csv")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer csvFile.Close()

	csvFile.WriteString("disease,symptoms\n")

	for disease, symptomsList := range symptoms {
		symptomsStr := strings.Join(symptomsList, ", ")
		csvFile.WriteString(fmt.Sprintf("%s,\"%s\"\n", disease, symptomsStr))
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

func extractSymptomsDOID() {
	file, err := os.Open("./data/icd10ToDO.csv")
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

	icd10ToDO := make(map[string]string)
	DOToICD10 := make(map[string]string)
	for _, record := range records {
		icd10ToDO[record[0]] = record[1]
		DOToICD10[record[1]] = record[0]
	}

	file, err = os.Open("./data/disease-symptom.csv")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	reader = csv.NewReader(file)
	records, err = reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}
	records = records[1:] // Skip header

	// changed := map[string]models.ICD10IndexRequest{}
	symptoms := make(map[string][]string)
	for _, record := range records {
		doid := strings.TrimPrefix(record[0], "DOID:")
		symptom := record[3]
		icd10 := DOToICD10[doid]
		if icd10 != "" {
			if !util.Contains(symptoms[icd10], symptom) {
				symptoms[icd10] = append(symptoms[icd10], symptom)
			}
			// entry, ok := changed[icd10]
			// if !ok {
			// 	entry = data[icd10]
			// }

			// if !util.Contains(entry.Symptoms, symptom) {
			// 	entry.Symptoms = append(entry.Symptoms, symptom)
			// }

			// changed[icd10] = entry
			// fmt.Println("ICD10:", icd10, "Symptoms:", ("\"" + strings.Join(entry.Symptoms, "\",\"") + "\""))
			// } else {
			// 	fmt.Println("No data for:", doid, icd10)
		}
	}

	csvFile, err := os.Create("./data/symptomsDOID.csv")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer csvFile.Close()

	csvFile.WriteString("icd10,symptoms\n")

	for icd10, symptomsList := range symptoms {
		symptomsStr := strings.Join(symptomsList, ", ")
		csvFile.WriteString(fmt.Sprintf("%s,\"%s\"\n", icd10, symptomsStr))
	}

	// for diseases := range symptoms
	// fmt.Println(len(symptoms))

}

func addSymptomsDataDO() {
	data, err := util.LoadICD10Data()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(len(data))

	file, err := os.Open("./data/icd10ToDO.csv")
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

	icd10ToDO := make(map[string]string)
	DOToICD10 := make(map[string]string)
	for _, record := range records {
		icd10ToDO[record[0]] = record[1]
		DOToICD10[record[1]] = record[0]
	}

	file, err = os.Open("./data/disease-symptom.csv")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	reader = csv.NewReader(file)
	records, err = reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}
	records = records[1:] // Skip header

	changed := map[string]models.ICD10IndexRequest{}
	for _, record := range records {
		doid := strings.TrimPrefix(record[0], "DOID:")
		symptom := record[3]
		icd10 := DOToICD10[doid]
		if icd10 != "" {
			entry, ok := changed[icd10]
			if !ok {
				entry = data[icd10]
			}

			if !util.Contains(entry.Symptoms, symptom) {
				entry.Symptoms = append(entry.Symptoms, symptom)
			}

			changed[icd10] = entry
			// fmt.Println("ICD10:", icd10, "Symptoms:", ("\"" + strings.Join(entry.Symptoms, "\",\"") + "\""))
			// } else {
			// 	fmt.Println("No data for:", doid, icd10)
		}
	}

	saveData(changed, "icd10_data_changed")
}

func saveData(data map[string]models.ICD10IndexRequest, dir string) {
	for _, value := range data {
		fileName := fmt.Sprintf("./"+dir+"/icd10_%s_%s_%s", value.ChapterCode, value.BlockCode, value.CategoryCode)
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

func main() {
	// dir := "./icd10_data_symptoms"
	// indexICD10Data(dir)
	// server()
	extractSymptomsDOID()

	// prettyPrint()
	// addSymptomsData()
	// addSymptomsDataDO()

	// time.Sleep(time.Hour)
}
