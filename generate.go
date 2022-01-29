package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/antchfx/htmlquery"
	"github.com/gomarkdown/markdown"
	"golang.org/x/net/html"
)

type Row struct {
	Doc       string
	DocSpan   int
	DocHTML   template.HTML
	Code      string
	CodeHTML  template.HTML
	CodeEmpty bool
	FirstCode bool
}

type Example struct {
	Id              string
	Name            string
	Rows            []*Row
	FullCode        template.JS
	PreviousExample *Example
	NextExample     *Example
}

const destDir = "public"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func idToName(id string) string {
	return strings.Title(strings.ToLower(strings.Replace(id, "-", " ", -1)))
}

func parseExample(exampleName string) *Example {
	src := "examples/" + exampleName + "/" + exampleName + ".yml"
	file, err := os.Open(src)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var rows []*Row
	var row *Row = &Row{}
	inSpan := false
	var spanningRow *Row = nil
	fullCode := ""
	foundFirstCode := false
	for scanner.Scan() {
		line := scanner.Text()
		commentPrefix := regexp.MustCompile(`^\s*#\s*`)
		if commentPrefix.MatchString(line) {
			// Process comment line (ie documentation)
			line = commentPrefix.ReplaceAllString(line, "")

			switch strings.TrimSpace(line) {
			case "::span-comment":
				inSpan = true
				spanningRow = row
			case "::end-span":
				inSpan = false
				row = &Row{}
			case "::newline":
				row.Doc += "  \n\n"
			default:
				// update the row's documentation
				if len(row.Doc) > 0 {
					line = " " + line
				}
				row.Doc += line
			}
		} else {
			// Process code line

			// skip completely empty lines
			if strings.TrimSpace(line) == "" && row.Doc == "" {
				continue
			}
			if strings.TrimSpace(line) != "" {
				fullCode += line + "\n"
			}

			row.Code = line
			if !foundFirstCode && line != "" {
				foundFirstCode = true
				row.FirstCode = true
			}

			if inSpan {
				spanningRow.DocSpan += 1
				rows = append(rows, row)
				// since we're still spanning a doc, the next row should have no doc (ie no span)
				row = &Row{}
				row.DocSpan = -1
			} else {
				row.DocSpan = 1
				rows = append(rows, row)
				row = &Row{}
			}
		}
	}

	err = scanner.Err()
	check(err)

	b := new(bytes.Buffer)
	style := styles.Get("autumn")
	formatter := chromahtml.New(chromahtml.WithClasses(true))
	lexer := lexers.Get("yaml")
	iterator, err := lexer.Tokenise(nil, fullCode)
	check(err)
	err = formatter.Format(b, style, iterator)
	check(err)
	codeDoc, err := htmlquery.Parse(strings.NewReader(b.String()))
	codeRows := htmlquery.Find(codeDoc, "//span[@class=\"line\"]")

	codeRowIdx := -1
	for _, row := range rows {
		// change docs markdown to html
		row.DocHTML = template.HTML(markdown.ToHTML([]byte(row.Doc), nil, nil))

		// if code is empty
		row.CodeEmpty = strings.TrimSpace(row.Code) == ""

		if row.FirstCode {
			codeRowIdx = 0
		}

		if codeRowIdx >= 0 {
			b := new(bytes.Buffer)
			err := html.Render(b, codeRows[codeRowIdx])
			check(err)
			row.CodeHTML = template.HTML(b.String())
			codeRowIdx += 1
		}
	}

	// make the code safe for inserting in template string
	fullCode = strings.Replace(fullCode, "`", "\\`", -1)
	fullCode = strings.Replace(fullCode, "$", "\\$", -1)

	return &Example{
		Id:              exampleName,
		Name:            idToName(exampleName),
		Rows:            rows,
		FullCode:        template.JS(fullCode),
		PreviousExample: nil,
		NextExample:     nil,
	}
}

func templatizeExample(example *Example) {
	fmt.Println("Generating " + example.Id)
	exampleTmpl := template.New("example")
	bytes, err := ioutil.ReadFile("templates/example.tmpl")
	check(err)
	_, err = exampleTmpl.Parse(string(bytes))
	check(err)

	exampleF, err := os.Create("public/" + example.Id + ".html")
	check(err)
	exampleTmpl.Execute(exampleF, example)
}

func templatizeIndex(examples []*Example) {
	fmt.Println("Generating index")
	indexTmpl := template.New("index")
	bytes, err := ioutil.ReadFile("templates/index.tmpl")
	check(err)
	_, err = indexTmpl.Parse(string(bytes))
	check(err)

	indexF, err := os.Create("public/index.html")
	check(err)
	indexTmpl.Execute(indexF, examples)
}

func generateStyles() {
	style := styles.Get("autumn")
	formatter := chromahtml.New(chromahtml.WithClasses(true))
	f, err := os.Create("public/default.css")
	check(err)
	defer f.Close()
	err = formatter.WriteCSS(f, style)
	check(err)
}

func main() {
	fmt.Println("Starting...")
	file, err := os.Open("examples.txt")
	check(err)
	defer file.Close()

	allExamples := []*Example{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		exampleName := scanner.Text()
		example := parseExample(exampleName)
		allExamples = append(allExamples, example)
	}

	for idx, example := range allExamples {
		if idx > 0 {
			example.PreviousExample = allExamples[idx-1]
		}
		if idx < len(allExamples)-1 {
			example.NextExample = allExamples[idx+1]
		}

		templatizeExample(example)
	}

	templatizeIndex(allExamples)

	fmt.Println("Generating CSS styles...")
	generateStyles()
	fmt.Println("Finished Successfully")
}
