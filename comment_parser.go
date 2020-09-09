package main

import (
	"strings"

	"github.com/eidolon/wordwrap"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func prettyPrintComments(c Comments, commentTree *string, indentlevel int) string {
	x, _ := terminal.Width()
	rightPadding := 3
	comment := parseComment(c.Comment)
	wrapper := wordwrap.Wrapper(int(x)-indentlevel-rightPadding, false)

	fullComment := ""
	commentLines := strings.Split(comment, "\n")
	for _, line := range commentLines {
		wrapped := wrapper(line)
		wrappedAndIndentedComment := wordwrap.Indent(wrapped, getIndentBlock(indentlevel), true)
		fullComment += wrappedAndIndentedComment + "\n" + "\n"
	}

	// wrapped := wrapper(comment)
	// wrappedAndIndentedComment := wordwrap.Indent(wrapped, getIndentBlock(indentlevel), true)
	wrappedAndIndentedAuthor := wordwrap.Indent(c.Author, getIndentBlock(indentlevel), true)

	wrappedAndIndentedComment := "\033[1m" + wrappedAndIndentedAuthor + "\033[0m" + "\n" + fullComment

	*commentTree = *commentTree + wrappedAndIndentedComment
	for _, s := range c.Replies {
		prettyPrintComments(*s, commentTree, indentlevel+5)
	}
	return *commentTree
}

func getIndentBlock(level int) string {
	indentation := " "
	for i := 1; i < level; i++ {
		indentation = indentation + " "
	}
	return indentation
}

func parseComment(comment string) string {
	fixedHTML := replaceHTML(comment)
	fixedHTMLAndCharacters := replaceCharacters(fixedHTML)
	return fixedHTMLAndCharacters
}

func replaceCharacters(input string) string {
	input = strings.ReplaceAll(input, "&#x27;", "'")
	input = strings.ReplaceAll(input, "&gt;", ">")
	input = strings.ReplaceAll(input, "&lt;", "<")
	input = strings.ReplaceAll(input, "&#x2F;", "/")
	input = strings.ReplaceAll(input, "&quot;", "\"")
	input = strings.ReplaceAll(input, "&amp;", "&")
	return input
}

func replaceHTML(input string) string {
	input = strings.Replace(input, "<p>", "", 1)
	input = strings.ReplaceAll(input, "<p>", "\n")
	input = strings.ReplaceAll(input, "<i>", "\033[3m")
	input = strings.ReplaceAll(input, "</i>", "\033[0m")
	return input
}