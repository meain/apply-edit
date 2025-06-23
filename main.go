package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <filename>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: echo 'diff content' | %s file.txt\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

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
