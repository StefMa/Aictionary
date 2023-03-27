package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const api = "https://api.openai.com/v1/completions"

func ExistHandler(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	word := r.URL.Query().Get("word")
	prompt := fmt.Sprintf(
		"Does the following word exist in the %s dictionary: %s. Answer only with true if its included or false if it's not included, in lowercase. I don't want to see explanations.",
		lang,
		word,
	)

	jsonBody := []byte(`{
		"model": "text-davinci-003",
		"prompt": "` + prompt + `",
		"temperature": 0
	}`)

	request, err := http.NewRequest(http.MethodPost, api, bytes.NewBuffer(jsonBody))
	if err != nil {
		writeErrorResponse(w, "Error while creating OpenAi Api request")
		return
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		errorMessage := fmt.Sprintf("Error while executing OpenAi Api request:\n %s", err)
		writeErrorResponse(w, errorMessage)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		openApiErrorReponse := &OpenApiErrorResponse{}
		err = json.NewDecoder(res.Body).Decode(openApiErrorReponse)
		if err != nil {
			errorMessage := fmt.Sprintf("Can't decode OpenApiErrorResponse:\n %s", err)
			writeErrorResponse(w, errorMessage)
			return
		}
		fmt.Fprintf(w, "Got OpenAi Api error. Message:\n %s", openApiErrorReponse.Error.Message)
		return
	}

	openApiReponse := &OpenApiResponse{}
	err = json.NewDecoder(res.Body).Decode(openApiReponse)
	if err != nil {
		errorMessage := fmt.Sprintf("Can't decode OpenApiResponse:\n %s", err)
		writeErrorResponse(w, errorMessage)
		return
	}
	wordExist := false
	if strings.Contains(openApiReponse.Choices[0].Text, "true") {
		wordExist = true
	}
	response := Response{
		Word:     word,
		Language: lang,
		Exist:    wordExist,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		errorMessage := fmt.Sprintf("Can't encode Response:\n %s", err)
		writeErrorResponse(w, errorMessage)
		return
	}
	fmt.Fprintf(w, string(jsonResponse))
}

func writeErrorResponse(w io.Writer, errorMessage string) {
	errorResponse := ErrorResponse{Message: errorMessage}
	jsonResponse, err := json.Marshal(errorResponse)
	if err != nil {
		fmt.Fprintf(w, genericErrorResponse())
	}
	fmt.Fprintf(w, string(jsonResponse))
}

func genericErrorResponse() string {
	jsonBody := []byte(`{
		"message": "Uh no! Something went wrong. Please try again later",
	}`)
	return string(jsonBody)
}

type Response struct {
	Word     string `json:"word"`
	Language string `json:"lang"`
	Exist    bool   `json:"exist"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type OpenApiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		Logprobs     any    `json:"logprobs"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type OpenApiErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   any    `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}
