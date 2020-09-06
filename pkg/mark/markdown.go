package mark

import (
	"bytes"
	"regexp"
	"strings"
	"strconv"

	"github.com/kovetskiy/mark/pkg/log"
	"github.com/kovetskiy/mark/pkg/mark/stdlib"
	"github.com/russross/blackfriday"
)

type ConfluenceRenderer struct {
	blackfriday.Renderer

	Stdlib *stdlib.Lib
}

func ParseLang(lang string) string {
	paramlist := strings.Fields(lang)
	if len(paramlist) == 0 {
		return lang
	}
	if paramlist[0] == "title" {
		return ""
	}
	return paramlist[0]
}

func ParseTitle(lang string) string {
	index := strings.Index(lang, "title")
	if index >= 0 {
		return lang[index+6:]
	}
	return ""
}

func (renderer ConfluenceRenderer) BlockCode(
	out *bytes.Buffer,
	text []byte,
	lang string,
) {
	renderer.Stdlib.Templates.ExecuteTemplate(
		out,
		"ac:code",
		struct {
			Language string
			Collapse string
			Title    string
			Text     string
		}{
			ParseLang(lang),
			strconv.FormatBool(strings.Contains(lang, "collapse")),
			ParseTitle(lang),
			string(text),
		},
	)
}

// compileMarkdown will replace tags like <ac:rich-tech-body> with escaped
// equivalent, because blackfriday markdown parser replaces that tags with
// <a href="ac:rich-text-body">ac:rich-text-body</a> for whatever reason.
func CompileMarkdown(
	markdown []byte,
	stdlib *stdlib.Lib,
) string {
	log.Tracef(nil, "rendering markdown:\n%s", string(markdown))

	colon := regexp.MustCompile(`---BLACKFRIDAY-COLON---`)

	tags := regexp.MustCompile(`<(/?\S+?):(\S+?)>`)

	markdown = tags.ReplaceAll(
		markdown,
		[]byte(`<$1`+colon.String()+`$2>`),
	)

	renderer := ConfluenceRenderer{
		Renderer: blackfriday.HtmlRenderer(
			blackfriday.HTML_USE_XHTML|
				blackfriday.HTML_USE_SMARTYPANTS|
				blackfriday.HTML_SMARTYPANTS_FRACTIONS|
				blackfriday.HTML_SMARTYPANTS_DASHES|
				blackfriday.HTML_SMARTYPANTS_LATEX_DASHES,
			"", "",
		),

		Stdlib: stdlib,
	}

	html := blackfriday.MarkdownOptions(
		markdown,
		renderer,
		blackfriday.Options{
			Extensions: blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
				blackfriday.EXTENSION_TABLES |
				blackfriday.EXTENSION_FENCED_CODE |
				blackfriday.EXTENSION_AUTOLINK |
				blackfriday.EXTENSION_LAX_HTML_BLOCKS |
				blackfriday.EXTENSION_STRIKETHROUGH |
				blackfriday.EXTENSION_SPACE_HEADERS |
				blackfriday.EXTENSION_HEADER_IDS |
				blackfriday.EXTENSION_AUTO_HEADER_IDS |
				blackfriday.EXTENSION_TITLEBLOCK |
				blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
				blackfriday.EXTENSION_DEFINITION_LISTS |
				blackfriday.EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK,
		},
	)

	html = colon.ReplaceAll(html, []byte(`:`))

	log.Tracef(nil, "rendered markdown to html:\n%s", string(html))

	return string(html)
}
