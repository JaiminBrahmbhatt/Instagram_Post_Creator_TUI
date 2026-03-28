package app

import (
	"context"
	"log"
	"net/http"
	"path/filepath"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
)

type Config struct {
	DBPath     string
	SkipTunnel bool
	HTTPPort   string // default "8080"
}

type App struct {
	DB        *db.Database
	Client    *api.Client
	Scheduler *api.Scheduler
	TunnelURL string
	server    *http.Server
}

func New(cfg Config) (*App, error) {
	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	client := api.NewClient(
		api.GetCredential("INSTA_ACCESS_TOKEN"),
		api.GetCredential("INSTA_IG_ID"),
	)
	return &App{
		DB:        database,
		Client:    client,
		Scheduler: api.NewScheduler(database, client),
	}, nil
}

func (a *App) StartInfra(ctx context.Context, cfg Config) {
	if cfg.SkipTunnel {
		return
	}
	port := cfg.HTTPPort
	if port == "" {
		port = "8080"
	}
	photosDir, _ := a.DB.GetSetting("photos_dir")
	if photosDir == "" {
		photosDir = "photos"
	}
	absPath, _ := filepath.Abs(photosDir)
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(absPath)))
	a.server = &http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		log.Printf("Photo server on :%s (serving %s)", port, absPath)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Photo server error: %v", err)
		}
	}()
	token := api.GetNgrokToken()
	if token == "" {
		log.Println("Ngrok token not found. Skipping tunnel.")
		return
	}
	go func() {
		url, err := api.StartTunnel(ctx, token, mux)
		if err != nil {
			log.Printf("Tunnel failed: %v", err)
			return
		}
		a.TunnelURL = url
		log.Printf("Tunnel active: %s", url)
	}()
}

func (a *App) Shutdown(ctx context.Context) {
	api.StopTunnel()
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}
}
