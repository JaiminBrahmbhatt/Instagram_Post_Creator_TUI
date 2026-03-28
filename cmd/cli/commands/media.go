package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/api"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/output"
	appPkg "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/spf13/cobra"
)

func NewMediaCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Manage local media files",
	}
	cmd.AddCommand(newMediaScanCmd(a, pretty))
	cmd.AddCommand(newMediaListCmd(a, pretty))
	cmd.AddCommand(newMediaIgnoreCmd(a, pretty))
	return cmd
}

func newMediaScanCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan photos directory and register new media",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dir == "" {
				d, err := (*a).DB.GetSetting("photos_dir")
				if err != nil || d == "" {
					d = "photos"
				}
				dir = d
			}
			absDir, _ := filepath.Abs(dir)
			entries, err := os.ReadDir(absDir)
			if err != nil {
				output.Err(cmd.OutOrStdout(), fmt.Sprintf("cannot read dir %s: %v", absDir, err), *pretty)
				return nil
			}
			var added int
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(e.Name()))
				if !slices.Contains(api.SupportedExtensions, ext) {
					continue
				}
				path := filepath.Join(absDir, e.Name())
				// Insert ignoring duplicates
				_, err := (*a).DB.Conn.Exec(
					`INSERT OR IGNORE INTO media (path, hash) VALUES (?, ?)`,
					path, path, // use path as hash placeholder if no hash function available
				)
				if err == nil {
					added++
				}
			}
			output.OK(cmd.OutOrStdout(), map[string]int{"scanned": len(entries), "added": added}, *pretty)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "Directory to scan (default: photos_dir setting)")
	return cmd
}

func newMediaListCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var unposted bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered media",
		RunE: func(cmd *cobra.Command, args []string) error {
			query := "SELECT id, path, is_posted, ignore FROM media ORDER BY created_at DESC"
			if unposted {
				query = "SELECT id, path, is_posted, ignore FROM media WHERE is_posted=0 AND ignore=0 ORDER BY created_at DESC"
			}
			rows, err := (*a).DB.Conn.Query(query)
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			defer rows.Close()
			type mediaItem struct {
				ID       int64  `json:"id"`
				Path     string `json:"path"`
				IsPosted bool   `json:"is_posted"`
				Ignored  bool   `json:"ignored"`
			}
			var items []mediaItem
			for rows.Next() {
				var m mediaItem
				rows.Scan(&m.ID, &m.Path, &m.IsPosted, &m.Ignored)
				items = append(items, m)
			}
			if items == nil {
				items = []mediaItem{}
			}
			output.OK(cmd.OutOrStdout(), map[string]interface{}{"media": items}, *pretty)
			return nil
		},
	}
	cmd.Flags().BoolVar(&unposted, "unposted", false, "Show only unposted, non-ignored media")
	return cmd
}

func newMediaIgnoreCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var id int64
	cmd := &cobra.Command{
		Use:   "ignore",
		Short: "Mark a media file as ignored",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := (*a).DB.Conn.Exec("UPDATE media SET ignore=1 WHERE id=?", id)
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			output.OK(cmd.OutOrStdout(), map[string]int64{"id": id}, *pretty)
			return nil
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "Media ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}
