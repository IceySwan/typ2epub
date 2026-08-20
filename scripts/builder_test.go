package main

import (
	"archive/zip"
	"io"
	"path/filepath"
	"regexp"
	"testing"
)

func TestEpubBuilderEndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	out := filepath.Join(tempDir, "test.epub")

	// No cover provided: the build must fall back to the generated SVG cover.
	b, err := NewEpubBuilder("main.typ", out, "", false)
	if err != nil {
		t.Fatalf("builder init failed: %v", err)
	}
	if err := b.Build(); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Verify ZIP structure and mimetype EPUB 3.3 OCF strict requirements
	r, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("failed to open generated epub zip: %v", err)
	}
	defer r.Close()

	if len(r.File) == 0 {
		t.Fatalf("epub zip archive is empty")
	}

	firstFile := r.File[0]
	if firstFile.Name != "mimetype" {
		t.Errorf("expected first zip entry to be 'mimetype', got %q", firstFile.Name)
	}
	if firstFile.Method != zip.Store {
		t.Errorf("expected 'mimetype' to be STORED (method 0), got method %d", firstFile.Method)
	}
	if len(firstFile.Extra) != 0 {
		t.Errorf("expected 'mimetype' ZIP header Extra field length to be 0 for EPUB 3.3 OCF compliance, got %d bytes", len(firstFile.Extra))
	}

	rc, err := firstFile.Open()
	if err != nil {
		t.Fatalf("failed to read mimetype file: %v", err)
	}
	mimeBytes, _ := io.ReadAll(rc)
	rc.Close()
	if string(mimeBytes) != "application/epub+zip" {
		t.Errorf("expected mimetype content 'application/epub+zip', got %q", string(mimeBytes))
	}

	// Verify required EPUB3 container files exist
	expectedFiles := map[string]bool{
		"mimetype":               false,
		"META-INF/container.xml": false,
		"OEBPS/content.opf":      false,
		"OEBPS/text/nav.xhtml":   false,
	}

	var opfContent string
	for _, f := range r.File {
		if _, ok := expectedFiles[f.Name]; ok {
			expectedFiles[f.Name] = true
		}
		if f.Name == "OEBPS/content.opf" {
			rc, _ := f.Open()
			bytes, _ := io.ReadAll(rc)
			rc.Close()
			opfContent = string(bytes)
		}
	}

	for name, found := range expectedFiles {
		if !found {
			t.Errorf("required entry %q missing from generated EPUB", name)
		}
	}

	// Verify Manifest Item ID and Href Uniqueness in content.opf
	reItem := regexp.MustCompile(`<item\s+[^>]*id="([^"]+)"[^>]*href="([^"]+)"`)
	matches := reItem.FindAllStringSubmatch(opfContent, -1)
	seenIDs := make(map[string]bool)
	seenHrefs := make(map[string]bool)

	for _, m := range matches {
		id, href := m[1], m[2]
		if seenIDs[id] {
			t.Errorf("duplicate manifest item id %q in content.opf", id)
		}
		seenIDs[id] = true

		if seenHrefs[href] {
			t.Errorf("duplicate manifest item href %q in content.opf", href)
		}
		seenHrefs[href] = true
	}
}
