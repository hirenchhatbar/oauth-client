package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}
	form := url.Values{}
	form.Add("code", code)
	form.Add("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	form.Add("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	form.Add("redirect_uri", os.Getenv("HTTP_SCHEME")+"://"+os.Getenv("HTTP_HOST")+":"+(os.Getenv("HTTP_PORT")))
	form.Add("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		http.Error(w, "Failed to make token request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.WriteFile("token.json", bodyBytes, 0o600)
	if err != nil {
		http.Error(w, "Failed to save token file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bodyBytes)
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("token.json")
	if err != nil {
		http.Error(w, "Unable to read token.json", http.StatusInternalServerError)
		return
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		http.Error(w, "Unable to parse token.json", http.StatusInternalServerError)
		return
	}

	refreshToken := token.RefreshToken
	if refreshToken == "" {
		http.Error(w, "No refresh_token found", http.StatusBadRequest)
		return
	}

	form := url.Values{}
	form.Add("refresh_token", refreshToken)
	form.Add("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	form.Add("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	form.Add("redirect_uri", os.Getenv("HTTP_SCHEME")+"://"+os.Getenv("HTTP_HOST")+":"+(os.Getenv("HTTP_PORT")))
	form.Add("grant_type", "refresh_token")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		http.Error(w, "Failed to make token request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var newToken Token
	if err := json.Unmarshal(bodyBytes, &newToken); err != nil {
		http.Error(w, "Failed to parse Google response", http.StatusInternalServerError)
		return
	}

	// Preserve old refresh_token if not included in response
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = token.RefreshToken
	}

	file, _ := json.MarshalIndent(newToken, "", "  ")

	err = os.WriteFile("token.json", file, 0o600)
	if err != nil {
		http.Error(w, "Failed to save token file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(file)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Environment variables loaded...")
	fmt.Println("Starting server " + os.Getenv("HTTP_SCHEME") + "://" + os.Getenv("HTTP_HOST") + ":" + os.Getenv("HTTP_PORT"))

	http.HandleFunc("/", handler)
	http.HandleFunc("/token/refresh", refreshHandler)

	errH := http.ListenAndServe(":"+os.Getenv("HTTP_PORT"), nil)
	if errH != nil {
		panic(errH)
	}
}
