package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	nethtml "golang.org/x/net/html"
)

// ============================================================
// EPUB Builder (Core Orchestration & Lifecycle)
// ============================================================

type EpubBuilder struct {
	InputPath   string
	OutputPath  string
	CustomCover string
	KeepTemp    bool

	RootDir     string
	BuildDir    string
	TempDir     string
	RawHTMLPath string

	OebpsDir   string
	TextDir    string
	StylesDir  string
	ImagesDir  string
	MetaInfDir string

	Meta          BookMeta
	ManifestItems []ManifestItem
	SpineItems    []SpineItem
	ExtraImages   []ManifestItem
}

func NewEpubBuilder(inputPath, outputPath, customCover string, keepTemp bool) (*EpubBuilder, error) {
	rootDir, _ := filepath.Abs(".")
	if _, err := os.Stat(filepath.Join(rootDir, inputPath)); err != nil {
		if _, err := os.Stat(filepath.Join(rootDir, "..", inputPath)); err == nil {
			rootDir, _ = filepath.Abs("..")
		}
	}
	buildDir := filepath.Join(rootDir, "build")

	b := &EpubBuilder{
		InputPath:   inputPath,
		OutputPath:  outputPath,
		CustomCover: customCover,
		KeepTemp:    keepTemp,
		RootDir:     rootDir,
		BuildDir:    buildDir,
		TempDir:     filepath.Join(buildDir, "_epub_contents"),
		RawHTMLPath: filepath.Join(buildDir, "raw_export.html"),
		OebpsDir:    filepath.Join(buildDir, "_epub_contents", "OEBPS"),
		TextDir:     filepath.Join(buildDir, "_epub_contents", "OEBPS", "text"),
		StylesDir:   filepath.Join(buildDir, "_epub_contents", "OEBPS", "styles"),
		ImagesDir:   filepath.Join(buildDir, "_epub_contents", "OEBPS", "images"),
		MetaInfDir:  filepath.Join(buildDir, "_epub_contents", "META-INF"),
	}

	meta, err := b.extractMetadata()
	if err != nil {
		return nil, err
	}
	b.Meta = meta

	if b.OutputPath == "" {
		safeTitle := sanitizeFilename(b.Meta.Title)
		b.OutputPath = filepath.Join(b.BuildDir, fmt.Sprintf("%s-v%s.epub", safeTitle, b.Meta.Version))
	} else {
		b.OutputPath, _ = filepath.Abs(b.OutputPath)
	}

	return b, nil
}

func (b *EpubBuilder) extractMetadata() (BookMeta, error) {
	queryCmd := exec.Command("typst", "query",
		"--features", "html",
		"--input", "target=epub",
		b.InputPath,
		"<t2e-metadata>",
		"--field", "value",
	)
	queryCmd.Dir = b.RootDir
	out, err := queryCmd.Output()
	if err != nil || len(out) == 0 {
		return BookMeta{}, fmt.Errorf("failed to extract metadata from Typst via `<t2e-metadata>` query: %w. Ensure the document imports template.typ and invokes elegantbook", err)
	}

	var queryList []BookMeta
	if jsonErr := json.Unmarshal(out, &queryList); jsonErr != nil || len(queryList) == 0 {
		return BookMeta{}, fmt.Errorf("invalid or empty JSON metadata returned from Typst query: %w", jsonErr)
	}

	m := queryList[0]
	if m.Title == "" {
		m.Title = "未命名文档"
	}
	if m.Author == "" {
		m.Author = "佚名"
	}
	if m.Version == "" {
		m.Version = "0.1.0"
	}
	if m.Date == "" {
		m.Date = time.Now().Format("2006-01-02")
	}
	if m.Language == "" {
		m.Language = "zh-CN"
	}

	m.Identifier = "urn:typ2epub:" + m.Title + "-" + m.Version
	return m, nil
}

func (b *EpubBuilder) compileTypst() error {
	cmd := exec.Command("typst", "compile",
		"--features", "html",
		"--input", "target=epub",
		b.InputPath,
		b.RawHTMLPath,
	)
	fmt.Printf("[*] Compiling Typst to HTML: %s\n", strings.Join(cmd.Args, " "))
	cmd.Dir = b.RootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[!] Typst compilation failed:\n%s\n", string(output))
		return err
	}
	fmt.Println("[+] Typst HTML compiled successfully.")
	return nil
}

func (b *EpubBuilder) prepareCSS() error {
	cssSrc := filepath.Join(b.RootDir, "scripts", "epub-style.css")
	cssDst := filepath.Join(b.StylesDir, "style.css")

	data, err := os.ReadFile(cssSrc)
	if err != nil {
		return err
	}
	return os.WriteFile(cssDst, data, 0644)
}

