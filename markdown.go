package gobook

import (
	"log"

	. "github.com/gobook/blackfriday"
)

func MarkdownToHtml(content string, renderParameters HtmlRendererParameters) (str string) {
	defer func() {
		e := recover()
		if e != nil {
			str = content
			log.Println("Render Markdown ERR:", e)
		}
	}()

	htmlFlags := 0

	htmlFlags |= HTML_USE_XHTML
	htmlFlags |= HTML_USE_SMARTYPANTS
	htmlFlags |= HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= HTML_SMARTYPANTS_LATEX_DASHES

	renderer := HtmlRendererWithParameters(htmlFlags, "", "", renderParameters)

	// set up the parser
	extensions := 0
	extensions |= EXTENSION_NO_INTRA_EMPHASIS
	extensions |= EXTENSION_TABLES
	extensions |= EXTENSION_FENCED_CODE
	extensions |= EXTENSION_AUTOLINK
	extensions |= EXTENSION_STRIKETHROUGH
	extensions |= EXTENSION_SPACE_HEADERS

	str = string(Markdown([]byte(content), renderer, extensions))

	return str
}
