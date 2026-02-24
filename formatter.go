package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

const defaultTabWidth = 4

// Format applies gofmt followed by GoLand-style "wrap if long" formatting.
func Format(src []byte, maxLen int) ([]byte, error) {
	return FormatWithTabWidth(src, maxLen, defaultTabWidth)
}

func FormatWithTabWidth(src []byte, maxLen, tabWidth int) ([]byte, error) {
	formatted, err := format.Source(src)
	if err != nil {
		return nil, err
	}

	result, changed, err := wrapIfLong(formatted, maxLen, tabWidth)
	if err != nil || !changed {
		return formatted, err
	}

	final, err := format.Source(result)
	if err != nil {
		return result, nil
	}
	return final, nil
}

type replacement struct {
	start int
	end   int
	text  string
}

type wrappable struct {
	openPos  token.Pos
	closePos token.Pos
	items    []token.Pos
	itemEnds []token.Pos
}

// visualLen computes the visual column width of a string, treating each tab
// as advancing to the next tab stop (multiples of tabWidth).
func visualLen(s string, tabWidth int) int {
	col := 0
	for _, ch := range s {
		if ch == '\t' {
			col = ((col / tabWidth) + 1) * tabWidth
		} else {
			col++
		}
	}
	return col
}

func wrapIfLong(src []byte, maxLen, tabWidth int) ([]byte, bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, false, err
	}

	var nodes []wrappable
	collectWrappables(file, &nodes)

	if len(nodes) == 0 {
		return src, false, nil
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].openPos > nodes[j].openPos
	})

	result := string(src)
	changed := false
	for _, n := range nodes {
		newResult := processNode(result, fset, n, maxLen, tabWidth)
		if newResult != result {
			changed = true
			result = newResult
		}
	}

	if !changed {
		return src, false, nil
	}
	return []byte(result), true, nil
}

func collectWrappables(file *ast.File, nodes *[]wrappable) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			if len(node.Args) > 0 {
				w := wrappable{openPos: node.Lparen, closePos: node.Rparen}
				for _, arg := range node.Args {
					w.items = append(w.items, arg.Pos())
					w.itemEnds = append(w.itemEnds, arg.End())
				}
				*nodes = append(*nodes, w)
			}
		case *ast.CompositeLit:
			if len(node.Elts) > 0 {
				w := wrappable{openPos: node.Lbrace, closePos: node.Rbrace}
				for _, elt := range node.Elts {
					w.items = append(w.items, elt.Pos())
					w.itemEnds = append(w.itemEnds, elt.End())
				}
				*nodes = append(*nodes, w)
			}
		case *ast.FuncDecl:
			if node.Type != nil {
				collectFuncType(node.Type, nodes)
			}
		case *ast.FuncLit:
			if node.Type != nil {
				collectFuncType(node.Type, nodes)
			}
		}
		return true
	})
}

func collectFuncType(ft *ast.FuncType, nodes *[]wrappable) {
	if ft.Params != nil && len(ft.Params.List) > 0 {
		w := wrappable{openPos: ft.Params.Opening, closePos: ft.Params.Closing}
		for _, field := range ft.Params.List {
			w.items = append(w.items, field.Pos())
			w.itemEnds = append(w.itemEnds, field.End())
		}
		*nodes = append(*nodes, w)
	}
	if ft.Results != nil && len(ft.Results.List) > 1 && ft.Results.Opening.IsValid() {
		w := wrappable{openPos: ft.Results.Opening, closePos: ft.Results.Closing}
		for _, field := range ft.Results.List {
			w.items = append(w.items, field.Pos())
			w.itemEnds = append(w.itemEnds, field.End())
		}
		*nodes = append(*nodes, w)
	}
}

func processNode(src string, fset *token.FileSet, n wrappable, maxLen, tabWidth int) string {
	openOff := fset.Position(n.openPos).Offset
	closeOff := fset.Position(n.closePos).Offset

	if openOff < 0 || closeOff < 0 || openOff >= len(src) || closeOff >= len(src) {
		return src
	}

	openChar := src[openOff]
	closeChar := src[closeOff]

	itemTexts := extractItems(src, fset, n)
	if len(itemTexts) == 0 {
		return src
	}

	lineStart := findLineStart(src, openOff)
	prefix := src[lineStart:openOff]
	linePrefix := extractIndent(prefix)

	singleLine := string(openChar) + strings.Join(itemTexts, ", ") + string(closeChar)
	fullLine := prefix + singleLine

	hasMultiLineItem := false
	for _, item := range itemTexts {
		if strings.Contains(item, "\n") {
			hasMultiLineItem = true
			break
		}
	}

	if !hasMultiLineItem && visualLen(fullLine, tabWidth) <= maxLen {
		return src[:openOff] + singleLine + src[closeOff+1:]
	}

	itemIndent := linePrefix + "\t"
	packed := packItems(itemTexts, itemIndent, maxLen, tabWidth)

	wrapped := string(openChar) + "\n" + packed + linePrefix + string(closeChar)

	return src[:openOff] + wrapped + src[closeOff+1:]
}

func extractItems(src string, fset *token.FileSet, n wrappable) []string {
	var items []string
	for i := range n.items {
		startOff := fset.Position(n.items[i]).Offset
		endOff := fset.Position(n.itemEnds[i]).Offset
		if startOff < 0 || endOff < 0 || startOff > len(src) || endOff > len(src) {
			continue
		}
		text := strings.TrimSpace(src[startOff:endOff])
		text = normalizeWhitespace(text)
		items = append(items, text)
	}
	return items
}

func normalizeWhitespace(s string) string {
	if !strings.Contains(s, "\n") {
		return s
	}
	var b strings.Builder
	inSpace := false
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			if !inSpace {
				b.WriteRune(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return b.String()
}

func packItems(items []string, indent string, maxLen, tabWidth int) string {
	var lines []string
	indentVisLen := visualLen(indent, tabWidth)

	var currentLineItems []string
	currentVisLen := indentVisLen

	for _, item := range items {
		itemVisLen := visualLen(item, tabWidth)

		if len(currentLineItems) == 0 {
			currentLineItems = append(currentLineItems, item)
			currentVisLen = indentVisLen + itemVisLen
		} else {
			// +2 for ", " separator, +1 for trailing comma
			newVisLen := currentVisLen + 2 + itemVisLen
			if newVisLen+1 <= maxLen {
				currentLineItems = append(currentLineItems, item)
				currentVisLen = newVisLen
			} else {
				lines = append(lines, indent+strings.Join(currentLineItems, ", ")+",")
				currentLineItems = []string{item}
				currentVisLen = indentVisLen + itemVisLen
			}
		}
	}

	if len(currentLineItems) > 0 {
		lines = append(lines, indent+strings.Join(currentLineItems, ", ")+",")
	}

	return strings.Join(lines, "\n") + "\n"
}

func findLineStart(src string, offset int) int {
	for i := offset - 1; i >= 0; i-- {
		if src[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

func extractIndent(linePrefix string) string {
	var indent strings.Builder
	for _, ch := range linePrefix {
		if ch == ' ' || ch == '\t' {
			indent.WriteRune(ch)
		} else {
			break
		}
	}
	return indent.String()
}
