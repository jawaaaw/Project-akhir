package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type AIModelConnector struct {
	Client *http.Client
}

type Inputs struct {
	Table        map[string][]string `json:"table"`
	Query        string              `json:"query"`
	WaitForModel bool                `json:"wait_for_model"`
}

type Response struct {
	Answer      string   `json:"answer"`
	Coordinates [][]int  `json:"coordinates"`
	Cells       []string `json:"cells"`
	Aggregator  string   `json:"aggregator"`
}

// CsvToSlice converts CSV data to a map
func CsvToSlice(data string) (map[string][]string, error) {
	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	header := records[0]
	result := make(map[string][]string)

	for _, h := range header {
		result[h] = []string{}
	}

	for _, record := range records[1:] {
		for i, value := range record {
			key := header[i]
			result[key] = append(result[key], value)
		}
	}

	return result, nil
}

func (c *AIModelConnector) ConnectAIModel(payload interface{}, token string) (Response, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequest("POST", "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq", strings.NewReader(string(payloadBytes)))
	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	var aiResponse Response
	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		return Response{}, err
	}

	return aiResponse, nil
}

func main() {
	// Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	// Read the CSV file
	csvFile, err := os.Open("data-series.csv")
	if err != nil {
		fmt.Println("Error opening CSV file:", err)
		return
	}
	defer csvFile.Close()

	csvData, err := ioutil.ReadAll(csvFile)
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	// Convert CSV to map
	table, err := CsvToSlice(string(csvData))
	if err != nil {
		fmt.Println("Error converting CSV to map:", err)
		return
	}

	// Create payload for AI model
	query := "Berapa Total Energy di tahun 2021-01-01?"
	payload := Inputs{
		Table:        table,
		Query:        query,
		WaitForModel: true,
	}

	// Connect to AI model
	connector := AIModelConnector{Client: &http.Client{}}
	token := os.Getenv("HUGGINGFACE_TOKEN")
	if token == "" {
		fmt.Println("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_TOKEN is not set in the environment variables.")
		return
	}
	response, err := connector.ConnectAIModel(payload, token)
	if err != nil {
		fmt.Println("Error connecting to AI model:", err)
		return
	}

	fmt.Printf("Answer: %s\n", response.Answer)
	fmt.Printf("Coordinates: %v\n", response.Coordinates)
	fmt.Printf("Cells: %v\n", response.Cells)
	fmt.Printf("Aggregator: %s\n", response.Aggregator)
}
