package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/commands"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	pretty      bool
	skipTunnel  bool
	dbPath      string
	application *app.App
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	if err := buildRoot().Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "post-creator-cli",
		Short: "Instagram post automation CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if strings.Contains(cmd.CommandPath(), " completion") || cmd.Name() == "help" {
				return nil
			}
			a, err := app.New(app.Config{DBPath: dbPath})
			if err != nil {
				return fmt.Errorf("init: %w", err)
			}
			application = a
			a.StartInfra(context.Background(), app.Config{
				SkipTunnel: skipTunnel,
				HTTPPort:   "8081",
			})
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if application != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				application.Shutdown(ctx)
			}
		},
	}

	root.PersistentFlags().BoolVar(&pretty, "pretty", false, "Pretty-print JSON")
	root.PersistentFlags().BoolVar(&skipTunnel, "skip-tunnel", false, "Skip HTTP server + ngrok")
	root.PersistentFlags().StringVar(&dbPath, "db", "post_creator.db", "SQLite DB path")

	root.AddCommand(commands.NewAuthCmd(&application, &pretty))
	root.AddCommand(commands.NewQuotaCmd(&application, &pretty))
	root.AddCommand(commands.NewSettingsCmd(&application, &pretty))
	root.AddCommand(commands.NewMediaCmd(&application, &pretty))
	root.AddCommand(commands.NewPostCmd(&application, &pretty))
	root.AddCommand(commands.NewSchedulerCmd(&application, &pretty))

	return root
}
