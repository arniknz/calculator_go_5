package main

import (
	"log"

	application "github.com/arniknz/calculator_go_5/internal/app/orchestrator"
)

func main() {
	app := application.NewOrchestrator()
	log.Println("Starting Orchestrator on port:", app.Config.Port)
	if err := app.StartServer(); err != nil {
		log.Fatal(err)
	}
}
