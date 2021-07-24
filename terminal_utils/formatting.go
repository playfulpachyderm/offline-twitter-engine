package terminal_utils

import (
    "time"
    "strings"
)

/**
 * Format a timestamp in human-readable form.
 */
func FormatDate(t time.Time) string {
    return t.Format("Jan 2, 2006  15:04:05")
}


/**
 * Wrap lines to fixed width, while respecting word breaks
 */
func WrapParagraph(paragraph string, width int) []string {
    var lines []string
    i := 0
    for i < len(paragraph) - width {
        // Find a word break at the end of the line to avoid splitting up words
        end := i + width
        for end > i && paragraph[end] != ' ' {  // Look for a space, starting at the end
            end -= 1
        }
        lines = append(lines, paragraph[i:end])
        i = end + 1
    }
    lines = append(lines, paragraph[i:])
    return lines
}


/**
 * Return the text as a wrapped, indented block
 */
func WrapText(text string, width int) string {
    paragraphs := strings.Split(text, "\n")
    var lines []string
    for _, paragraph := range paragraphs {
        lines = append(lines, WrapParagraph(paragraph, width)...)
    }
    return strings.Join(lines, "\n    ")
}
