package oauth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type Config struct {
	HttpScheme         string
	HttpHost           string
	HttpPort           int
	GoogleClientId     string
	GoogleClientSecret string
	TokenJsonPath      string
}

func (config *Config) url() string {
	return config.HttpScheme + "://" + config.HttpHost + ":" + strconv.Itoa(config.HttpPort)
}

func LoadToken(config *Config, code string) ([]byte, error) {
	form := url.Values{}
	form.Add("code", code)
	form.Add("client_id", config.GoogleClientId)
	form.Add("client_secret", config.GoogleClientSecret)
	form.Add("redirect_uri", config.url())
	form.Add("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		return nil, errors.New("Error while running POST request: " + err.Error())
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error while reading response of POST request: " + err.Error())
	}

	err = os.WriteFile(config.TokenJsonPath, bodyBytes, 0o600)
	if err != nil {
		return nil, errors.New("Error while writing JSON file: " + err.Error())
	}

	return bodyBytes, nil
}

func RefreshToken(config *Config) ([]byte, error) {
	data, err := os.ReadFile(config.TokenJsonPath)
	if err != nil {
		return nil, errors.New("Unable to read " + config.TokenJsonPath + ": " + err.Error())
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, errors.New("Unable to parse " + config.TokenJsonPath + ": " + err.Error())
	}

	refreshToken := token.RefreshToken
	if refreshToken == "" {
		return nil, errors.New("No refresh_token found: " + err.Error())
	}

	form := url.Values{}
	form.Add("refresh_token", refreshToken)
	form.Add("client_id", config.GoogleClientId)
	form.Add("client_secret", config.GoogleClientSecret)
	form.Add("redirect_uri", config.url())
	form.Add("grant_type", "refresh_token")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		return nil, errors.New("Failed to make token request: " + err.Error())
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Failed to read response: " + err.Error())
	}

	var newToken Token
	if err := json.Unmarshal(bodyBytes, &newToken); err != nil {
		return nil, errors.New("Failed to parse Google response: " + err.Error())
	}

	// Preserve old refresh_token if not included in response
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = token.RefreshToken
	}

	file, _ := json.MarshalIndent(newToken, "", "  ")

	err = os.WriteFile(config.TokenJsonPath, file, 0o600)
	if err != nil {
		return nil, errors.New("Failed to save token file: " + err.Error())
	}

	return file, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	port, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))

	token, err := LoadToken(&Config{
		HttpScheme:         os.Getenv("HTTP_SCHEME"),
		HttpHost:           os.Getenv("HTTP_HOST"),
		HttpPort:           port,
		GoogleClientId:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		TokenJsonPath:      os.Getenv("TOKEN_JSON_PATH"),
	}, code)
	if err != nil {
		http.Error(w, "Failed to load token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(token)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	port, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))

	token, err := RefreshToken(&Config{
		HttpScheme:         os.Getenv("HTTP_SCHEME"),
		HttpHost:           os.Getenv("HTTP_HOST"),
		HttpPort:           port,
		GoogleClientId:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		TokenJsonPath:      os.Getenv("TOKEN_JSON_PATH"),
	})
	if err != nil {
		http.Error(w, "Failed to load token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(token)
}

func Listen(port int) error {
	http.HandleFunc("/", handler)
	http.HandleFunc("/token/refresh", refreshHandler)

	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
