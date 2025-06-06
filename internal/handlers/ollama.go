package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"go-csv-import/internal/logger"
	"net/http"
	"time"

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

type OllamaChunkResponse struct {
	Response  string `json:"response"`
	CreatedAt string `json:"created_at"`
	Done      bool   `json:"done"`
}

// Simple ask a question and return single response string
func AskOllama() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Call endpoint /ask")
		var prompt PromptRequest
		if err := json.NewDecoder(c.Request.Body).Decode(&prompt); err != nil {
			logger.Error("Invalid request", "error", err)
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
			logger.Error("Cannot connect to Ollama API", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Cannot connect to Ollama API: " + err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		//body, _ := io.ReadAll(resp.Body)
		var response OllamaResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			logger.Error("Invalid Ollama response", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Invalid Ollama Response: " + err.Error(),
			})
			return
		}

		duration := time.Duration(int64(response.Duration))
		c.JSON(http.StatusOK, gin.H{
			"message":  response.Response,
			"duration": duration.String(),
		})
	}
}

// Chat with streamed response
func ChatWithOllama() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Call endpoint /chat")
		var prompt PromptRequest
		if err := json.NewDecoder(c.Request.Body).Decode(&prompt); err != nil {
			logger.Error("Invalid request", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Invalid Request: " + err.Error(),
			})
			return
		}

		ollamaReq := OllamaRequest{
			Model:  "llama3",
			Stream: true,
			Prompt: prompt.Prompt,
		}

		jsonBody, _ := json.Marshal(ollamaReq)
		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://ollama:11434/api/generate", bytes.NewBuffer(jsonBody))
		if err != nil {
			logger.Error("Cannot connect to Ollama API", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Cannot connect to Ollama API: " + err.Error(),
			})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Cannot connect to Ollama API", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Connection to Ollama failed"})
			return
		}
		defer resp.Body.Close()

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Flush()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			//var chunk map[string]interface{}
			var chunk OllamaChunkResponse
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				logger.Error("Malformed stream response", "error", err)
				continue // ignore parsing errors
			} else {
				logger.Info("token streamed", "chunk", chunk)
			}

			token := chunk.Response
			c.Writer.Write([]byte(token))
			c.Writer.Flush()

			if chunk.Done {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Error("Error reading Ollama stream", "error", err)
		}
	}
}
