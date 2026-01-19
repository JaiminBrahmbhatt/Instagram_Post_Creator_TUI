package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaiminb/insta-auto-post/api"
	"github.com/jaiminb/insta-auto-post/db"
	"github.com/jaiminb/insta-auto-post/tui"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Initialize DB
	database, err := db.InitDB("insta_auto_post.db")
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Initialize API Client
	client := api.NewClient(
		os.Getenv("INSTA_ACCESS_TOKEN"),
		os.Getenv("INSTA_IG_ID"),
	)

	// Start Scheduler
	scheduler := api.NewScheduler(database, client)
	scheduler.Start()

	// Start TUI
	p := tea.NewProgram(tui.InitialModel(database), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
