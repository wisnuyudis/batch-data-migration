package tokenizer

import (
	"batch-data-migration/config"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type Helper struct {
	baseURL          string
	tokenGroup       string
	tokenTemplate    string
	tokenizeUser     string
	tokenizePassword string
}

type TokenizeRequest struct {
	TokenGroup    string `json:"tokengroup"`
	Data          string `json:"data"`
	TokenTemplate string `json:"tokentemplate"`
}

type TokenizeResponse struct {
	Token string `json:"token"`
}

type BatchTokenizeRequest struct {
	Requests []TokenizeRequest `json:"requests"`
}

type BatchTokenizeResponse struct {
	Tokens []TokenizeResponse `json:"responses"`
}

// NewHelper untuk membuat helper baru
func NewHelper() *Helper {
	return &Helper{
		baseURL:          config.AppConfig.Tokenization.BaseURL,
		tokenGroup:       config.AppConfig.Tokenization.TokenGroup,
		tokenTemplate:    config.AppConfig.Tokenization.TokenTemplate,
		tokenizeUser:     config.AppConfig.Tokenization.TokenizeUser,
		tokenizePassword: config.AppConfig.Tokenization.TokenizePassword,
	}
}

// Tokenize untuk melakukan tokenisasi cc_number
func (h *Helper) Tokenize(ccNumber string) (string, error) {
	url := fmt.Sprintf("%s/tokenize", h.baseURL)

	// Menyiapkan request body untuk API tokenisasi
	reqBody := TokenizeRequest{
		TokenGroup:    h.tokenGroup,
		Data:          ccNumber,
		TokenTemplate: h.tokenTemplate,
	}

	// Mengubah body menjadi JSON
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Membuat HTTP client dengan Transport untuk menonaktifkan verifikasi sertifikat SSL
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	// Menambahkan Basic Authentication Header
	auth := fmt.Sprintf("%s:%s", h.tokenizeUser, h.tokenizePassword)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Content-Type", "application/json")

	// Melakukan request ke API
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to tokenization API: %v", err)
	}
	defer resp.Body.Close()

	// Memeriksa status code dari response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tokenization API responded with status: %v", resp.Status)
	}

	// Membaca dan mengurai response dari API
	var tokenResp TokenizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode API response: %v", err)
	}

	return tokenResp.Token, nil
}

// TokenizeBatch handles tokenizing multiple values in a single API call
func (h *Helper) TokenizeBatch(ccNumbers []string) ([]string, error) {
	url := fmt.Sprintf("%s/tokenize", h.baseURL)

	// Build array of tokenize requests
	var requests []TokenizeRequest
	for _, ccNumber := range ccNumbers {
		req := TokenizeRequest{
			TokenGroup:    h.tokenGroup,
			Data:          ccNumber,
			TokenTemplate: h.tokenTemplate,
		}
		requests = append(requests, req)
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %v", err)
	}

	// Create HTTP client with insecure transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	// Add authentication
	auth := fmt.Sprintf("%s:%s", h.tokenizeUser, h.tokenizePassword)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send batch request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tokenization API responded with status: %v", resp.Status)
	}

	// Parse response
	var tokenResponses []TokenizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponses); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %v", err)
	}

	// Extract tokens
	tokens := make([]string, len(tokenResponses))
	for i, tokenResp := range tokenResponses {
		tokens[i] = tokenResp.Token
	}

	return tokens, nil
}
