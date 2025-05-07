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
