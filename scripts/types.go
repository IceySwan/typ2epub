package main

import (
	"strings"
	"unicode"

	nethtml "golang.org/x/net/html"
)

// ============================================================
// Data Models
// ============================================================

type BookMeta struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Author     string `json:"author"`
	Version    string `json:"version"`
	Date       string `json:"date"`
	Language   string `json:"lang"`
	Identifier string `json:"identifier"`
}

type Chapter struct {
	Title    string
	Filename string
	Matter   string // "front", "main", "appendix" / "back"
	Nodes    []*nethtml.Node
}

type TocNode struct {
	Title    string
	Href     string
	Level    int
	Children []*TocNode
}

type ManifestItem struct {
	ID         string
	Href       string
	MediaType  string
	Properties string // e.g. "cover-image", "nav", "mathml", "svg"
}

type SpineItem struct {
	IDRef  string
	Linear string // e.g. "no", ""
}

type FootnoteItem struct {
	DefID     string
	NoterefID string
	Node      *nethtml.Node
}

type FootnoteStore struct {
	Items []*FootnoteItem
	ByID  map[string]*FootnoteItem
}

func NewFootnoteStore() *FootnoteStore {
	return &FootnoteStore{
		Items: make([]*FootnoteItem, 0),
		ByID:  make(map[string]*FootnoteItem),
	}
}

func (s *FootnoteStore) Add(item *FootnoteItem) {
	s.Items = append(s.Items, item)
	if item.DefID != "" {
		s.ByID[item.DefID] = item
	}
	if item.NoterefID != "" {
		s.ByID[item.NoterefID] = item
	}
}

// ============================================================
// String Utility Helpers
// ============================================================

// sanitizeFilename replaces characters that are unsafe in file names
// (path separators, wildcards, quotes, whitespace, etc.) with '_'.
func sanitizeFilename(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\n', '\r', '\t', ' ':
			sb.WriteRune('_')
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// sanitizeChapterName keeps only letters, digits and '_', capped at 16 runes.
func sanitizeChapterName(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			sb.WriteRune(r)
		}
	}
	res := sb.String()
	runes := []rune(res)
	if len(runes) > 16 {
		return string(runes[:16])
	}
	return res
}
