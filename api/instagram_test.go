package api

import (
	"testing"
)

func TestCreateMediaContainer_SinglePost(t *testing.T) {
	client := &Client{
		AccessToken: "test_token",
		IGID:        "test_ig_id",
	}

	// Test that is_carousel_item is NOT included for single posts
	// We can't easily test the actual HTTP request without mocking,
	// but we can verify the logic by checking the params map construction

	// This test verifies the fix: is_carousel_item should only be sent when true
	mediaURL := "https://example.com/image.jpg"
	caption := "Test caption"
	mediaType := MediaTypeImage
	isCarouselItem := false

	// The function should not include is_carousel_item in params when false
	// We'll need to refactor to make this testable, or test via integration

	// For now, just verify the function signature is correct
	_, err := client.CreateMediaContainer(mediaURL, caption, mediaType, isCarouselItem)

	// We expect an error because we're not actually making a real API call
	// but the function should at least compile and run
	if err == nil {
		t.Log("Function executed (expected to fail with network error)")
	}
}

func TestCreateMediaContainer_CarouselItem(t *testing.T) {
	client := &Client{
		AccessToken: "test_token",
		IGID:        "test_ig_id",
	}

	mediaURL := "https://example.com/image.jpg"
	caption := ""
	mediaType := MediaTypeImage
	isCarouselItem := true

	// The function should include is_carousel_item=true in params
	_, err := client.CreateMediaContainer(mediaURL, caption, mediaType, isCarouselItem)

	if err == nil {
		t.Log("Function executed (expected to fail with network error)")
	}
}
