package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

const apiKey = "sk-proj-UchxBLJyVrfedWH7JJFlT3BlbkFJ2yasZndADMp5uxSS3ji0"
const apiEndpoint = "https://api.openai.com/v1/chat/completions" // Updated endpoint

var history []string

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderHomePage(w, r)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	userInput := r.FormValue("userInput")

	if isSpecialRequest(userInput) {
		http.Error(w, "Your request was declined because it's not related to the vision of the touristic company", http.StatusBadRequest)
		return
	}

	client := resty.New()
	response, err := client.R().
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"role":    "system",
					"content": "You are a helpful assistant.",
				},
				{
					"role":    "user",
					"content": userInput,
				},
			},
			"model":      "gpt-3.5-turbo-1106", // Added model parameter
			"max_tokens": 512,
		}).
		Post(apiEndpoint)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error while sending the request: %v", err), http.StatusInternalServerError)
		return
	}

	body := response.Body()

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error while decoding JSON response: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("API Response:", data)

	choices, ok := data["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		http.Error(w, "Invalid response from GPT-3.5 Turbo API", http.StatusInternalServerError)
		return
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid message format in GPT-3.5 Turbo API response", http.StatusInternalServerError)
		return
	}

	content, ok := message["content"].(string)
	if !ok {
		http.Error(w, "Invalid content format in GPT-3.5 Turbo API message", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, content)

	history = append(history, userInput)
}

func isSpecialRequest(input string) bool {
	filterWords := []string{"virtual assistant", "tourist"}
	for _, word := range filterWords {
		if strings.Contains(strings.ToLower(input), word) {
			return true
		}
	}
	return false
}

func renderHomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