func (b *EpubBuilder) writeChapters(chapters []*Chapter) ([]*TocNode, error) {
	var tocRoots []*TocNode

	for idx, chap := range chapters {
		var bodySB strings.Builder
		bodySB.WriteString("  <div class=\"chapter-body\">\n")
		hasMathML := false
		hasSVG := false

		for _, node := range chap.Nodes {
			if !hasMathML || !hasSVG {
				_ = findFirst(node, func(n *nethtml.Node) bool {
					if n.Type == nethtml.ElementNode {
						if n.Data == "math" {
							hasMathML = true
						} else if n.Data == "svg" {
							hasSVG = true
						}
					}
					return hasMathML && hasSVG // Stop early only if both found
				})
			}
			bodySB.WriteString(renderNodeToString(node))
		}
		bodySB.WriteString("\n  </div>")

		var props []string
		if hasMathML {
			props = append(props, "mathml")
		}
		if hasSVG {
			props = append(props, "svg")
		}

		chapHTML := makeXHTMLPage(chap.Title, bodySB.String(), "../styles/style.css", b.Meta.Language)
		if err := os.WriteFile(filepath.Join(b.TextDir, chap.Filename), []byte(chapHTML), 0644); err != nil {
			return nil, err
		}

		itemID := fmt.Sprintf("chap_%d", idx+1)
		b.ManifestItems = append(b.ManifestItems, ManifestItem{
			ID:         itemID,
			Href:       "text/" + chap.Filename,
			MediaType:  "application/xhtml+xml",
			Properties: strings.Join(props, " "),
		})
		b.SpineItems = append(b.SpineItems, SpineItem{IDRef: itemID})

		rootNode := buildChapterTocTree(chap, idx)
		tocRoots = append(tocRoots, rootNode)
	}

	return tocRoots, nil
}

func (b *EpubBuilder) Build() error {
	_ = os.RemoveAll(b.TempDir)
	for _, dir := range []string{b.TextDir, b.StylesDir, b.ImagesDir, b.MetaInfDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	defer func() {
		if !b.KeepTemp {
			_ = os.RemoveAll(b.TempDir)
			_ = os.Remove(b.RawHTMLPath)
		}
	}()

	// 1. Compile Typst to HTML
	if err := b.compileTypst(); err != nil {
		return err
	}

	// 2. Prepare CSS
	if err := b.prepareCSS(); err != nil {
		return err
	}

	// 3. Prepare Cover
	coverFileName, coverMediaType, err := b.prepareCover()
	if err != nil {
		return err
	}
	coverXHTMLName := "cover.xhtml"

	// 4. Parse HTML and Post-process
	rawHTMLData, err := os.ReadFile(b.RawHTMLPath)
	if err != nil {
		return err
	}
	doc, err := nethtml.Parse(bytes.NewReader(rawHTMLData))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	body := findFirst(doc, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "body"
	})
	if body == nil {
		body = doc
	}

	// 4a. Extract Footnotes (ordered)
	footnoteStore := extractFootnotes(body)

	// 4b. Remove Typst Static TOC
	removeTypstTOC(body)

	// 4c. Split Chapters
	chapters := splitChapters(body)
	if len(chapters) == 0 {
		return fmt.Errorf("no chapters or content found in document")
	}

	// 4d. Assign Footnotes to Chapters in document reference order
	assignFootnotes(chapters, footnoteStore)

	// 4e. Collect Images
	if err := b.collectImages(chapters); err != nil {
		return fmt.Errorf("collect images error: %w", err)
	}

	// 4f. Rewrite Cross References
	rewriteCrossReferences(chapters)

	// 5. Initialize Manifest and Spine
	b.ManifestItems = []ManifestItem{
		{ID: "style", Href: "styles/style.css", MediaType: "text/css"},
		{ID: "cover_img", Href: "images/" + coverFileName, MediaType: coverMediaType, Properties: "cover-image"},
		{ID: "cover_page", Href: "text/" + coverXHTMLName, MediaType: "application/xhtml+xml"},
		{ID: "nav", Href: "text/nav.xhtml", MediaType: "application/xhtml+xml", Properties: "nav"},
	}
	b.ManifestItems = append(b.ManifestItems, b.ExtraImages...)

	b.SpineItems = []SpineItem{
		{IDRef: "cover_page", Linear: "no"},
		{IDRef: "nav"},
	}

	// 6. Write Chapters and Build TOC
	tocRoots, err := b.writeChapters(chapters)
	if err != nil {
		return err
	}

	bodymatterFilename := chapters[0].Filename
	for _, chap := range chapters {
		if chap.Matter == "main" {
			bodymatterFilename = chap.Filename
			break
		}
	}

	// 7. Navigation
	if err := b.generateNavXHTML(tocRoots, bodymatterFilename); err != nil {
		return err
	}

	// 8. Generate content.opf
	if err := b.generateContentOPF(coverFileName, coverMediaType, coverXHTMLName, bodymatterFilename); err != nil {
		return err
	}

	// 9. Write Container Files
	if err := b.writeContainerFiles(); err != nil {
		return err
	}

	// 10. Package EPUB (ZIP)
	if err := b.packageEPUB(); err != nil {
		return err
	}

	fmt.Printf("[✓] EPUB build successful: %s\n", b.OutputPath)
	fmt.Printf("    - Title:   %s (%s)\n", b.Meta.Title, b.Meta.Subtitle)
	fmt.Printf("    - Author:  %s\n", b.Meta.Author)
	fmt.Printf("    - Version: %s\n", b.Meta.Version)
	fmt.Printf("    - Chapters: %d\n", len(chapters))

	return nil
}
