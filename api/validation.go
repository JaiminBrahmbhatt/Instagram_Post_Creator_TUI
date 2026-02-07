// Package api provides validation functions per implementation_plan.md (Phase 1–4).
package api

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateLocationID checks that location_id is a numeric string (Facebook Page ID format).
// Per implementation_plan.md §1.2: "Validate format (numeric string)".
func ValidateLocationID(locationID string) error {
	if locationID == "" {
		return nil
	}
	s := strings.TrimSpace(locationID)
	if s == "" {
		return nil
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return fmt.Errorf("location_id must be numeric (Facebook Page ID), got non-digit")
		}
	}
	return nil
}

// ValidateUsername checks Instagram username format: alphanumeric, underscores, periods only.
// Per implementation_plan.md §2.2: "Username format validation (alphanumeric, underscores, periods)".
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._]+$`)

func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("invalid username format: only letters, numbers, dots and underscores allowed")
	}
	return nil
}

// ValidateUserTags validates each username in the list.
func ValidateUserTags(usernames []string) error {
	for i, u := range usernames {
		if err := ValidateUsername(strings.TrimSpace(u)); err != nil {
			return fmt.Errorf("user tag #%d (%q): %w", i+1, u, err)
		}
	}
	return nil
}

// MaxAltTextLen is the maximum allowed length for alt_text (implementation_plan.md §3.1).
const MaxAltTextLen = 300

// ValidateAltText checks length and that alt_text is only used for image (isImage=true).
// Per implementation_plan.md §3.1: "alt_text only supported for IMAGE media type", "Character limit (example: 300 characters)".
func ValidateAltText(altText string, isImage bool) error {
	if altText == "" {
		return nil
	}
	if !isImage {
		return fmt.Errorf("alt_text is only supported for image posts")
	}
	if len(altText) > MaxAltTextLen {
		return fmt.Errorf("alt_text exceeds maximum length of %d characters", MaxAltTextLen)
	}
	return nil
}

// IsVideoPath returns true if the file extension indicates video (.mp4, .mov).
// Used to enforce "alt_text only for images" and "Story = single image or video" when validating from paths.
func IsVideoPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".mp4" || ext == ".mov"
}
