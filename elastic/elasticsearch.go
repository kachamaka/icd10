package elastic

import (
	"ICD-10/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"gopkg.in/ini.v1"
)

var (
	elasticClient    *elasticsearch.Client
	SEARCH_INDEX     string
	ELASTIC_ENDPOINT string
	API_KEY          string
)

type Config struct {
	SearchIndex     string
	ElasticEndpoint string
	APIKey          string
}

func init() {
	config, err := loadElasticConfig("./elastic/config.ini")
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	SEARCH_INDEX = config.SearchIndex
	ELASTIC_ENDPOINT = config.ElasticEndpoint
	API_KEY = config.APIKey

	createClient()
	createIndexIfNotExists()
}

func loadElasticConfig(filePath string) (Config, error) {
	// Load the INI file
	cfg, err := ini.Load(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load INI file: %v", err)
	}

	// Initialize config struct with values from the INI file
	var config Config
	section := cfg.Section("elasticsearch")

	config.SearchIndex = section.Key("search_index").String()
	config.ElasticEndpoint = section.Key("elastic_endpoint").String()
	config.APIKey = section.Key("api_key").String()

	return config, nil
}

func createClient() {
	cfg := elasticsearch.Config{
		Addresses: []string{ELASTIC_ENDPOINT},
		APIKey:    "U2YxemU1UUJWeS1lWHllZnhNV2I6QXI4VXhlaVNTX0dYM1FQbFU3YS1IUQ==",
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	elasticClient = client
}

func createIndexIfNotExists() {
	// Define the index settings and mappings with edge_ngram analyzer
	indexMapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"analysis": map[string]interface{}{
				"tokenizer": map[string]interface{}{
					"edge_ngram_tokenizer": map[string]interface{}{
						"type":        "edge_ngram",
						"min_gram":    3,
						"max_gram":    25,
						"token_chars": []string{"letter", "digit"},
					},
				},
				"analyzer": map[string]interface{}{
					"edge_ngram_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "edge_ngram_tokenizer",
						"filter":    []string{"lowercase"}, // Add lowercase filter for case insensitivity
					},
				},
				"normalizer": map[string]interface{}{
					"case_insensitive_normalizer": map[string]interface{}{
						"type":   "custom",
						"filter": []string{"lowercase"},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":     "text",                // Used for full-text search, including prefix matching
					"analyzer": "edge_ngram_analyzer", // Prefix matching
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":       "keyword",                     // Exact match field
							"normalizer": "case_insensitive_normalizer", // Case-insensitive exact match
						},
					},
				},
				"category": map[string]interface{}{
					"type":     "text",
					"analyzer": "edge_ngram_analyzer",
				},
				"chapter": map[string]interface{}{
					"type":     "text",
					"analyzer": "edge_ngram_analyzer",
				},
				"block": map[string]interface{}{
					"type":     "text",
					"analyzer": "edge_ngram_analyzer",
				},
				"icd10code": map[string]interface{}{
					"type":     "text",
					"analyzer": "edge_ngram_analyzer",
				},
				"inclusion": map[string]interface{}{
					"type":     "text",
					"analyzer": "edge_ngram_analyzer",
				},
				"exclusion": map[string]interface{}{
					"type":     "text",
					"analyzer": "edge_ngram_analyzer",
				},
				"symptoms": map[string]interface{}{
					"type":     "text",
					"analyzer": "edge_ngram_analyzer", // Apply the edge_ngram_analyzer for partial matching
				},
			},
		},
	}

	// Convert the map to JSON
	body, err := json.Marshal(indexMapping)
	if err != nil {
		log.Fatalf("Error marshaling index mapping: %s", err)
	}

	// check if index exists
	existsReq := esapi.IndicesExistsRequest{
		Index: []string{SEARCH_INDEX},
	}

	existsRes, err := existsReq.Do(context.Background(), elasticClient)
	if err != nil {
		log.Fatalf("Error checking if index exists: %s", err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode == http.StatusOK {
		fmt.Printf("Index %s already exists!\n", SEARCH_INDEX)
		return
	}

	res, err := elasticClient.Indices.Create(
		SEARCH_INDEX,
		elasticClient.Indices.Create.WithBody(bytes.NewReader(body)))
	if err != nil {
		log.Fatalf("Error creating index: %s", err)
	}

	if res.IsError() {
		log.Printf("Error creating index: %s", res.String())
	} else {
		fmt.Printf("Index %s created successfully!\n", SEARCH_INDEX)
	}
	defer res.Body.Close()
}

func optimizeQuery(query string) string {
	if words := strings.Fields(query); len(words) == 1 {
		query += " unspecified"
	}

	return query
}

func Search(query string) ([]models.ICD10SearchResponse, error) {
	var buf bytes.Buffer

	query = optimizeQuery(query)

	// Construct the query
	elasticQuery := map[string]interface{}{
		"size": 42,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query": query,
							"fields": []string{
								"title^3",         // Boost `title` field
								"title.keyword^5", // Boost exact matches on `title.keyword`
								"category^3",
								"categoryCode",
								"block",
								"blockCode",
								"chapter",
								"icd10code",
							},
							"operator": "or",
						},
					},
				},
				"should": []interface{}{
					// Prefix matching in title for partial matches
					map[string]interface{}{
						"match_phrase_prefix": map[string]interface{}{
							"title": map[string]interface{}{
								"query": query,
								"boost": 2, // Lower boost for fallback
							},
						},
					},
					// Optional inclusion match
					map[string]interface{}{
						"match_phrase_prefix": map[string]interface{}{
							"inclusion": query,
						},
					},
				},
				// "must_not": []interface{}{
				// 	map[string]interface{}{
				// 		"match_phrase": map[string]interface{}{
				// 			"exclusion": query,
				// 		},
				// 	},
				// },
				"minimum_should_match": 0, // Ensure results return even if `inclusion` doesn't match
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(elasticQuery); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	res, err := elasticClient.Search(
		elasticClient.Search.WithIndex(SEARCH_INDEX),
		elasticClient.Search.WithBody(&buf),
	)

	defer res.Body.Close()

	if err != nil {
		log.Println("Error executing search: ", err)
		return nil, err
	}

	if res.IsError() {
		log.Printf("Error response from Elasticsearch: %s", res.String())
		return nil, fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Println("error decoding res", err)
		return nil, err
	}

	// Print the result in JSON format
	// resultJSON, err := json.MarshalIndent(r, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Error marshalling result to JSON: %s", err)
	// }
	// fmt.Println(string(resultJSON))

	var icd10Codes []models.ICD10SearchResponse

	if hits, ok := r["hits"].(map[string]interface{}); ok {
		if hitsHits, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitsHits {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					if icd10code, ok := hitMap["_id"].(string); ok {
						if source, ok := hitMap["_source"].(map[string]interface{}); ok {
							symptoms := []string{}
							if source["symptoms"] != nil {
								symptomsSource := source["symptoms"].([]interface{})
								for _, s := range symptomsSource {
									symptoms = append(symptoms, s.(string))
								}
							}
							icd10Codes = append(icd10Codes, models.ICD10SearchResponse{
								ICD10Code:    icd10code,
								Score:        hitMap["_score"].(float64),
								Type:         source["type"].(string),
								Title:        source["title"].(string),
								Chapter:      source["chapter"].(string),
								ChapterCode:  source["chapterCode"].(string),
								Category:     source["category"].(string),
								CategoryCode: source["categoryCode"].(string),
								Block:        source["block"].(string),
								BlockCode:    source["blockCode"].(string),
								Subcategory:  source["subcategory"].(string),
								Symptoms:     symptoms,
							})
						}
					}
				}
			}
		}
	}

	// sort by score highest -> lowest
	sort.Slice(icd10Codes, func(i, j int) bool {
		if icd10Codes[i].Score == icd10Codes[j].Score {
			return icd10Codes[i].ICD10Code < icd10Codes[j].ICD10Code
		}

		return icd10Codes[i].Score > icd10Codes[j].Score
	})

	return icd10Codes, nil
}

func IndexateICD10Code(icd10IndexRequest models.ICD10IndexRequest) {
	data, err := json.Marshal(icd10IndexRequest)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	req := esapi.IndexRequest{
		Index:      SEARCH_INDEX,
		DocumentID: string(icd10IndexRequest.ICD10Code),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	resp, err := req.Do(context.Background(), elasticClient)
	if err != nil {
		log.Fatalf("Error getting response: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Indexed document %s to index %s\n", resp.String(), SEARCH_INDEX)
}
