// This is adapted from:
//   - https://github.com/a-h/templ/blob/main/generator/htmldiff/diff.go
//   - https://github.com/a-h/htmlformat/blob/main/format.go
//
// The exported function `DiffStrings` determines whether one HTML string is an acceptable
// substitute for the other.
//
// Differences from the original:
//   - <script> tags are compared after removing comments, whitespace, and blank lines
//   - completely empty elements like "<div></div>" won't have a blank line inserted into them (the
//     original would render this as "<div>\n</div>", breaking opening and closing tags onto
//     separate lines.  I don't know why, seems like a bug to me)
//
// It appears that `html/template` removes comments from script tags and in some cases does other
// minification as well, while Templ does some pretty-printing.  So here we make them as similar as
// possible to ensure an equality check remains high-fidelity.
package diffs

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func DiffStrings(expected, actual string) (diff string, err error) {
	// Format both strings.
	var wg sync.WaitGroup
	wg.Add(2)

	errs := make([]error, 2)

	// Format expected.
	go func() {
		defer wg.Done()
		e := new(strings.Builder)
		err := fragment(e, strings.NewReader(expected))
		if err != nil {
			errs[0] = fmt.Errorf("expected html formatting error: %w", err)
		}
		expected = e.String()
	}()

	// Format actual.
	go func() {
		defer wg.Done()
		a := new(strings.Builder)
		err := fragment(a, strings.NewReader(actual))
		if err != nil {
			errs[1] = fmt.Errorf("actual html formatting error: %w", err)
		}
		actual = a.String()
	}()

	// Wait for processing.
	wg.Wait()

	return cmp.Diff(expected, actual), errors.Join(errs...)
}

// func Fmt(s string) string {
// 	b := new(strings.Builder)
// 	err := fragment(b, strings.NewReader(s))
// 	if err != nil {
// 		panic(err)
// 	}
// 	return strings.TrimSpace(b.String())
// }

// `fragment` formats a fragment of a HTML document.
func fragment(w io.Writer, r io.Reader) (err error) {
	context := &html.Node{
		Type: html.ElementNode,
	}
	nodes, err := html.ParseFragment(r, context)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if err = printNode(w, node, false, 0); err != nil {
			return
		}
	}
	return
}

// Is this node a tag with no end tag such as <meta> or <br>?
// http://www.w3.org/TR/html-markup/syntax.html#syntax-elements
func isVoidElement(n *html.Node) bool {
	switch n.DataAtom {
	case atom.Area, atom.Base, atom.Br, atom.Col, atom.Command, atom.Embed,
		atom.Hr, atom.Img, atom.Input, atom.Keygen, atom.Link,
		atom.Meta, atom.Param, atom.Source, atom.Track, atom.Wbr:
		return true
	}
	return false
}

func getFirstRune(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

func hasSingleTextChild(n *html.Node) bool {
	return n != nil && n.FirstChild != nil && n.FirstChild == n.LastChild && n.FirstChild.Type == html.TextNode
}

func remove_js_comments(code string) string {
	// Remove block comments
	block := regexp.MustCompile(`(?s)/\*.*?\*/`) // `*?` is a non-greedy version of `*`; `(?s)` makes `.` match newlines
	code = block.ReplaceAllString(code, "")

	// Remove line comments
	code = regexp.MustCompile(`(?m)//.*$`).ReplaceAllString(code, "") // `(?m)` enables multiline mode

	// Remove indentation and trailing spaces
	code = regexp.MustCompile(`(?m)^\s*`).ReplaceAllString(code, "")
	code = regexp.MustCompile(`(?m)\s*$`).ReplaceAllString(code, "")

	// Remove blank lines
	code = regexp.MustCompile(`\n+`).ReplaceAllString(code, "\n")
	return strings.TrimSpace(code)
}

func printNode(w io.Writer, n *html.Node, pre bool, level int) (err error) {
	switch n.Type {
	case html.TextNode:
		if n.Parent != nil && n.Parent.Data == "script" {
			lines := strings.Split(remove_js_comments(n.Data), "\n")
			for _, l := range lines {
				if err = printIndent(w, level); err != nil {
					return
				}
				if _, err = fmt.Fprint(w, strings.TrimSpace(l)); err != nil {
					return
				}
			}
			return
		}
		if pre {
			if _, err = fmt.Fprint(w, n.Data); err != nil {
				return
			}
			return nil
		}
		s := strings.TrimSpace(n.Data)
		if s != "" {
			if !hasSingleTextChild(n.Parent) &&
				(n.PrevSibling == nil || !unicode.IsPunct(getFirstRune(s))) {
				if err = printIndent(w, level); err != nil {
					return
				}
			}
			if _, err = fmt.Fprint(w, s); err != nil {
				return
			}
			if !hasSingleTextChild(n.Parent) {
				if _, err = fmt.Fprint(w, "\n"); err != nil {
					return
				}
			}
		}
	case html.ElementNode:
		if err = printIndent(w, level); err != nil {
			return
		}
		if _, err = fmt.Fprintf(w, "<%s", n.Data); err != nil {
			return
		}
		for _, a := range n.Attr {
			val := html.EscapeString(a.Val)
			if _, err = fmt.Fprintf(w, ` %s="%s"`, a.Key, val); err != nil {
				return
			}
		}
		if _, err = fmt.Fprint(w, ">"); err != nil {
			return
		}
		if !hasSingleTextChild(n) && n.FirstChild != nil {
			if _, err = fmt.Fprint(w, "\n"); err != nil {
				return
			}
		}
		if !isVoidElement(n) {
			if err = printChildren(w, n, isPreFormatted(n.Data), level+1); err != nil {
				return
			}
			if !hasSingleTextChild(n) {
				if err = printIndent(w, level); err != nil {
					return
				}
			}
			if _, err = fmt.Fprintf(w, "</%s>", n.Data); err != nil {
				return
			}

			if n.NextSibling == nil ||
				(!unicode.IsPunct(getFirstRune(n.NextSibling.Data)) || n.NextSibling.Type == html.ElementNode) {
				if _, err = fmt.Fprint(w, "\n"); err != nil {
					return
				}
			}
		}
	case html.CommentNode:
		if err = printIndent(w, level); err != nil {
			return
		}
		if _, err = fmt.Fprintf(w, "<!--%s-->\n", n.Data); err != nil {
			return
		}
		if err = printChildren(w, n, false, level); err != nil {
			return
		}
	case html.DoctypeNode, html.DocumentNode:
		if err = printChildren(w, n, false, level); err != nil {
			return
		}
	}
	return
}

func isPreFormatted(s string) bool {
	return s == "pre" || s == "script" || s == "style"
}

func printChildren(w io.Writer, n *html.Node, pre bool, level int) (err error) {
	child := n.FirstChild
	for child != nil {
		if err = printNode(w, child, pre, level); err != nil {
			return
		}
		child = child.NextSibling
	}
	return
}

func printIndent(w io.Writer, level int) (err error) {
	_, err = fmt.Fprint(w, strings.Repeat(" ", level))
	return err
}
