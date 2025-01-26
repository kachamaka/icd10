package models

// Model for indexing data in elastic search
type ICD10IndexRequest struct {
	ICD10Code    string   `json:"icd10code"`
	Type         string   `json:"type"`
	Title        string   `json:"title"`
	Chapter      string   `json:"chapter"`
	ChapterCode  string   `json:"chapterCode"`
	Block        string   `json:"block"`
	BlockCode    string   `json:"blockCode"`
	Category     string   `json:"category"`
	CategoryCode string   `json:"categoryCode"`
	Subcategory  string   `json:"subcategory"`
	Inclusion    []string `json:"inclusion"`
	Exclusion    []string `json:"exclusion"`
	Symptoms     []string `json:"symptoms"`
}

// Model for search handler request
type ICD10SearchQuery struct {
	Query string `json:"query"`
}

// Model for elastic search response
type ICD10SearchResponse struct {
	ICD10Code    string   `json:"icd10code"`
	Score        float64  `json:"score"`
	Title        string   `json:"title"`
	Type         string   `json:"type"`
	Chapter      string   `json:"chapter"`
	ChapterCode  string   `json:"chapterCode"`
	Block        string   `json:"block"`
	BlockCode    string   `json:"blockCode"`
	Category     string   `json:"category"`
	CategoryCode string   `json:"categoryCode"`
	Subcategory  string   `json:"subcategory"`
	Symptoms     []string `json:"symptoms"`
}

// Model for search handler response
type SearchResponse struct {
	ICD10Codes []ICD10SearchResponse `json:"icd10codes"`
}
