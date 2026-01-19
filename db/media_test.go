package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsPathSafe(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-photos")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize test database
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Conn.Close()

	// Set photos_dir
	if err := db.SetSetting("photos_dir", tmpDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		path     string
		wantSafe bool
	}{
		{
			name:     "valid path inside photos_dir",
			path:     filepath.Join(tmpDir, "photo.jpg"),
			wantSafe: true,
		},
		{
			name:     "valid subdirectory path",
			path:     filepath.Join(tmpDir, "subdir", "photo.jpg"),
			wantSafe: true,
		},
		{
			name:     "path traversal attempt with ..",
			path:     filepath.Join(tmpDir, "..", "etc", "passwd"),
			wantSafe: false,
		},
		{
			name:     "absolute path outside photos_dir",
			path:     "/etc/passwd",
			wantSafe: false,
		},
		{
			name:     "relative path that escapes",
			path:     filepath.Join(tmpDir, "..", "..", "etc", "passwd"),
			wantSafe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe, err := db.isPathSafe(tt.path)
			if err != nil && tt.wantSafe {
				t.Errorf("isPathSafe() unexpected error = %v", err)
				return
			}
			if safe != tt.wantSafe {
				t.Errorf("isPathSafe() = %v, want %v for path %s", safe, tt.wantSafe, tt.path)
			}
		})
	}
}

func TestRegisterMediaPathSafety(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-photos")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Conn.Close()

	if err := db.SetSetting("photos_dir", tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test valid registration
	err = db.RegisterMedia(testFile)
	if err != nil {
		t.Errorf("RegisterMedia() failed for valid path: %v", err)
	}

	// Test invalid registration (path outside photos_dir)
	err = db.RegisterMedia("/etc/passwd")
	if err == nil {
		t.Error("RegisterMedia() should fail for path outside photos_dir")
	}
}
