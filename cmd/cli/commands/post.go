package commands

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/cmd/cli/output"
	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
	appPkg "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/internal/app"
	"github.com/spf13/cobra"
)

func NewPostCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post",
		Short: "Manage Instagram posts",
	}
	cmd.AddCommand(newPostCreateCmd(a, pretty))
	cmd.AddCommand(newPostListCmd(a, pretty))
	cmd.AddCommand(newPostPublishCmd(a, pretty))
	cmd.AddCommand(newPostDeleteCmd(a, pretty))
	return cmd
}

func newPostCreateCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var media []string
	var caption, schedule string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a post (draft or scheduled)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(media) == 0 {
				output.Err(cmd.OutOrStdout(), "at least one --media path required", *pretty)
				return nil
			}
			if len(caption) > 2200 {
				output.Err(cmd.OutOrStdout(), "caption exceeds 2200 character limit", *pretty)
				return nil
			}
			status := db.StatusDraft
			if schedule != "" {
				if _, err := time.Parse(time.RFC3339, schedule); err != nil {
					output.Err(cmd.OutOrStdout(), "invalid --schedule: must be RFC3339 e.g. 2026-12-01T10:00:00Z", *pretty)
					return nil
				}
				status = db.StatusScheduled
			}
			id, err := (*a).DB.SavePost(caption, media, schedule, status)
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			if status == db.StatusScheduled && (*a).Scheduler != nil {
				(*a).Scheduler.Trigger()
			}
			output.OK(cmd.OutOrStdout(), map[string]interface{}{
				"post_id": id,
				"status":  string(status),
			}, *pretty)
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&media, "media", nil, "Media file path (repeatable)")
	cmd.Flags().StringVar(&caption, "caption", "", "Post caption")
	cmd.Flags().StringVar(&schedule, "schedule", "", "Schedule time in RFC3339 (e.g. 2026-12-01T10:00:00Z)")
	return cmd
}

func newPostListCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var statusFilter string
	var limit int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List posts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				limit = 50
			}
			type postItem struct {
				ID          int64  `json:"id"`
				Status      string `json:"status"`
				Caption     string `json:"caption"`
				ScheduledAt string `json:"scheduled_at,omitempty"`
				PublishedAt string `json:"published_at,omitempty"`
				CreatedAt   string `json:"created_at"`
				MediaCount  int    `json:"media_count"`
			}
			var items []postItem
			if statusFilter != "" {
				query := `
					SELECT p.id, p.caption,
					       COALESCE(p.scheduled_at, ''),
					       COALESCE(p.published_at, ''),
					       p.created_at,
					       p.status,
					       COUNT(pm.media_id)
					FROM posts p
					LEFT JOIN post_media pm ON p.id = pm.post_id
					WHERE p.status = ?
					GROUP BY p.id
					ORDER BY p.id DESC
					LIMIT ? OFFSET ?
				`
				rows, err := (*a).DB.Conn.Query(query, statusFilter, limit, 0)
				if err != nil {
					output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
					return nil
				}
				defer rows.Close()
				for rows.Next() {
					var p db.Post
					if err := rows.Scan(&p.ID, &p.Caption, &p.ScheduledAt, &p.PublishedAt, &p.CreatedAt, &p.Status, &p.MediaCount); err != nil {
						output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
						return nil
					}
					items = append(items, postItem{
						ID:          p.ID,
						Status:      string(p.Status),
						Caption:     p.Caption,
						ScheduledAt: p.ScheduledAt,
						PublishedAt: p.PublishedAt,
						CreatedAt:   p.CreatedAt,
						MediaCount:  p.MediaCount,
					})
				}
			} else {
				posts, err := (*a).DB.GetPosts(limit, 0)
				if err != nil {
					output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
					return nil
				}
				for _, p := range posts {
					items = append(items, postItem{
						ID:          p.ID,
						Status:      string(p.Status),
						Caption:     p.Caption,
						ScheduledAt: p.ScheduledAt,
						PublishedAt: p.PublishedAt,
						CreatedAt:   p.CreatedAt,
						MediaCount:  p.MediaCount,
					})
				}
			}
			if items == nil {
				items = []postItem{}
			}
			output.OK(cmd.OutOrStdout(), map[string]interface{}{"posts": items}, *pretty)
			return nil
		},
	}
	cmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status: draft|scheduled|published|failed")
	cmd.Flags().IntVar(&limit, "limit", 50, "Max results")
	return cmd
}

func newPostPublishCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var id int64
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a post immediately",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id <= 0 {
				output.Err(cmd.OutOrStdout(), "invalid --id: must be > 0", *pretty)
				return nil
			}
			if (*a).Scheduler == nil {
				output.Err(cmd.OutOrStdout(), "scheduler not initialised (is infrastructure running?)", *pretty)
				return nil
			}
			var caption string
			err := (*a).DB.Conn.QueryRow("SELECT caption FROM posts WHERE id=?", id).Scan(&caption)
			if err == sql.ErrNoRows {
				output.Err(cmd.OutOrStdout(), fmt.Sprintf("post %d not found", id), *pretty)
				return nil
			}
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			go (*a).Scheduler.PublishPost(id, caption)
			output.OK(cmd.OutOrStdout(), map[string]interface{}{
				"post_id": id,
				"message": fmt.Sprintf("publish triggered for post %d", id),
			}, *pretty)
			return nil
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "Post ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newPostDeleteCmd(a **appPkg.App, pretty *bool) *cobra.Command {
	var id int64
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a post",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := (*a).DB.Conn.Exec("DELETE FROM posts WHERE id=?", id)
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			rows, err := result.RowsAffected()
			if err != nil {
				output.Err(cmd.OutOrStdout(), err.Error(), *pretty)
				return nil
			}
			if rows == 0 {
				output.Err(cmd.OutOrStdout(), "post not found", *pretty)
				return nil
			}
			output.OK(cmd.OutOrStdout(), map[string]int64{"deleted_id": id}, *pretty)
			return nil
		},
	}
	cmd.Flags().Int64Var(&id, "id", 0, "Post ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}
