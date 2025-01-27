package util

import (
	"ICD-10/models"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Call the Python script to process the text
func ProcessTextWithPython(symptomText string) (string, error) {
	// Prepare the command to run the Python script

	cmd := exec.Command("python3", "./util/process_text_nltk.py", symptomText)

	// Capture the output of the Python script
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	// Return the processed text or an error
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

// func indexICD10Data(dir string) {
// 	files, err := os.ReadDir(dir)
// 	if err != nil {
// 		log.Fatalf("Error reading directory: %v", err)
// 	}

// 	var wg sync.WaitGroup
// 	for _, file := range files {
// 		fileName := file.Name()
// 		if !strings.HasSuffix(fileName, ".json") {
// 			continue
// 		}

// 		wg.Add(1)
// 		go func(fileName string) {
// 			defer wg.Done()
// 			f, err := os.Open(dir + "/" + fileName)
// 			if err != nil {
// 				log.Printf("Error opening file: %v", err)
// 				return
// 			}
// 			defer f.Close()

// 			var icd10IndexRequest models.ICD10IndexRequest
// 			err = json.NewDecoder(f).Decode(&icd10IndexRequest)
// 			if err != nil {
// 				log.Printf("Error decoding JSON: %v", err)
// 				return
// 			}

// 			icd10IndexRequest.ICD10Code = icd10IndexRequest.Subcategory
// 			if icd10IndexRequest.Subcategory == "" {
// 				icd10IndexRequest.ICD10Code = icd10IndexRequest.CategoryCode
// 			}

// 			elastic.IndexateICD10Code(icd10IndexRequest)
// 		}(fileName)
// 	}
// 	wg.Wait()
// }

// func LoadICD10Data() (map[string]models.ICD10IndexRequest, error) {
// 	dataMap := make(map[string]models.ICD10IndexRequest)

// 	files, err := os.ReadDir("./icd10_data")
// 	if err != nil {
// 		log.Printf("Error reading directory: %v", err)
// 		return nil, err
// 	}

// 	for _, file := range files {
// 		fileName := file.Name()
// 		if !strings.HasSuffix(fileName, ".json") {
// 			continue
// 		}

// 		f, err := os.Open(fmt.Sprintf("./icd10_data/%s", fileName))
// 		if err != nil {
// 			log.Printf("Error opening file: %v", err)
// 			return nil, err
// 		}
// 		defer f.Close()

// 		var icd10IndexRequest models.ICD10IndexRequest
// 		err = json.NewDecoder(f).Decode(&icd10IndexRequest)
// 		if err != nil {
// 			log.Printf("Error decoding JSON: %v", err)
// 			return nil, err
// 		}

// 		icd10IndexRequest.ICD10Code = icd10IndexRequest.Subcategory
// 		if icd10IndexRequest.Subcategory == "" {
// 			icd10IndexRequest.ICD10Code = icd10IndexRequest.CategoryCode
// 		}

// 		dataMap[icd10IndexRequest.ICD10Code] = icd10IndexRequest
// 	}

// 	return dataMap, nil
// }

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
