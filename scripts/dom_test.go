package main

import (
	"bytes"
	"testing"

	nethtml "golang.org/x/net/html"
)

func parseHTMLFragment(t *testing.T, htmlStr string) *nethtml.Node {
	t.Helper()
	doc, err := nethtml.Parse(bytes.NewReader([]byte(htmlStr)))
	if err != nil {
		t.Fatalf("failed to parse test HTML fragment: %v", err)
	}
	return doc
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello_World"},
		{"a/b\\c:d*e?f\"g<h>i|j", "a_b_c_d_e_f_g_h_i_j"},
		{"My-Document_v1.0", "My-Document_v1.0"},
	}

	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSanitizeChapterName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.1 绪论与背景", "11绪论与背景"},
		{"Chapter_01: Math & Physics", "Chapter_01MathPh"},
		{"", ""},
	}

	for _, tt := range tests {
		got := sanitizeChapterName(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeChapterName(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestExtractFootnotes(t *testing.T) {
	htmlStr := `<html><body>
		<div><p>Some text<sup role="doc-noteref" id="fnref1"><a href="#fn1">1</a></sup></p></div>
		<section role="doc-endnotes">
			<ol>
				<li id="fn1">Footnote 1 text <sup role="doc-backlink"><a href="#fnref1">↩</a></sup></li>
			</ol>
		</section>
	</body></html>`

	doc := parseHTMLFragment(t, htmlStr)
	store := extractFootnotes(doc)

	if len(store.Items) != 1 {
		t.Fatalf("expected 1 footnote, got %d", len(store.Items))
	}
	item := store.ByID["fn1"]
	if item == nil {
		t.Fatalf("expected footnote 'fn1' in store")
	}
	if item.NoterefID != "fnref1" {
		t.Errorf("expected NoterefID 'fnref1', got %q", item.NoterefID)
	}

	// Verify endnotes section was detached
	endnotes := findFirst(doc, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "section" && getAttr(n, "role") == "doc-endnotes"
	})
	if endnotes != nil {
		t.Errorf("expected doc-endnotes section to be detached from DOM")
	}
}

func TestRemoveTypstTOC(t *testing.T) {
	// Verify semantic data-t2e-role="toc" is removed, but genuine chapters named "Contents" without semantic tag are preserved
	htmlStr := `<html><body>
		<nav data-t2e-role="toc">
			<h2>Contents</h2>
			<ol><li><a href="#chap1">Chapter 1</a></li></ol>
		</nav>
		<h2>第一章 正文</h2>
		<p>正文内容</p>
		<h2>Contents & Web Analysis</h2>
		<p>关于 Web 内容的讨论章节</p>
	</body></html>`

	doc := parseHTMLFragment(t, htmlStr)
	removeTypstTOC(doc)

	h2Nodes := findAll(doc, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "h2"
	})
	if len(h2Nodes) != 2 {
		t.Fatalf("expected 2 h2 remaining, got %d", len(h2Nodes))
	}
	if getText(h2Nodes[0]) != "第一章 正文" {
		t.Errorf("expected '第一章 正文', got %q", getText(h2Nodes[0]))
	}
	if getText(h2Nodes[1]) != "Contents & Web Analysis" {
		t.Errorf("expected 'Contents & Web Analysis', got %q", getText(h2Nodes[1]))
	}

	navNodes := findAll(doc, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "nav"
	})
	if len(navNodes) != 0 {
		t.Errorf("expected 0 nav elements, got %d", len(navNodes))
	}
}

func TestSplitChapters(t *testing.T) {
	htmlStr := `<body>
		<section data-t2e-matter="front">
			<h2>题记</h2>
			<p>前言题记内容</p>
		</section>
		<section data-t2e-matter="main">
			<h2>第一章 绪论</h2>
			<p>第一章内容</p>
			<h2>第二章 展开</h2>
			<p>第二章内容</p>
		</section>
		<section data-t2e-matter="appendix">
			<h2>附录 A</h2>
			<p>附录内容</p>
		</section>
	</body>`

	doc := parseHTMLFragment(t, htmlStr)
	body := findFirst(doc, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "body"
	})

	chapters := splitChapters(body)
	if len(chapters) != 4 {
		t.Fatalf("expected 4 chapters (front + 2 main + 1 appendix), got %d", len(chapters))
	}

	if chapters[0].Title != "题记" || chapters[0].Matter != "front" {
		t.Errorf("unexpected chapter 0: %+v", chapters[0])
	}
	if chapters[1].Title != "第一章 绪论" || chapters[1].Matter != "main" {
		t.Errorf("unexpected chapter 1: %+v", chapters[1])
	}
	if chapters[2].Title != "第二章 展开" || chapters[2].Matter != "main" {
		t.Errorf("unexpected chapter 2: %+v", chapters[2])
	}
	if chapters[3].Title != "附录 A" || chapters[3].Matter != "appendix" {
		t.Errorf("unexpected chapter 3: %+v", chapters[3])
	}
}

