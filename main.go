package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hirenchhatbar/oauth-client/pkg/oauth"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Environment variables loaded...")

	fmt.Println("Starting server " + os.Getenv("HTTP_SCHEME") + "://" + os.Getenv("HTTP_HOST") + ":" + os.Getenv("HTTP_PORT"))

	errH := oauth.Listen()

	if errH != nil {
		panic(errH)
	}
}
