package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type DeepSeekRequest struct {
	Model string `json:"model"`
	Messages []struct {
		Role string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

func callDeepSeek(query string, context []string, apiKey string) (string, int,error){
	//Combine context and new query
	fullPrompt := fmt.Sprintf("Context:\n%s\n\nNew Query: %s\n\nInstructions: Make the prompt concise and respond in a Telegram-friendly format. Use *bold* and _italic_ sparingly, avoid markdown headers, keep paragraphs short unless told to provide more detail, and use emoji when appropriate.", context, query)

	//build the request
	payload := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []struct{
			Role string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: fullPrompt},
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewReader(payloadBytes))
	req.Header.Set("Authorization", "Bearer " + apiKey)
	req.Header.Set("Content-Type", "application/json")

	//Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	
	if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
	return "", 0, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	//parse the response
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(result.Choices) == 0 {
		return "", 0, fmt.Errorf("no response from AI")
	}

	return result.Choices[0].Message.Content, result.Usage.TotalTokens, nil
}

func sanitizeForTelegram(text string) string {
	// Remove markdown headers
	text = regexp.MustCompile(`(?m)^#+\s*`).ReplaceAllString(text, "")

	// Convert bold/italic to Telegram format
    text = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(text, "*$1*")
    text = regexp.MustCompile(`__(.*?)__`).ReplaceAllString(text, "_$1_")

	// Remove code blocks
    text = regexp.MustCompile("(?s)```.*?```").ReplaceAllString(text, "")
    
    return text
}