func TestAssignFootnotes(t *testing.T) {
	htmlStr := `<body>
		<h2>第一章</h2>
		<p>段落一<sup role="doc-noteref" id="fnref1"><a href="#fn1">1</a></sup>并带有第二个引用<sup role="doc-noteref" id="fnref2"><a href="#fn2">2</a></sup></p>
		<h2>第二章</h2>
		<p>段落二<sup role="doc-noteref" id="fnref3"><a href="#fn3">3</a></sup></p>
	</body>`

	doc := parseHTMLFragment(t, htmlStr)
	body := findFirst(doc, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "body"
	})

	chapters := splitChapters(body)
	fnNode1 := &nethtml.Node{Type: nethtml.ElementNode, Data: "li", Attr: []nethtml.Attribute{{Key: "id", Val: "fn1"}}}
	fnNode2 := &nethtml.Node{Type: nethtml.ElementNode, Data: "li", Attr: []nethtml.Attribute{{Key: "id", Val: "fn2"}}}
	fnNode3 := &nethtml.Node{Type: nethtml.ElementNode, Data: "li", Attr: []nethtml.Attribute{{Key: "id", Val: "fn3"}}}

	store := NewFootnoteStore()
	store.Add(&FootnoteItem{DefID: "fn1", NoterefID: "fnref1", Node: fnNode1})
	store.Add(&FootnoteItem{DefID: "fn2", NoterefID: "fnref2", Node: fnNode2})
	store.Add(&FootnoteItem{DefID: "fn3", NoterefID: "fnref3", Node: fnNode3})

	assignFootnotes(chapters, store)

	fnSec0 := findFirst(chapters[0].Nodes[len(chapters[0].Nodes)-1], func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "section" && getAttr(n, "class") == "chapter-footnotes"
	})
	if fnSec0 == nil {
		t.Fatalf("expected chapter 0 to have footnotes section")
	}

	// Verify both footnotes in chapter 0 are preserved in order
	liList := findAll(fnSec0, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "li"
	})
	if len(liList) != 2 {
		t.Fatalf("expected 2 footnotes in chapter 0, got %d", len(liList))
	}
	if getAttr(liList[0], "id") != "fn1" || getAttr(liList[1], "id") != "fn2" {
		t.Errorf("footnotes not preserved in expected order: fn1, fn2")
	}
}

func TestRewriteCrossReferences(t *testing.T) {
	chap1 := &Chapter{
		Title:    "第一章",
		Filename: "chap_01.xhtml",
		Nodes: []*nethtml.Node{
			{
				Type: nethtml.ElementNode,
				Data: "div",
				Attr: []nethtml.Attribute{{Key: "id", Val: "target-in-chap1"}},
			},
		},
	}

	aLink := &nethtml.Node{
		Type: nethtml.ElementNode,
		Data: "a",
		Attr: []nethtml.Attribute{{Key: "href", Val: "#target-in-chap1"}},
	}
	chap2 := &Chapter{
		Title:    "第二章",
		Filename: "chap_02.xhtml",
		Nodes:    []*nethtml.Node{aLink},
	}

	rewriteCrossReferences([]*Chapter{chap1, chap2})

	gotHref := getAttr(aLink, "href")
	expectedHref := "chap_01.xhtml#target-in-chap1"
	if gotHref != expectedHref {
		t.Errorf("rewriteCrossReferences() href = %q; want %q", gotHref, expectedHref)
	}
}

func TestBuildChapterTocTree(t *testing.T) {
	h2Node := &nethtml.Node{Type: nethtml.ElementNode, Data: "h2", FirstChild: &nethtml.Node{Type: nethtml.TextNode, Data: "第一章 核心"}}
	h3Node1 := &nethtml.Node{Type: nethtml.ElementNode, Data: "h3", FirstChild: &nethtml.Node{Type: nethtml.TextNode, Data: "1.1 背景"}}
	h4Node := &nethtml.Node{Type: nethtml.ElementNode, Data: "h4", FirstChild: &nethtml.Node{Type: nethtml.TextNode, Data: "1.1.1 历史"}}
	h3Node2 := &nethtml.Node{Type: nethtml.ElementNode, Data: "h3", FirstChild: &nethtml.Node{Type: nethtml.TextNode, Data: "1.2 目标"}}

	chap := &Chapter{
		Title:    "第一章 核心",
		Filename: "chap_01.xhtml",
		Nodes:    []*nethtml.Node{h2Node, h3Node1, h4Node, h3Node2},
	}

	tree := buildChapterTocTree(chap, 1)

	if tree.Title != "第一章 核心" || tree.Href != "chap_01.xhtml" {
		t.Errorf("unexpected root TOC node: %+v", tree)
	}
	if len(tree.Children) != 2 {
		t.Fatalf("expected 2 level-3 children, got %d", len(tree.Children))
	}
	if tree.Children[0].Title != "1.1 背景" {
		t.Errorf("unexpected first sub-node: %+v", tree.Children[0])
	}
	if len(tree.Children[0].Children) != 1 || tree.Children[0].Children[0].Title != "1.1.1 历史" {
		t.Errorf("unexpected nested level-4 child: %+v", tree.Children[0].Children)
	}
	if tree.Children[1].Title != "1.2 目标" {
		t.Errorf("unexpected second sub-node: %+v", tree.Children[1])
	}
}
