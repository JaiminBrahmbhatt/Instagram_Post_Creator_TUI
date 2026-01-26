package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/pkg/instagram"
)

func main() {
	// Initialize the SDK
	sdk, err := instagram.NewSDK(instagram.Config{
		AccessToken: "YOUR_ACCESS_TOKEN",
		IGUserID:    "YOUR_IG_USER_ID",
		APIVersion:  "v19.0",
		HTTPTimeout: 30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create SDK: %v", err)
	}

	// Example 1: Publish a single image
	fmt.Println("📸 Publishing single image...")
	singleImageExample(sdk)

	// Example 2: Publish a carousel
	fmt.Println("\n🎠 Publishing carousel...")
	carouselExample(sdk)

	// Example 3: Publish a video
	fmt.Println("\n🎥 Publishing video...")
	videoExample(sdk)

	// Example 4: Publish a Reel
	fmt.Println("\n🎬 Publishing Reel...")
	reelExample(sdk)
}

func singleImageExample(sdk *instagram.SDK) {
	ctx := context.Background()

	post := instagram.Post{
		ImageURL: "https://example.com/beautiful-sunset.jpg",
		Caption:  "Amazing sunset today! 🌅 #sunset #nature #photography",
		Type:     instagram.MediaTypeImage,
	}

	mediaID, err := sdk.PublishSinglePost(ctx, post)
	if err != nil {
		log.Printf("❌ Failed to publish image: %v", err)
		return
	}

	fmt.Printf("✅ Image published successfully! Media ID: %s\n", mediaID)
}

func carouselExample(sdk *instagram.SDK) {
	ctx := context.Background()

	carousel := instagram.CarouselPost{
		Caption: "Swipe to see my weekend adventures! 👉 #weekend #travel #adventure",
		Items: []instagram.Post{
			{
				ImageURL: "https://example.com/photo1.jpg",
				Type:     instagram.MediaTypeImage,
			},
			{
				ImageURL: "https://example.com/photo2.jpg",
				Type:     instagram.MediaTypeImage,
			},
			{
				ImageURL: "https://example.com/photo3.jpg",
				Type:     instagram.MediaTypeImage,
			},
			{
				ImageURL: "https://example.com/photo4.jpg",
				Type:     instagram.MediaTypeImage,
			},
		},
	}

	mediaID, err := sdk.PublishCarousel(ctx, carousel)
	if err != nil {
		log.Printf("❌ Failed to publish carousel: %v", err)
		return
	}

	fmt.Printf("✅ Carousel published successfully! Media ID: %s\n", mediaID)
}

func videoExample(sdk *instagram.SDK) {
	ctx := context.Background()

	post := instagram.Post{
		VideoURL: "https://example.com/my-video.mp4",
		Caption:  "Check out this amazing video! 🎥 #video #content",
		Type:     instagram.MediaTypeVideo,
	}

	mediaID, err := sdk.PublishSinglePost(ctx, post)
	if err != nil {
		log.Printf("❌ Failed to publish video: %v", err)
		return
	}

	fmt.Printf("✅ Video published successfully! Media ID: %s\n", mediaID)
}

func reelExample(sdk *instagram.SDK) {
	ctx := context.Background()

	post := instagram.Post{
		VideoURL: "https://example.com/my-reel.mp4",
		Caption:  "New Reel alert! 🎬 #reels #trending",
		Type:     instagram.MediaTypeReels,
	}

	mediaID, err := sdk.PublishSinglePost(ctx, post)
	if err != nil {
		log.Printf("❌ Failed to publish Reel: %v", err)
		return
	}

	fmt.Printf("✅ Reel published successfully! Media ID: %s\n", mediaID)
}
