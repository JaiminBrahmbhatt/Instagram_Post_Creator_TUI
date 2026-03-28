package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
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
		fmt.Printf("error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	a, err := app.New(app.Config{DBPath: "post_creator.db"})
	if err != nil {
		log.Fatalf("Error initialising app: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	skipTunnel := os.Getenv("SKIP_TUNNEL") == "true"
	a.StartInfra(ctx, app.Config{SkipTunnel: skipTunnel, HTTPPort: "8080"})
	defer a.Shutdown(context.Background())

	a.Scheduler.Start()

	p := tea.NewProgram(
		tui.InitialModel(a.DB, a.Client, a.Scheduler.ReportChan, a.Scheduler.TriggerChan),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
		os.Exit(1)
	}
}
