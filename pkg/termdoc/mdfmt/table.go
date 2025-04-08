package mdfmt

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// WriteTable writes a markdown table with equal spaced columns.
// Text within the columns will be left-aligned.
func WriteTable(header []string, rows [][]string) string {
	return WriteTableWithAlignment(header, rows, nil)
}

// TextAlignment defines text alignment.
type TextAlignment uint8

// Defined text alignments.
const (
	TextAlignmentDefault TextAlignment = iota
	TextAlignmentLeft
	TextAlignmentCenter
	TextAlignmentRight
)

// WriteTableWithAlignment writes a markdown table with equal spaced columns.
// Text within the columns will be aligned according to the alignments given.
func WriteTableWithAlignment(header []string, rows [][]string, alignment []TextAlignment) string {
	// Default the alignment
	if alignment == nil {
		alignment = slices.Repeat([]TextAlignment{TextAlignmentLeft}, len(header))
	}

	// Get maximum width of each column
	colMaxLens := make([]int, len(header))
	for _, row := range rows {
		for col, cell := range row {
			cellLen := ansi.StringWidth(cell) // ansi-aware string width
			if cellLen > colMaxLens[col] {
				colMaxLens[col] = cellLen
			}
		}
	}

	fmtStrings := make([]string, len(header))
	for col, maxLen := range colMaxLens {
		switch alignment[col] {
		case TextAlignmentRight:
			// %5s
			fmtStrings[col] = "%" + strconv.Itoa(maxLen) + "s"
		// case TextAlignmentCenter:
		// 	// __%s___
		// 	left := maxLen / 2
		// 	right := maxLen - left
		// 	fmtStrings[col] = strings.Repeat(" ", left) + "%s" + strings.Repeat(" ", right)
		// case
		// 	TextAlignmentLeft,
		// 	TextAlignmentDefault:
		// 	fallthrough
		default:
			// %-5s
			fmtStrings[col] = "%-" + strconv.Itoa(maxLen) + "s"
		}
	}

	w := &strings.Builder{}

	writeRow := func(row []string) {
		for col, cell := range row {
			_, _ = w.WriteString("| " + fmt.Sprintf(fmtStrings[col], cell) + " ")
		}
		_, _ = w.WriteString("|\n")
	}

	// Write header row
	writeRow(header)

	// Write separator row
	for col := range header {
		switch alignment[col] {
		case TextAlignmentRight:
			_, _ = fmt.Fprintf(w, "| %s: ", strings.Repeat("-", colMaxLens[col]-1))
		case TextAlignmentCenter:
			_, _ = fmt.Fprintf(w, "| :%s: ", strings.Repeat("-", colMaxLens[col]-2))
		case
			TextAlignmentLeft,
			TextAlignmentDefault:
			fallthrough
		default:
			_, _ = fmt.Fprintf(w, "| %s ", strings.Repeat("-", colMaxLens[col]))
		}
	}
	_, _ = w.WriteString("|\n")

	// Write separator row
	for _, row := range rows {
		writeRow(row)
	}

	return w.String()
}
