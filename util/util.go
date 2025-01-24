package util

import (
	"ICD-10-project/models"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

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

// func printHtml(n *html.Node) {
// 	if n == nil {
// 		return
// 	}
// 	var b bytes.Buffer
// 	_ = html.Render(&b, n)
// 	fmt.Println("data: ", b.String())
// }
