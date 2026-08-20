package main

import (
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"strings"

	nethtml "golang.org/x/net/html"
)

// ============================================================
// Media & Cover Handling
// ============================================================

func guessMediaType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

// generateSimpleCover writes a plain white-background, black-text SVG cover
// carrying the book title, subtitle, author and version.
func (b *EpubBuilder) generateSimpleCover(outputPath string) error {
	title := html.EscapeString(b.Meta.Title)
	subtitle := html.EscapeString(b.Meta.Subtitle)
	author := html.EscapeString(b.Meta.Author)
	version := html.EscapeString(b.Meta.Version)

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 1200" width="800" height="1200">
  <rect width="800" height="1200" fill="#ffffff"/>
  <rect x="40" y="40" width="720" height="1120" fill="none" stroke="#000000" stroke-width="1.5"/>
  <text x="400" y="520" font-family="sans-serif" font-size="64" font-weight="bold" fill="#000000" text-anchor="middle">%s</text>
  <text x="400" y="600" font-family="sans-serif" font-size="28" fill="#000000" text-anchor="middle">%s</text>
  <line x1="300" y1="640" x2="500" y2="640" stroke="#000000" stroke-width="2"/>
  <text x="400" y="960" font-family="sans-serif" font-size="28" fill="#000000" text-anchor="middle">%s</text>
  <text x="400" y="1010" font-family="sans-serif" font-size="20" fill="#333333" text-anchor="middle">%s</text>
</svg>
`, title, subtitle, author, version)

	return os.WriteFile(outputPath, []byte(svg), 0644)
}

// prepareCover provides the cover image for the EPUB. Resolution order:
// -c flag, then assets/cover.png, cover.jpg or cover.pdf; if none exists,
// a plain generated SVG cover is used.
func (b *EpubBuilder) prepareCover() (string, string, error) {
	coverSrc := ""
	mediaType := ""

	if b.CustomCover != "" {
		coverSrc = b.CustomCover
		mediaType = guessMediaType(b.CustomCover)
	} else {
		for _, cand := range []string{"cover.png", "cover.jpg", "cover.pdf"} {
			p := filepath.Join(b.RootDir, "assets", cand)
			if _, err := os.Stat(p); err == nil {
				coverSrc = p
				mediaType = guessMediaType(p)
				break
			}
		}
	}

	coverFileName := "cover.svg"
	if coverSrc != "" {
		coverFileName = "cover" + strings.ToLower(filepath.Ext(coverSrc))
		if err := copyFile(coverSrc, filepath.Join(b.ImagesDir, coverFileName)); err != nil {
			return "", "", fmt.Errorf("failed to copy cover %q: %w", coverSrc, err)
		}
	} else {
		mediaType = "image/svg+xml"
		if err := b.generateSimpleCover(filepath.Join(b.ImagesDir, coverFileName)); err != nil {
			return "", "", fmt.Errorf("failed to generate default cover: %w", err)
		}
	}

	coverHTML := fmt.Sprintf(`  <div class="cover-page">
    <img src="../images/%s" alt="Cover" class="cover-svg" />
  </div>`, coverFileName)

	err := os.WriteFile(filepath.Join(b.TextDir, "cover.xhtml"),
		[]byte(makeXHTMLPage("封面 - "+b.Meta.Title, coverHTML, "../styles/style.css", b.Meta.Language)), 0644)

	return coverFileName, mediaType, err
}

// collectImages copies local images referenced by <img> into OEBPS/images/
// with sequential names (img_0001.png, img_0002.png, ...) and rewrites src.
func (b *EpubBuilder) collectImages(chapters []*Chapter) error {
	imgIdx := 1
	for _, chap := range chapters {
		for _, node := range chap.Nodes {
			imgTags := findAll(node, func(n *nethtml.Node) bool {
				return n.Type == nethtml.ElementNode && n.Data == "img" && hasAttr(n, "src")
			})
			for _, img := range imgTags {
				src := getAttr(img, "src")
				if strings.HasPrefix(src, "data:") {
					continue
				}

				localPath := filepath.Clean(src)
				if !filepath.IsAbs(localPath) {
					localPath = filepath.Join(b.RootDir, localPath)
				}

				if _, err := os.Stat(localPath); err != nil {
					return fmt.Errorf("image file %q referenced in %s not found: %w", src, chap.Filename, err)
				}

				ext := strings.ToLower(filepath.Ext(localPath))
				destName := fmt.Sprintf("img_%04d%s", imgIdx, ext)
				imgIdx++

				data, err := os.ReadFile(localPath)
				if err != nil {
					return fmt.Errorf("failed to read image %q (%s): %w", src, localPath, err)
				}
				if writeErr := os.WriteFile(filepath.Join(b.ImagesDir, destName), data, 0644); writeErr != nil {
					return fmt.Errorf("failed to write image %q to %s: %w", src, destName, writeErr)
				}

				b.ExtraImages = append(b.ExtraImages, ManifestItem{
					ID:        fmt.Sprintf("img_%04d", imgIdx-1),
					Href:      "images/" + destName,
					MediaType: guessMediaType(localPath),
				})

				setAttr(img, "src", "../images/"+destName)
			}
		}
	}
	return nil
}
