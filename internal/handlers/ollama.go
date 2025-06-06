package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Duration int    `json:"total_duration"`
}

func ChatWithOllama() gin.HandlerFunc {
	return func(c *gin.Context) {

		var prompt PromptRequest
		if err := json.NewDecoder(c.Request.Body).Decode(&prompt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Invalid Request: " + err.Error(),
			})
			return
		}

		ollamaReq := OllamaRequest{
			Model:  "llama3",
			Stream: false,
			Prompt: prompt.Prompt,
		}

		jsonBody, _ := json.Marshal(ollamaReq)
		resp, err := http.Post("http://ollama:11434/api/generate", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Cannot connect to Ollama API: " + err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		//body, _ := io.ReadAll(resp.Body)
		var response OllamaResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Invalid Ollama Response: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  response.Response,
			"duration": response.Duration,
		})
	}
}
