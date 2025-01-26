package util

import (
	"ICD-10/elastic"
	"ICD-10/models"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

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

func DOIDToICD10() {
	// Input and output file paths
	inputFile := "./data/allXREFinDO.tsv"
	outputFile := "./data/icd10ToDO2.csv"

	// Open the TSV file for reading
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		return
	}
	defer file.Close()

	// Create the CSV file for writing
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write the CSV header
	err = writer.Write([]string{"icd10", "doid"})
	if err != nil {
		fmt.Printf("Error writing header: %v\n", err)
		return
	}

	// Read the TSV file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Split the line into columns by tab
		columns := strings.Split(line, "\t")
		if len(columns) < 3 {
			fmt.Printf("Skipping malformed line: %s\n", line)
			continue
		}

		// Extract DOID and ICD10CM fields
		doid := strings.Trim(columns[0], "\"")
		xrefs := strings.Trim(columns[2], "\"")

		// the icd10cm part can have multiple ICD-10 codes separated by a comma and space
		icd10cmCodes := strings.Split(xrefs, ", ")
		for _, icd10cm := range icd10cmCodes {
			icd10cm = strings.TrimPrefix(icd10cm, "ICD10CM:")
			doid = strings.TrimPrefix(doid, "DOID:")

			// Write the row to the CSV
			err = writer.Write([]string{icd10cm, doid})
			if err != nil {
				fmt.Printf("Error writing to CSV: %v\n", err)
				return
			}
		}
	}

	// Handle any errors during file reading
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		return
	}

	fmt.Println("Conversion completed successfully! Check the output.csv file.")
}

func LoadICD10Data() (map[string]models.ICD10IndexRequest, error) {
	dataMap := make(map[string]models.ICD10IndexRequest)

	files, err := os.ReadDir("./icd10_data")
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return nil, err
	}

	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}

		f, err := os.Open(fmt.Sprintf("./icd10_data/%s", fileName))
		if err != nil {
			log.Printf("Error opening file: %v", err)
			return nil, err
		}
		defer f.Close()

		var icd10IndexRequest models.ICD10IndexRequest
		err = json.NewDecoder(f).Decode(&icd10IndexRequest)
		if err != nil {
			log.Printf("Error decoding JSON: %v", err)
			return nil, err
		}

		icd10IndexRequest.ICD10Code = icd10IndexRequest.Subcategory
		if icd10IndexRequest.Subcategory == "" {
			icd10IndexRequest.ICD10Code = icd10IndexRequest.CategoryCode
		}

		dataMap[icd10IndexRequest.ICD10Code] = icd10IndexRequest
	}

	return dataMap, nil
}

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
