package commands

import (
	"time"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/output"
	appPkg "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/spf13/cobra"
)

func NewSchedulerCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Manage post scheduling",
	}
	cmd.AddCommand(newSchedulerRunCmd(a, pretty))
	cmd.AddCommand(newSchedulerStatusCmd(a, pretty))
	cmd.AddCommand(newSchedulerDaemonCmd(a, pretty))
	return cmd
}

func newSchedulerRunCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "One-shot: check and publish all due posts now",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Drain report channel so PublishPost never blocks
			go func() {
				for range (*a).Scheduler.ReportChan {
				}
			}()
			(*a).Scheduler.CheckAndPublish()
			output.OK(cmd.OutOrStdout(), map[string]string{"message": "scheduler run complete"}, *pretty)
			return nil
		},
	}
}

func newSchedulerStatusCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show scheduled post queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			posts, err := (*a).DB.GetScheduledPosts()
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			type queueItem struct {
				ID          int64  `json:"id"`
				Caption     string `json:"caption"`
				ScheduledAt string `json:"scheduled_at"`
			}
			var items []queueItem
			for _, p := range posts {
				items = append(items, queueItem{
					ID:          p.ID,
					Caption:     p.Caption,
					ScheduledAt: p.ScheduledAt,
				})
			}
			if items == nil {
				items = []queueItem{}
			}
			output.OK(cmd.OutOrStdout(), map[string]interface{}{
				"queue_depth": len(items),
				"posts":       items,
			}, *pretty)
			return nil
		},
	}
}

func newSchedulerDaemonCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "daemon",
		Short: "Run scheduler as a long-lived background process (blocks)",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.OK(cmd.OutOrStdout(), map[string]string{"message": "scheduler daemon started"}, *pretty)
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			// Drain report channel so PublishPost never blocks
			go func() {
				for range (*a).Scheduler.ReportChan {
				}
			}()
			for range ticker.C {
				(*a).Scheduler.CheckAndPublish()
			}
			return nil
		},
	}
}
