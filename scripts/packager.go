package main

import (
	"archive/zip"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================
// EPUB Packaging & OCF
// ============================================================

func (b *EpubBuilder) generateContentOPF(coverFileName, coverMediaType, coverXHTMLName, bodymatterFilename string) error {
	modifiedISO := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	var manifestLines []string
	for _, item := range b.ManifestItems {
		propAttr := ""
		if item.Properties != "" {
			propAttr = fmt.Sprintf(` properties="%s"`, html.EscapeString(item.Properties))
		}
		manifestLines = append(manifestLines,
			fmt.Sprintf(`    <item id="%s" href="%s" media-type="%s"%s/>`,
				html.EscapeString(item.ID),
				html.EscapeString(item.Href),
				html.EscapeString(item.MediaType),
				propAttr,
			),
		)
	}

	var spineLines []string
	for _, item := range b.SpineItems {
		linearAttr := ""
		if item.Linear != "" {
			linearAttr = fmt.Sprintf(` linear="%s"`, html.EscapeString(item.Linear))
		}
		spineLines = append(spineLines,
			fmt.Sprintf(`    <itemref idref="%s"%s/>`,
				html.EscapeString(item.IDRef),
				linearAttr,
			),
		)
	}

	opfContent := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookId" version="3.0" xml:lang="%s">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:identifier id="BookId">%s</dc:identifier>
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:language>%s</dc:language>
    <dc:date>%s</dc:date>
    <meta property="dcterms:modified">%s</meta>
    <meta name="cover" content="cover_img" />
  </metadata>
  <manifest>
%s
  </manifest>
  <spine page-progression-direction="ltr">
%s
  </spine>
  <guide>
    <reference type="cover" title="封面" href="text/%s"/>
    <reference type="toc" title="目录" href="text/nav.xhtml"/>
    <reference type="text" title="正文" href="text/%s"/>
  </guide>
</package>`,
		b.Meta.Language,
		b.Meta.Identifier,
		html.EscapeString(b.Meta.Title),
		html.EscapeString(b.Meta.Author),
		b.Meta.Language,
		b.Meta.Date,
		modifiedISO,
		strings.Join(manifestLines, "\n"),
		strings.Join(spineLines, "\n"),
		coverXHTMLName,
		bodymatterFilename,
	)

	return os.WriteFile(filepath.Join(b.OebpsDir, "content.opf"), []byte(opfContent), 0644)
}

func (b *EpubBuilder) writeContainerFiles() error {
	if err := os.WriteFile(filepath.Join(b.TempDir, "mimetype"), []byte("application/epub+zip"), 0644); err != nil {
		return err
	}

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	return os.WriteFile(filepath.Join(b.MetaInfDir, "container.xml"), []byte(containerXML), 0644)
}

func (b *EpubBuilder) packageEPUB() error {
	fmt.Printf("[*] Packaging EPUB to: %s\n", b.OutputPath)
	tmpPath := b.OutputPath + ".tmp"
	_ = os.Remove(tmpPath)

	zipFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary output file %q: %w", tmpPath, err)
	}

	success := false
	defer func() {
		if !success {
			_ = zipFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	w := zip.NewWriter(zipFile)

	// mimetype MUST be first, STORED (uncompressed), and have ZERO extra field bytes (EPUB 3.3 OCF)
	mimetypeData, err := os.ReadFile(filepath.Join(b.TempDir, "mimetype"))
	if err != nil {
		return err
	}
	mimeHeader := &zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	}
	mimeHeader.Extra = nil
	mimeWriter, err := w.CreateHeader(mimeHeader)
	if err != nil {
		return err
	}
	if _, err := mimeWriter.Write(mimetypeData); err != nil {
		return err
	}

	// Add remaining files with Deflate compression
	var allFiles []string
	err = filepath.Walk(b.TempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(b.TempDir, path)
			if rel != "mimetype" {
				allFiles = append(allFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, filePath := range allFiles {
		relPath, _ := filepath.Rel(b.TempDir, filePath)
		relPath = filepath.ToSlash(relPath)

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		header := &zip.FileHeader{
			Name:   relPath,
			Method: zip.Deflate,
		}
		fileWriter, err := w.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := fileWriter.Write(fileData); err != nil {
			return err
		}
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to finalize EPUB zip: %w", err)
	}

	if err := zipFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err := os.Rename(tmpPath, b.OutputPath); err != nil {
		return fmt.Errorf("failed to move EPUB to target output path %q: %w", b.OutputPath, err)
	}

	success = true
	return nil
}
