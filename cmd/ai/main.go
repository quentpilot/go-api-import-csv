package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	var prompt PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&prompt); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ollamaReq := OllamaRequest{
		Model:  "llama3",
		Prompt: prompt.Prompt,
	}

	jsonBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "Error contacting Ollama", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Fprint(w, string(body))
}

func main() {
	http.HandleFunc("/ask", handler)
	fmt.Println("Listening on :4242")
	http.ListenAndServe(":4242", nil)
}
