package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hirenchhatbar/oauth-client/pkg/oauth"
)

func main() {
	err := oauth.Init()
	if err != nil {
		log.Fatal("Failed to init")
	}

	fmt.Println("App initiated...")

	fmt.Println("Starting server " + os.Getenv("HTTP_SCHEME") + "://" + os.Getenv("HTTP_HOST") + ":" + os.Getenv("HTTP_PORT"))

	errH := oauth.Listen()

	if errH != nil {
		panic(errH)
	}
}
