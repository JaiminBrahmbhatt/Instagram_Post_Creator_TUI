package main

import (
	"fmt"
	"log"
	"os"

	"net/http"
	"path/filepath"

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

	// Redirect logs to a file to avoid messing up the TUI
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	// Initialize DB
	database, err := db.InitDB("insta_auto_post.db")
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Initialize API Client
	client := api.NewClient(
		api.GetCredential("INSTA_ACCESS_TOKEN"),
		api.GetCredential("INSTA_IG_ID"),
	)

	// Start Scheduler
	scheduler := api.NewScheduler(database, client)
	scheduler.Start()

	// Start File Server to expose photos to Instagram
	go func() {
		// Get photos directory from DB
		var photosDir string
		err := database.Conn.QueryRow("SELECT value FROM settings WHERE key = 'photos_dir'").Scan(&photosDir)
		if err != nil || photosDir == "" {
			photosDir = "photos" // Fallback
		}

		absPath, _ := filepath.Abs(photosDir)
		mux := http.NewServeMux()
		mux.Handle("/", http.FileServer(http.Dir(absPath)))

		log.Printf("Ready! Photo server listening on http://localhost:8080 (serving %s)", absPath)
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Printf("Photo server error: %v", err)
		}
	}()

	// Start TUI
	p := tea.NewProgram(tui.InitialModel(database, client, scheduler.ReportChan, scheduler.TriggerChan), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
