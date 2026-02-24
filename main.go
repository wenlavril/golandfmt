package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	maxLen := flag.Int("m", 120, "maximum line length")
	tabWidth := flag.Int("t", 4, "tab width in spaces")
	write := flag.Bool("w", false, "write result to (source) file instead of stdout")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		src, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "golandfmt: reading stdin: %v\n", err)
			os.Exit(1)
		}
		out, err := FormatWithTabWidth(src, *maxLen, *tabWidth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "golandfmt: %v\n", err)
			os.Exit(1)
		}
		os.Stdout.Write(out)
		return
	}

	for _, path := range args {
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "golandfmt: %v\n", err)
			os.Exit(1)
		}
		out, err := FormatWithTabWidth(src, *maxLen, *tabWidth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "golandfmt: %s: %v\n", path, err)
			os.Exit(1)
		}
		if *write {
			if err := os.WriteFile(path, out, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "golandfmt: %v\n", err)
				os.Exit(1)
			}
		} else {
			os.Stdout.Write(out)
		}
	}
}
