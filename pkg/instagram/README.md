# Instagram Content Publishing SDK

A production-ready Go SDK for the Instagram Content Publishing API that handles all the edge cases and quirks of the Instagram Graph API.

## Features

✅ **Simple API** - Clean, intuitive interface  
✅ **Edge Case Handling** - Automatically handles Instagram API quirks  
✅ **Type Safety** - Strong typing for all parameters  
✅ **Validation** - Built-in validation for all inputs  
✅ **Error Handling** - Comprehensive error messages  
✅ **Context Support** - Proper context handling for cancellation  
✅ **Production Ready** - Battle-tested against real Instagram API

## Installation

```bash
go get github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/pkg/instagram
```

## Quick Start

### Single Image Post

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/pkg/instagram"
)

func main() {
    // Create SDK instance
    sdk, err := instagram.NewSDK(instagram.Config{
        AccessToken: "your_access_token",
        IGUserID:    "your_instagram_user_id",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Publish a single image
    post := instagram.Post{
        ImageURL: "https://example.com/image.jpg",
        Caption:  "Check out this amazing photo! #instagram",
        Type:     instagram.MediaTypeImage,
    }

    mediaID, err := sdk.PublishSinglePost(context.Background(), post)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("✅ Published! Media ID: %s\n", mediaID)
}
```

### Carousel Post

```go
// Create a carousel with multiple images
carousel := instagram.CarouselPost{
    Caption: "Swipe to see more! 👉",
    Items: []instagram.Post{
        {
            ImageURL: "https://example.com/image1.jpg",
            Type:     instagram.MediaTypeImage,
        },
        {
            ImageURL: "https://example.com/image2.jpg",
            Type:     instagram.MediaTypeImage,
        },
        {
            ImageURL: "https://example.com/image3.jpg",
            Type:     instagram.MediaTypeImage,
        },
    },
}

mediaID, err := sdk.PublishCarousel(context.Background(), carousel)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("✅ Carousel published! Media ID: %s\n", mediaID)
```

### Video Post

```go
post := instagram.Post{
    VideoURL: "https://example.com/video.mp4",
    Caption:  "Amazing video content!",
    Type:     instagram.MediaTypeVideo,
}

mediaID, err := sdk.PublishSinglePost(context.Background(), post)
```

### Reels

```go
post := instagram.Post{
    VideoURL: "https://example.com/reel.mp4",
    Caption:  "New Reel! 🎬",
    Type:     instagram.MediaTypeReels,
}

mediaID, err := sdk.PublishSinglePost(context.Background(), post)
```

## Configuration

### SDK Config

```go
config := instagram.Config{
    AccessToken: "your_access_token",  // Required
    IGUserID:    "your_ig_user_id",    // Required
    APIVersion:  "v19.0",               // Optional (default: v19.0)
    HTTPTimeout: 30 * time.Second,     // Optional (default: 30s)
}

sdk, err := instagram.NewSDK(config)
```

### Getting Access Token

1. Create a Facebook App
2. Add Instagram Basic Display or Instagram Graph API
3. Generate a User Access Token with `instagram_content_publish` permission
4. Get your Instagram User ID

See: [Instagram Content Publishing Documentation](https://developers.facebook.com/docs/instagram-platform/content-publishing/)

## API Reference

### Types

#### `MediaType`

```go
const (
    MediaTypeImage   MediaType = "IMAGE"
    MediaTypeVideo   MediaType = "VIDEO"
    MediaTypeReels   MediaType = "REELS"
    MediaTypeStories MediaType = "STORIES"
)
```

#### `Post`

```go
type Post struct {
    ImageURL string    // URL to the image (must be publicly accessible)
    VideoURL string    // URL to the video (must be publicly accessible)
    Caption  string    // Post caption
    Type     MediaType // Type of media
}
```

#### `CarouselPost`

```go
type CarouselPost struct {
    Items   []Post // 2-10 items
    Caption string // Caption for the carousel
}
```

### Methods

#### `PublishSinglePost(ctx context.Context, post Post) (string, error)`

Publishes a single image or video post.

**Parameters:**

- `ctx` - Context for cancellation
- `post` - Post configuration

**Returns:**

- `string` - Published media ID
- `error` - Error if any

**Example:**

```go
mediaID, err := sdk.PublishSinglePost(ctx, post)
```

#### `PublishCarousel(ctx context.Context, carousel CarouselPost) (string, error)`

Publishes a carousel post with 2-10 items.

**Parameters:**

- `ctx` - Context for cancellation
- `carousel` - Carousel configuration

**Returns:**

- `string` - Published carousel media ID
- `error` - Error if any

**Example:**

```go
mediaID, err := sdk.PublishCarousel(ctx, carousel)
```

## Edge Cases Handled

### 1. Media Type Parameter

❌ **Wrong:** Sending `media_type: IMAGE` for images  
✅ **Correct:** SDK automatically omits `media_type` for images

The Instagram API only accepts `media_type` for VIDEO, REELS, or STORIES. For images, the parameter must be omitted entirely.

### 2. Carousel Item Flag

❌ **Wrong:** Sending `is_carousel_item: false`  
✅ **Correct:** SDK only sends `is_carousel_item: true` when needed

The Instagram API rejects requests with `is_carousel_item=false`. The SDK only includes this parameter when it's `true`.

### 3. Container Processing

The SDK automatically:

- Waits for Instagram to process media containers
- Polls container status every 5 seconds
- Times out after 5 minutes
- Handles all container statuses (FINISHED, IN_PROGRESS, ERROR, EXPIRED)

### 4. URL Validation

The SDK validates:

- URLs must be HTTP(S)
- Images must be .jpg, .jpeg, or .png
- Videos must be .mp4 or .mov
- URLs must be publicly accessible

### 5. Carousel Validation

The SDK ensures:

- Minimum 2 items
- Maximum 10 items
- All items are valid
- Proper item ordering

## Error Handling

The SDK provides detailed error messages:

```go
mediaID, err := sdk.PublishSinglePost(ctx, post)
if err != nil {
    // Error messages are descriptive and actionable
    // Examples:
    // - "invalid post: image_url is required for images"
    // - "failed to create container: API error: Invalid media URL (code: 100)"
    // - "container processing failed: timeout waiting for container 12345"
    log.Printf("Failed to publish: %v", err)
    return
}
```

## Best Practices

### 1. Use Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

mediaID, err := sdk.PublishSinglePost(ctx, post)
```

### 2. Validate Media URLs

Ensure your media URLs are:

- Publicly accessible (no authentication required)
- Direct links to the media file
- Hosted on a reliable server
- Using HTTPS (recommended)

### 3. Handle Rate Limits

Instagram has rate limits for content publishing:

- 25 posts per day per user
- 5 posts per hour per user

```go
// Implement retry logic with exponential backoff
for i := 0; i < 3; i++ {
    mediaID, err := sdk.PublishSinglePost(ctx, post)
    if err == nil {
        break
    }

    if strings.Contains(err.Error(), "rate limit") {
        time.Sleep(time.Duration(i+1) * time.Minute)
        continue
    }

    return err
}
```

### 4. Use Ngrok for Local Testing

When testing locally, use ngrok to expose your local server:

```bash
ngrok http 8080
```

Then use the ngrok URL in your image/video URLs.

## Troubleshooting

### "Only photo or video can be accepted as media type"

This error occurs when:

- The media URL is not publicly accessible
- The URL doesn't point directly to the media file
- The media file format is not supported

**Solution:** Ensure your URL is publicly accessible and points directly to the file.

### "Container processing failed: timeout"

This occurs when Instagram takes too long to process the media.

**Solution:**

- Check if the media file is too large
- Ensure the URL is fast and reliable
- Try with a smaller file

### "Invalid carousel: carousel must have at least 2 items"

**Solution:** Carousels require 2-10 items. Use `PublishSinglePost` for single items.

## Testing

```bash
# Run tests
go test ./pkg/instagram/...

# Run with coverage
go test -cover ./pkg/instagram/...
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

- 📖 [Instagram Content Publishing Docs](https://developers.facebook.com/docs/instagram-platform/content-publishing/)
- 🐛 [Report Issues](https://github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/issues)
- 💬 [Discussions](https://github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/discussions)

## Changelog

### v1.0.0 (2026-01-25)

- ✨ Initial release
- ✅ Single post publishing
- ✅ Carousel post publishing
- ✅ Video and Reels support
- ✅ Comprehensive validation
- ✅ Edge case handling
- ✅ Context support
- ✅ Detailed error messages
