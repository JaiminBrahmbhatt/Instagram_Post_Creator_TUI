package commands

import (
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/output"
	appPkg "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/spf13/cobra"
)

func NewQuotaCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "quota",
		Short: "Fetch Instagram publishing quota",
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, err := (*a).Client.GetPublishingLimit()
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			output.OK(cmd.OutOrStdout(), map[string]interface{}{
				"quota_usage": limit.QuotaUsage,
				"quota_total": limit.Config.QuotaTotal,
			}, *pretty)
			return nil
		},
	}
}
