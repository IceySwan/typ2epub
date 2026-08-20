package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	nethtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ============================================================
// Low-Level DOM Utilities (golang.org/x/net/html helpers)
// ============================================================

func getAttr(n *nethtml.Node, key string) string {
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func setAttr(n *nethtml.Node, key, val string) {
	if n == nil {
		return
	}
	for i, a := range n.Attr {
		if a.Key == key {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, nethtml.Attribute{Key: key, Val: val})
}

func hasAttr(n *nethtml.Node, key string) bool {
	return getAttr(n, key) != ""
}

func getText(n *nethtml.Node) string {
	if n == nil {
		return ""
	}
	var sb strings.Builder
	var traverse func(*nethtml.Node)
	traverse = func(curr *nethtml.Node) {
		if curr.Type == nethtml.TextNode {
			sb.WriteString(curr.Data)
		}
		for c := curr.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return sb.String()
}

func isWhitespaceText(n *nethtml.Node) bool {
	return n.Type == nethtml.TextNode && strings.TrimSpace(n.Data) == ""
}

func detachNode(n *nethtml.Node) {
	if n == nil || n.Parent == nil {
		return
	}
	parent := n.Parent
	if parent.FirstChild == n {
		parent.FirstChild = n.NextSibling
	}
	if parent.LastChild == n {
		parent.LastChild = n.PrevSibling
	}
	if n.PrevSibling != nil {
		n.PrevSibling.NextSibling = n.NextSibling
	}
	if n.NextSibling != nil {
		n.NextSibling.PrevSibling = n.PrevSibling
	}
	n.Parent = nil
	n.PrevSibling = nil
	n.NextSibling = nil
}

func appendChild(parent, child *nethtml.Node) {
	detachNode(child)
	child.Parent = parent
	if parent.LastChild == nil {
		parent.FirstChild = child
		parent.LastChild = child
	} else {
		parent.LastChild.NextSibling = child
		child.PrevSibling = parent.LastChild
		parent.LastChild = child
	}
}

func findAll(n *nethtml.Node, predicate func(*nethtml.Node) bool) []*nethtml.Node {
	var results []*nethtml.Node
	var walk func(*nethtml.Node)
	walk = func(curr *nethtml.Node) {
		if curr == nil {
			return
		}
		if predicate(curr) {
			results = append(results, curr)
		}
		for c := curr.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return results
}

func findFirst(n *nethtml.Node, predicate func(*nethtml.Node) bool) *nethtml.Node {
	var walk func(*nethtml.Node) *nethtml.Node
	walk = func(curr *nethtml.Node) *nethtml.Node {
		if curr == nil {
			return nil
		}
		if predicate(curr) {
			return curr
		}
		for c := curr.FirstChild; c != nil; c = c.NextSibling {
			if found := walk(c); found != nil {
				return found
			}
		}
		return nil
	}
	return walk(n)
}

func renderNodeToString(n *nethtml.Node) string {
	var buf bytes.Buffer
	if err := nethtml.Render(&buf, n); err != nil {
		return ""
	}
	return buf.String()
}

// ============================================================
// HTML Content Processing
//
// Standalone functions that transform parsed HTML content.
// None of these depend on EpubBuilder state — they operate
// purely on DOM nodes and data structures.
// ============================================================

// extractFootnotes detaches the global doc-endnotes section and returns
// an ordered FootnoteStore containing footnotes in their original appearance order.
func extractFootnotes(doc *nethtml.Node) *FootnoteStore {
	store := NewFootnoteStore()
	endnotesSection := findFirst(doc, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "section" &&
			(getAttr(n, "role") == "doc-endnotes" || getAttr(n, "epub:type") == "footnotes")
	})

	if endnotesSection == nil {
		return store
	}

	liNodes := findAll(endnotesSection, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && n.Data == "li" && hasAttr(n, "id")
	})

	for _, li := range liNodes {
		fnDefID := getAttr(li, "id")
		backlink := findFirst(li, func(n *nethtml.Node) bool {
			return n.Type == nethtml.ElementNode && n.Data == "sup" && getAttr(n, "role") == "doc-backlink"
		})
		var noterefID string
		if backlink != nil {
			aTag := findFirst(backlink, func(n *nethtml.Node) bool {
				return n.Type == nethtml.ElementNode && n.Data == "a" && hasAttr(n, "href")
			})
			if aTag != nil {
				noterefID = strings.TrimPrefix(getAttr(aTag, "href"), "#")
			}
		}
		detachNode(li)
		store.Add(&FootnoteItem{DefID: fnDefID, NoterefID: noterefID, Node: li})
	}

	detachNode(endnotesSection)
	return store
}

// removeTypstTOC strips the redundant static outline that Typst exports
// based strictly on semantic contracts (data-t2e-role="toc", role="doc-toc", or data-t2e-kind="toc").
func removeTypstTOC(doc *nethtml.Node) {
	tocNodes := findAll(doc, func(n *nethtml.Node) bool {
		if n.Type != nethtml.ElementNode {
			return false
		}
		return getAttr(n, "data-t2e-role") == "toc" ||
			getAttr(n, "role") == "doc-toc" ||
			getAttr(n, "data-t2e-kind") == "toc"
	})
	for _, n := range tocNodes {
		detachNode(n)
	}
}

type contentItem struct {
	node   *nethtml.Node
	matter string
}

func collectContentItems(n *nethtml.Node, currentMatter string) []contentItem {
	var items []contentItem
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if isWhitespaceText(c) {
			continue
		}
		if c.Type == nethtml.ElementNode && (c.Data == "section" || c.Data == "div") && hasAttr(c, "data-t2e-matter") {
			matter := getAttr(c, "data-t2e-matter")
			items = append(items, collectContentItems(c, matter)...)
			continue
		}
		items = append(items, contentItem{node: c, matter: currentMatter})
	}
	return items
}

func newChapter(title, filename, matter string, nodes []*nethtml.Node) *Chapter {
	return &Chapter{
		Title:    title,
		Filename: filename,
		Matter:   matter,
		Nodes:    nodes,
	}
}

// splitChapters divides the HTML body into chapters by data-t2e-matter and <h2> boundaries.
func splitChapters(body *nethtml.Node) []*Chapter {
	var chapters []*Chapter
	items := collectContentItems(body, "")

	current := newChapter("前言", "chap_00_intro.xhtml", "front", make([]*nethtml.Node, 0))

	chapIdx := 1
	for _, item := range items {
		child := item.node
		matter := item.matter
		if matter == "" {
			matter = "main"
		}

		isSplit := child.Type == nethtml.ElementNode &&
			(child.Data == "h2" || getAttr(child, "data-t2e-split") == "chapter")

		if isSplit {
			chapTitle := strings.TrimSpace(getText(child))
			chapID := fmt.Sprintf("chap_%02d", chapIdx)
			safeName := sanitizeChapterName(chapTitle)
			if safeName == "" {
				safeName = fmt.Sprintf("section_%d", chapIdx)
			}
			chapFilename := fmt.Sprintf("%s_%s.xhtml", chapID, safeName)

			if len(current.Nodes) > 0 {
				chapters = append(chapters, current)
			}

			current = newChapter(chapTitle, chapFilename, matter, []*nethtml.Node{child})
			chapIdx++
		} else {
			if len(current.Nodes) == 0 && item.matter != "" {
				current.Matter = item.matter
			}
			current.Nodes = append(current.Nodes, child)
		}
	}

	if len(current.Nodes) > 0 {
		chapters = append(chapters, current)
	}

	return chapters
}

// assignFootnotes distributes extracted footnotes back into their
// originating chapters as per-chapter endnote sections in exact document reference order.
func assignFootnotes(chapters []*Chapter, store *FootnoteStore) {
	if store == nil || len(store.Items) == 0 {
		return
	}

	for _, chap := range chapters {
		var chapFootnoteNodes []*nethtml.Node
		seenDefIDs := make(map[string]bool)

		for _, node := range chap.Nodes {
			noterefs := findAll(node, func(n *nethtml.Node) bool {
				return n.Type == nethtml.ElementNode && n.Data == "sup" && getAttr(n, "role") == "doc-noteref"
			})
			for _, sup := range noterefs {
				noterefID := getAttr(sup, "id")
				var item *FootnoteItem
				if noterefID != "" {
					item = store.ByID[noterefID]
				}
				if item == nil {
					aTag := findFirst(sup, func(n *nethtml.Node) bool {
						return n.Type == nethtml.ElementNode && n.Data == "a" && hasAttr(n, "href")
					})
					if aTag != nil {
						targetDefID := strings.TrimPrefix(getAttr(aTag, "href"), "#")
						item = store.ByID[targetDefID]
					}
				}

				if item != nil && !seenDefIDs[item.DefID] {
					seenDefIDs[item.DefID] = true
					chapFootnoteNodes = append(chapFootnoteNodes, item.Node)
				}
			}
		}

		if len(chapFootnoteNodes) > 0 {
			fnSection := &nethtml.Node{
				Type:     nethtml.ElementNode,
				Data:     "section",
				DataAtom: atom.Section,
				Attr: []nethtml.Attribute{
					{Key: "class", Val: "chapter-footnotes"},
					{Key: "role", Val: "doc-endnotes"},
					{Key: "epub:type", Val: "footnotes"},
				},
			}
			fnOL := &nethtml.Node{
				Type:     nethtml.ElementNode,
				Data:     "ol",
				DataAtom: atom.Ol,
				Attr: []nethtml.Attribute{
					{Key: "style", Val: "list-style-type: none"},
				},
			}
			appendChild(fnSection, fnOL)
			for _, li := range chapFootnoteNodes {
				appendChild(fnOL, li)
			}
			chap.Nodes = append(chap.Nodes, fnSection)
		}
	}
}

func getElementsWithID(node *nethtml.Node) []*nethtml.Node {
	return findAll(node, func(n *nethtml.Node) bool {
		return n.Type == nethtml.ElementNode && hasAttr(n, "id")
	})
}

// rewriteCrossReferences scans all chapters for in-document #id links
// and rewrites them to point to the correct chapter file.
func rewriteCrossReferences(chapters []*Chapter) {
	idToFile := make(map[string]string)
	for _, chap := range chapters {
		for _, node := range chap.Nodes {
			for _, el := range getElementsWithID(node) {
				idToFile[getAttr(el, "id")] = chap.Filename
			}
		}
	}

	for _, chap := range chapters {
		for _, node := range chap.Nodes {
			aTags := findAll(node, func(n *nethtml.Node) bool {
				return n.Type == nethtml.ElementNode && n.Data == "a" && hasAttr(n, "href")
			})
			for _, a := range aTags {
				href := getAttr(a, "href")
				if strings.HasPrefix(href, "#") {
					targetID := href[1:]
					if targetFile, ok := idToFile[targetID]; ok && targetFile != chap.Filename {
						setAttr(a, "href", targetFile+"#"+targetID)
					}
				}
			}
		}
	}
}

// buildChapterTocTree scans a chapter's nodes for sub-headings (h3–h5) and
// returns a hierarchical TocNode tree rooted at the chapter level.
func buildChapterTocTree(chap *Chapter, chapIdx int) *TocNode {
	rootNode := &TocNode{Title: chap.Title, Href: chap.Filename, Level: 2}
	headingStack := []*TocNode{rootNode}

	var headingsInChap []*nethtml.Node
	for _, node := range chap.Nodes {
		headingsInChap = append(headingsInChap, findAll(node, func(n *nethtml.Node) bool {
			return n.Type == nethtml.ElementNode && (n.Data == "h3" || n.Data == "h4" || n.Data == "h5")
		})...)
	}

	for idx, h := range headingsInChap {
		lvl, _ := strconv.Atoi(h.Data[1:])
		hTitle := strings.TrimSpace(getText(h))
		hID := getAttr(h, "id")
		if hID == "" {
			hID = fmt.Sprintf("h_%02d_%02d", chapIdx, idx+1)
			setAttr(h, "id", hID)
		}
		hHref := fmt.Sprintf("%s#%s", chap.Filename, hID)
		subNode := &TocNode{Title: hTitle, Href: hHref, Level: lvl}

		for len(headingStack) > 1 && headingStack[len(headingStack)-1].Level >= lvl {
			headingStack = headingStack[:len(headingStack)-1]
		}

		top := headingStack[len(headingStack)-1]
		top.Children = append(top.Children, subNode)
		headingStack = append(headingStack, subNode)
	}

	return rootNode
}
