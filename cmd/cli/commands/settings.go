package commands

import (
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/output"
	appPkg "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/spf13/cobra"
)

func NewSettingsCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage application settings",
	}
	cmd.AddCommand(newSettingsGetCmd(a, pretty))
	cmd.AddCommand(newSettingsSetCmd(a, pretty))
	cmd.AddCommand(newSettingsListCmd(a, pretty))
	return cmd
}

func newSettingsGetCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var key string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a setting value",
		RunE: func(cmd *cobra.Command, args []string) error {
			val, err := (*a).DB.GetSetting(key)
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			output.OK(cmd.OutOrStdout(), map[string]string{"key": key, "value": val}, *pretty)
			return nil
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Setting key (required)")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newSettingsSetCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var key, value string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a setting value",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := (*a).DB.SetSetting(key, value); err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			output.OK(cmd.OutOrStdout(), map[string]string{"key": key, "value": value}, *pretty)
			return nil
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Setting key (required)")
	cmd.Flags().StringVar(&value, "value", "", "Setting value (required)")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("value")
	return cmd
}

func newSettingsListCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			rows, err := (*a).DB.Conn.Query("SELECT key, value FROM settings ORDER BY key")
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			defer rows.Close()
			var settings []map[string]string
			for rows.Next() {
				var k, v string
				if err := rows.Scan(&k, &v); err != nil {
					continue
				}
				settings = append(settings, map[string]string{"key": k, "value": v})
			}
			if settings == nil {
				settings = []map[string]string{}
			}
			output.OK(cmd.OutOrStdout(), map[string]interface{}{"settings": settings}, *pretty)
			return nil
		},
	}
}
