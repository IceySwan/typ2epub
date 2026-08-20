package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================
// EPUB Navigation (EPUB 3 Nav XHTML)
// ============================================================

func makeXHTMLPage(title, bodyHTML string, cssRelPath string, lang string) string {
	if cssRelPath == "" {
		cssRelPath = "../styles/style.css"
	}
	if lang == "" {
		lang = "zh-CN"
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="%s" xml:lang="%s">
<head>
  <meta charset="utf-8" />
  <title>%s</title>
  <link rel="stylesheet" type="text/css" href="%s" />
</head>
<body>
%s
</body>
</html>`, lang, lang, html.EscapeString(title), cssRelPath, bodyHTML)
}

func renderNavOL(items []*TocNode, indent int) string {
	pad := strings.Repeat(" ", indent)
	var lines []string
	lines = append(lines, pad+"<ol>")
	for _, item := range items {
		titleEsc := html.EscapeString(item.Title)
		if len(item.Children) > 0 {
			lines = append(lines, pad+"  <li>")
			lines = append(lines, fmt.Sprintf(`%s    <a href="%s">%s</a>`, pad, item.Href, titleEsc))
			lines = append(lines, renderNavOL(item.Children, indent+4))
			lines = append(lines, pad+"  </li>")
		} else {
			lines = append(lines, fmt.Sprintf(`%s  <li><a href="%s">%s</a></li>`, pad, item.Href, titleEsc))
		}
	}
	lines = append(lines, pad+"</ol>")
	return strings.Join(lines, "\n")
}

func (b *EpubBuilder) generateNavXHTML(tocRoots []*TocNode, bodymatterFilename string) error {
	navOL := renderNavOL(tocRoots, 4)
	bodyHTML := fmt.Sprintf(`  <nav epub:type="toc" id="toc">
    <h1>目录</h1>
%s
  </nav>
  <nav epub:type="landmarks" id="landmarks" hidden="">
    <h2>地标</h2>
    <ol>
      <li><a epub:type="cover" href="cover.xhtml">封面</a></li>
      <li><a epub:type="toc" href="nav.xhtml">目录</a></li>
      <li><a epub:type="bodymatter" href="%s">正文</a></li>
    </ol>
  </nav>`, navOL, bodymatterFilename)

	navHTML := makeXHTMLPage("目录 - "+b.Meta.Title, bodyHTML, "../styles/style.css", b.Meta.Language)

	return os.WriteFile(filepath.Join(b.TextDir, "nav.xhtml"), []byte(navHTML), 0644)
}
