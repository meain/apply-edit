package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	var explain bool
	flag.BoolVar(&explain, "explain", false, "Show example usage")
	flag.Parse()

	if explain {
		showExample()
		return
	}

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--explain] <filename>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Use --explain to see example usage\n")
		os.Exit(1)
	}

	filename := flag.Arg(0)

	// Read diff from stdin
	diff, err := readDiffFromStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading diff from stdin: %v\n", err)
		os.Exit(1)
	}

	// Parse the diff
	searchBlock, replaceBlock, err := parseDiff(diff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing diff: %v\n", err)
		os.Exit(1)
	}

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	// Perform the edit
	newContent, err := performEdit(string(content), searchBlock, replaceBlock)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error performing edit: %v\n", err)
		os.Exit(1)
	}

	// Write the modified content back to the file
	err = os.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", filename, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully applied edit to %s\n", filename)
}

func showExample() {
	fmt.Println("apply-edit - Apply search and replace edits to files")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Printf("  %s [--explain] <filename>\n", os.Args[0])
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Reads a diff from stdin and applies it to the specified file.")
	fmt.Println("  The diff uses a special format with SEARCH and REPLACE blocks.")
	fmt.Println()
	fmt.Println("EXAMPLE:")
	fmt.Println("  Given a file 'app.py' with contents:")
	fmt.Println("    from flask import Flask")
	fmt.Println("    app = Flask(__name__)")
	fmt.Println()
	fmt.Println("  Run this command:")
	fmt.Printf("    cat <<EOF | %s app.py\n", os.Args[0])
	fmt.Println("    <<<<<<< SEARCH")
	fmt.Println("    from flask import Flask")
	fmt.Println("    =======")
	fmt.Println("    import math")
	fmt.Println("    from flask import Flask")
	fmt.Println("    >>>>>>> REPLACE")
	fmt.Println("    EOF")
	fmt.Println()
	fmt.Println("  Result: The file will be updated to:")
	fmt.Println("    import math")
	fmt.Println("    from flask import Flask")
	fmt.Println("    app = Flask(__name__)")
	fmt.Println()
	fmt.Println("FORMAT:")
	fmt.Println("  <<<<<<< SEARCH")
	fmt.Println("  [text to find]")
	fmt.Println("  =======")
	fmt.Println("  [text to replace with]")
	fmt.Println("  >>>>>>> REPLACE")
	fmt.Println()
	fmt.Println("NOTES:")
	fmt.Println("  - The search text must match exactly (including whitespace)")
	fmt.Println("  - If multiple matches exist, the operation will fail to avoid ambiguity")
	fmt.Println("  - Empty replace blocks will delete the search text")
	fmt.Println("  - The original file is overwritten with the changes")
}

func readDiffFromStdin() (string, error) {
	var builder strings.Builder
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line != "" {
					builder.WriteString(line)
				}
				break
			}
			return "", err
		}
		builder.WriteString(line)
	}

	return builder.String(), nil
}

func parseDiff(diff string) (searchBlock, replaceBlock string, err error) {
	lines := strings.Split(strings.TrimSpace(diff), "\n")
	
	var searchLines, replaceLines []string
	var inSearch, inReplace bool
	
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "<<<<<<< SEARCH"):
			inSearch = true
			inReplace = false
		case strings.HasPrefix(line, "======="):
			inSearch = false
			inReplace = true
		case strings.HasPrefix(line, ">>>>>>> REPLACE"):
			inSearch = false
			inReplace = false
		case inSearch:
			searchLines = append(searchLines, line)
		case inReplace:
			replaceLines = append(replaceLines, line)
		}
	}
	
	if len(searchLines) == 0 {
		return "", "", fmt.Errorf("no search block found in diff")
	}
	
	searchBlock = strings.Join(searchLines, "\n")
	replaceBlock = strings.Join(replaceLines, "\n")
	
	return searchBlock, replaceBlock, nil
}

func performEdit(content, searchBlock, replaceBlock string) (string, error) {
	// Handle the case where search block might have different line endings
	normalizedContent := strings.ReplaceAll(content, "\r\n", "\n")
	normalizedSearch := strings.ReplaceAll(searchBlock, "\r\n", "\n")
	
	// Find the search block in the content
	index := strings.Index(normalizedContent, normalizedSearch)
	if index == -1 {
		return "", fmt.Errorf("search block not found in file:\n%s", searchBlock)
	}
	
	// Check if there are multiple occurrences
	if strings.Index(normalizedContent[index+len(normalizedSearch):], normalizedSearch) != -1 {
		return "", fmt.Errorf("multiple occurrences of search block found - edit would be ambiguous")
	}
	
	// Perform the replacement
	newContent := normalizedContent[:index] + replaceBlock + normalizedContent[index+len(normalizedSearch):]
	
	return newContent, nil
}
