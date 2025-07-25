package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/hirenchhatbar/oauth-client/pkg/oauth"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Environment variables loaded...")

	port, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))

	fmt.Println("Starting server " + os.Getenv("HTTP_SCHEME") + "://" + os.Getenv("HTTP_HOST") + ":" + os.Getenv("HTTP_PORT"))

	errH := oauth.Listen(port)

	if errH != nil {
		panic(errH)
	}
}
