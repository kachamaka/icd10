package render_data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const ICD10_WHO_BASIC_TOKEN = "Basic YTNlMTM2N2EtZmRiYi00OTM2LTgwMjItZjI3MzM0YjY2ZTg5XzU4YTAzMGRhLTJmYmMtNDlkYi1iMTgwLTgyYjVkN2FkM2E2MTpUTTBrU3Jsd05FbmFSVUhqUDNRNU1XdU5SZlA5MzhhMFJJUHlsR2J2S2ZjPQ=="

var ICD10_CHAPTERS = []string{
	"http://id.who.int/icd/release/10/2019/I",
	"http://id.who.int/icd/release/10/2019/II",
	"http://id.who.int/icd/release/10/2019/III",
	"http://id.who.int/icd/release/10/2019/IV",
	"http://id.who.int/icd/release/10/2019/V",
	"http://id.who.int/icd/release/10/2019/VI",
	"http://id.who.int/icd/release/10/2019/VII",
	"http://id.who.int/icd/release/10/2019/VIII",
	"http://id.who.int/icd/release/10/2019/IX",
	"http://id.who.int/icd/release/10/2019/X",
	"http://id.who.int/icd/release/10/2019/XI",
	"http://id.who.int/icd/release/10/2019/XII",
	"http://id.who.int/icd/release/10/2019/XIII",
	"http://id.who.int/icd/release/10/2019/XIV",
	"http://id.who.int/icd/release/10/2019/XV",
	"http://id.who.int/icd/release/10/2019/XVI",
	"http://id.who.int/icd/release/10/2019/XVII",
	"http://id.who.int/icd/release/10/2019/XVIII",
	"http://id.who.int/icd/release/10/2019/XIX",
	"http://id.who.int/icd/release/10/2019/XX",
	"http://id.who.int/icd/release/10/2019/XXI",
	"http://id.who.int/icd/release/10/2019/XXII",
}

type Token struct {
	AccessToken   string `json:"access_token"`
	ExpiresInUnix int64  `json:"expires_in"`
	TokenType     string `json:"token_type"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func getOrGenerateToken() func() string {
	var token Token
	return func() string {
		if token.AccessToken != "" && token.ExpiresInUnix < time.Now().Unix() {
			return token.TokenType + " " + token.AccessToken
		}

		client := &http.Client{}
		req, err := http.NewRequest("POST", "https://icdaccessmanagement.who.int/connect/token", bytes.NewBufferString("grant_type=client_credentials&scope=icdapi_access"))
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}

		req.Header.Set("Authorization", ICD10_WHO_BASIC_TOKEN)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Access-Control-Allow-Origin", "https://id.who.int")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
		}
		defer resp.Body.Close()

		var tokenResponse TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
			log.Fatalf("Error decoding response: %v", err)
		}

		token.AccessToken = tokenResponse.AccessToken
		token.ExpiresInUnix = time.Now().Unix() + int64(tokenResponse.ExpiresIn)
		token.TokenType = tokenResponse.TokenType

		return token.TokenType + " " + token.AccessToken
	}
}

func getToken() string {
	return getOrGenerateToken()()
}

func makeRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("API-Version", "v2")
	req.Header.Set("Authorization", getToken())

	return req, nil
}

type Title struct {
	Value string `json:"@value"`
}

type InEx struct {
	Label Title `json:"label"`
}

type ICD10Response struct {
	Code      string   `json:"code"`
	Children  []string `json:"child"`
	Title     Title    `json:"title"`
	Inclusion []InEx   `json:"inclusion"`
	Exclusion []InEx   `json:"exclusion"`
	ClassKind string   `json:"classKind"`
	Parent    []string `json:"parent"`
}

type ICD10Data struct {
	Type         string `json:"type"`
	ChapterCode  string `json:"chapterCode"`
	Chapter      string `json:"chapter"`
	BlockCode    string `json:"blockCode"`
	Block        string `json:"block"`
	CategoryCode string `json:"categoryCode"`
	Category     string `json:"category"`
	Subcategory  string `json:"subcategory"`

	Title     string   `json:"title"`
	Inclusion []string `json:"inclusion"`
	Exclusion []string `json:"exclusion"`
	Children  []string `json:"children"`
}

func updateData(data *ICD10Data, body ICD10Response) {
	if body.ClassKind == "category" && strings.Contains(body.Code, ".") {
		body.ClassKind = "subcategory"
	}
	data.Type = body.ClassKind

	switch body.ClassKind {
	case "chapter":
		data.ChapterCode = body.Code
		data.Chapter = body.Title.Value
	case "block":
		data.BlockCode = body.Code
		data.Block = body.Title.Value
	case "category":
		data.CategoryCode = body.Code
		data.Category = body.Title.Value
	case "subcategory":
		data.Subcategory = body.Code
		data.Title = body.Title.Value
	}

	for _, incl := range body.Inclusion {
		data.Inclusion = append(data.Inclusion, incl.Label.Value)
	}

	for _, excl := range body.Exclusion {
		data.Exclusion = append(data.Exclusion, excl.Label.Value)
	}

}

func traverseEndpointWrapper(url string, result ICD10Data, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	traverseEndpoint(url, result, wg, mu)

}

func traverseEndpoint(url string, result ICD10Data, wg *sync.WaitGroup, mu *sync.Mutex) {
	client := &http.Client{}

	req, err := makeRequest(url)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	var body ICD10Response
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		log.Fatalf("Error decoding response body: %v", err)
	}

	updateData(&result, body)

	if len(body.Children) == 0 {
		result.Title = body.Title.Value
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}

		filePath := fmt.Sprintf("./icd10_data/icd10_%s_%s_%s_%s.json", result.ChapterCode, result.BlockCode, result.CategoryCode, result.Subcategory)
		if result.Subcategory == "" {
			filePath = fmt.Sprintf("./icd10_data/icd10_%s_%s_%s.json", result.ChapterCode, result.BlockCode, result.CategoryCode)
		}

		if _, err := os.Stat(filePath); err == nil {
			return
		}

		mu.Lock()
		defer mu.Unlock()

		fmt.Println("Writing to file: ", filePath)
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Error opening file: %v", err)
		}
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	} else {
		for _, childUrl := range body.Children {
			wg.Add(1)
			go traverseEndpointWrapper(childUrl, result, wg, mu)
		}
	}
}

var semaphore = make(chan struct{}, 100)

// func main() {
// 	var wg sync.WaitGroup
// 	var mu sync.Mutex

// 	for _, chapter := range ICD10_CHAPTERS {
// 		wg.Add(1)
// 		go traverseEndpointWrapper(chapter, ICD10Data{}, &wg, &mu)

// 	}

// 	wg.Wait()
// }
