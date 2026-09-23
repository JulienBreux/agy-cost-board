package web_test

import (
	"io"
	"strings"
	"testing"

	"github.com/julienbreux/agy-ge-board/web"
)

func TestEmbeddedWebFS(t *testing.T) {
	staticFS, err := web.FS()
	if err != nil {
		t.Fatalf("unexpected error loading embedded FS: %v", err)
	}

	indexFile, err := staticFS.Open("index.html")
	if err != nil {
		t.Fatalf("failed to open embedded index.html: %v", err)
	}
	defer indexFile.Close()

	content, err := io.ReadAll(indexFile)
	if err != nil {
		t.Fatalf("failed to read embedded index.html: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "AGY & Gemini Enterprise") && !strings.Contains(contentStr, "<div id=\"root\"></div>") {
		t.Errorf("embedded index.html does not contain expected HTML template content: %s", contentStr)
	}
}
