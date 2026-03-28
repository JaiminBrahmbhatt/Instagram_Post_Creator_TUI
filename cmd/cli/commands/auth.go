package commands

import (
	"io"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/output"
	appPkg "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/spf13/cobra"
)

func NewAuthCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Instagram credentials",
	}
	cmd.AddCommand(newAuthSetCmd(pretty))
	cmd.AddCommand(newAuthStatusCmd(pretty))
	return cmd
}

func newAuthSetCmd(pretty *bool) *cobra.Command {
	var token, igID string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Save Instagram credentials to keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := api.SetCredential("INSTA_ACCESS_TOKEN", token); err != nil {
				output.Err(cmd.OutOrStdout(), "failed to save access token: "+err.Error(), *pretty)
				return nil
			}
			if err := api.SetCredential("INSTA_IG_ID", igID); err != nil {
				output.Err(cmd.OutOrStdout(), "failed to save IG ID: "+err.Error(), *pretty)
				return nil
			}
			output.OK(cmd.OutOrStdout(), map[string]string{"message": "credentials saved"}, *pretty)
			return nil
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Instagram access token (required)")
	cmd.Flags().StringVar(&igID, "ig-id", "", "Instagram user ID (required)")
	cmd.MarkFlagRequired("token")
	cmd.MarkFlagRequired("ig-id")
	return cmd
}

func newAuthStatusCmd(pretty *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check whether credentials are present (values never printed)",
		RunE: func(cmd *cobra.Command, args []string) error {
			RunAuthStatus(cmd.OutOrStdout(), *pretty)
			return nil
		},
	}
}

// RunAuthStatus is exported for testing.
func RunAuthStatus(w io.Writer, pretty bool) {
	token := api.GetCredential("INSTA_ACCESS_TOKEN")
	igID := api.GetCredential("INSTA_IG_ID")
	output.OK(w, map[string]bool{
		"access_token": token != "",
		"ig_id":        igID != "",
	}, pretty)
}
