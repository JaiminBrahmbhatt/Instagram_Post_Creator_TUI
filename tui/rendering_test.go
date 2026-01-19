package tui

import (
	"testing"
	"strings"
)

// Mock model for testing layout
func TestViewLayout(t *testing.T) {
	// Initialize a minimal model
	m := Model{
		width: 100,
		height: 40,
		currentView: MenuView,
		// We can't easily initialize the full model with DB/Client here without mocking them, 
		// but for View() testing of the shell, we might just check if it calls the render functions.
		// However, Model relies on 'list' which needs initialization.
		// Let's rely on checking the implementation of renderAppShell which we will expose or test directly if possible,
		// or just check if View() output contains our new Header.
	}
	
	// Since initializing the full model is complex due to dependencies,
	// I will focus on testing a new helper function 'renderAppShell' that I plan to create in rendering.go.
	// This function will take the content string and wrap it in the header/footer.
	
	content := "Test Content"
	output := m.renderAppShell(content) // This method doesn't exist yet (Red Phase)
	
	if !strings.Contains(output, "Insta Auto-Post") {
		t.Error("Output should contain the App Title")
	}
	
	// Check for padding/margins implied by the new layout
	// This is hard to test precisely without the implementation, but let's assume 
	// we want a specific structure.
}

// Check that Composer uses the new styles
func TestComposerLayout(t *testing.T) {
	m := Model{
		selectedMedia: []string{"test.jpg"},
		input: NewCaptionInput(),
		width: 80,
		height: 24,
	}
	// We need to verify that viewComposer returns a string containing key elements
	// viewComposer uses 'Theme' so it should work if Theme is initialized (it is var)
	
	output := m.viewComposer()
	
	if !strings.Contains(output, "Compose New Post") {
		t.Error("Composer should have the new header")
	}
	
	if !strings.Contains(output, "1 Files Selected") { // Badge text
		t.Error("Composer should show file count badge")
	}
}

func TestFooterLayout(t *testing.T) {
	m := Model{
		browserDir: "/tmp/test",
		currentView: BrowserView,
		statusMsg: "Test Status",
	}
	// We need to initialize Help because viewFooter calls m.help.View
	// help.New()
	
	output := m.viewFooter()
	
	if !strings.Contains(output, "📍 /tmp/test") {
		t.Error("Footer should show path")
	}
	
	if !strings.Contains(output, "Test Status") {
		t.Error("Footer should show status message")
	}
}
