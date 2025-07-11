package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type fileEntry struct {
	name     string
	typ      string
	size     int64
	modified time.Time
}

var (
	header    = []string{"#", "name", "type", "size", "modified"}
	colWidths = make([]int, len(header))
	rows      [][]string

	colorBlue   = "\033[1;92m" // Folder
	colorYellow = "\033[1;97m" // File
	colorReset  = "\033[0m"

	ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	files := readDir(dir)

	sort.SliceStable(files, func(i, j int) bool {
		if files[i].typ == "folder" && files[j].typ != "folder" {
			return true
		}
		if files[i].typ != "folder" && files[j].typ == "folder" {
			return false
		}
		return files[i].name < files[j].name
	})

	for i, f := range files {
		rawName := f.name
		if f.typ == "folder" {
			rawName = colorBlue + rawName + colorReset
		} else {
			rawName = colorYellow + rawName + colorReset
		}

		row := []string{
			fmt.Sprintf("%d", i),
			rawName,
			f.typ,
			humanSize(f.size),
			humanTime(f.modified),
		}
		rows = append(rows, row)
	}

	for i, h := range header {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row {
			vlen := visualLength(col)
			if vlen > colWidths[i] {
				colWidths[i] = vlen
			}
		}
	}

	printTableHeader()
	printTableRows()
	printTableFooter()
}

func readDir(path string) []fileEntry {
	var list []fileEntry
	entries, err := os.ReadDir(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return list
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		ftype := "file"
		if entry.IsDir() {
			ftype = "folder"
		}

		list = append(list, fileEntry{
			name:     entry.Name(),
			typ:      ftype,
			size:     info.Size(),
			modified: info.ModTime(),
		})
	}
	return list
}

func printTableHeader() {
	borderTop := "╭"
	borderMid := "├"

	for i, w := range colWidths {
		borderTop += strings.Repeat("─", w+2)
		borderMid += strings.Repeat("─", w+2)
		if i < len(colWidths)-1 {
			borderTop += "┬"
			borderMid += "┼"
		}
	}
	borderTop += "╮"
	borderMid += "┤"

	fmt.Println(borderTop)
	fmt.Print("│")
	for i, h := range header {
		fmt.Printf(" %-*s │", colWidths[i], h)
	}
	fmt.Println()
	fmt.Println(borderMid)
}

func printTableRows() {
	for _, row := range rows {
		fmt.Print("│")
		for i, col := range row {
			padding := colWidths[i] - visualLength(col)
			fmt.Printf(" %s%s │", col, strings.Repeat(" ", padding))
		}
		fmt.Println()
	}
}

func printTableFooter() {
	borderBot := "╰"
	for i, w := range colWidths {
		borderBot += strings.Repeat("─", w+2)
		if i < len(colWidths)-1 {
			borderBot += "┴"
		}
	}
	borderBot += "╯"
	fmt.Println(borderBot)
}

func humanSize(bytes int64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1f kB", float64(bytes)/1024)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
	}
}

func humanTime(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	case diff < 30*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	default:
		return t.Format("Jan 02, 2006")
	}
}

func visualLength(s string) int {
	return len(ansiRegexp.ReplaceAllString(s, ""))
}
