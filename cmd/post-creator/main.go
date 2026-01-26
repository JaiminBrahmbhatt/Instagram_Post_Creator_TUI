package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"net/http"
	"path/filepath"
	"strings"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	database, err := db.InitDB("post_creator.db")
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	client := api.NewClient(
		api.GetCredential("INSTA_ACCESS_TOKEN"),
		api.GetCredential("INSTA_IG_ID"),
	)

	scheduler := api.NewScheduler(database, client)
	scheduler.Start()

	var photosDir string
	err = database.Conn.QueryRow("SELECT value FROM settings WHERE key = 'photos_dir'").Scan(&photosDir)
	if err != nil || photosDir == "" {
		photosDir = "photos"
	}

	absPath, _ := filepath.Abs(photosDir)
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(absPath)))

	server := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		log.Printf("Ready! Photo server listening on http://localhost:8080 (serving %s)", absPath)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Photo server error: %v", err)
		}
	}()
	defer func() {
		log.Println("Shutting down photo server...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if os.Getenv("SKIP_TUNNEL") != "true" {
		token := api.GetNgrokToken()
		if token != "" {
			go func() {
				log.Println("Attempting to start native Ngrok tunnel...")
				url, err := api.StartTunnel(context.Background(), token, mux)
				if err != nil {
					log.Printf("⚠️ Failed to start Ngrok tunnel: %v", err)
				} else {
					log.Printf("Tunnel active! Public URL: %s", url)
					if !strings.HasSuffix(url, "/") {
						url += "/"
					}
					os.Setenv("PUBLIC_URL_PREFIX", url)
				}
			}()
			defer api.StopTunnel()
		} else {
			log.Println("Ngrok token not found in keyring. Skipping auto-tunnel.")
		}
	}

	p := tea.NewProgram(tui.InitialModel(database, client, scheduler.ReportChan, scheduler.TriggerChan), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
