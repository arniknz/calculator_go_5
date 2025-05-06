package main

import (
	"log"

	application "github.com/arniknz/calculator_go_5/internal/app/agent"
)

func main() {
	agent := application.NewAgent()
	log.Println("Starting Agent...")
	agent.Run()
}
