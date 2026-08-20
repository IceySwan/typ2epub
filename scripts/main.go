package main

import (
	"flag"
	"fmt"
	"os"
)

// ============================================================
// CLI Entrypoint
// ============================================================

func main() {
	const defaultInput = "main.typ"
	const defaultOutput = ""
	const defaultCover = ""

	inputFlag := flag.String("input", defaultInput, "Path to main.typ (default: main.typ)")
	flag.StringVar(inputFlag, "i", defaultInput, "Path to main.typ (shorthand)")

	outputFlag := flag.String("output", defaultOutput, "Output EPUB path (default: build/<title>-v<version>.epub)")
	flag.StringVar(outputFlag, "o", defaultOutput, "Output EPUB path (shorthand)")

	coverFlag := flag.String("cover", defaultCover, "Optional cover image path (PNG/JPG/SVG)")
	flag.StringVar(coverFlag, "c", defaultCover, "Optional cover image path (shorthand)")

	keepTempFlag := flag.Bool("keep-temp", false, "Keep temporary build directory")

	flag.Parse()

	builder, err := NewEpubBuilder(*inputFlag, *outputFlag, *coverFlag, *keepTempFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[!] Initialization Error: %v\n", err)
		os.Exit(1)
	}
	if err := builder.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "[!] Error: %v\n", err)
		os.Exit(1)
	}
}
