package analyzer

import (
	"strings"

	"baliance.com/gooxml/document"
	"github.com/xuri/excelize/v2"
	"rsc.io/pdf"
)

func ReadDocx(path string) (string, error) {
	doc, err := document.Open(path)
	if err != nil {
		return "", err
	}

	var textBuilder strings.Builder
	for _, p := range doc.Paragraphs() {
		for _, r := range p.Runs() {
			textBuilder.WriteString(r.Text())
		}
		textBuilder.WriteString("\n")
	}
	return textBuilder.String(), nil
}

func ReadXlsx(path string) (string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var textBuilder strings.Builder
	for _, sheetName := range f.GetSheetList() {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			continue
		}
		for _, row := range rows {
			textBuilder.WriteString(strings.Join(row, "\t") + "\n")
		}
	}
	return textBuilder.String(), nil
}

func ReadPdf(path string) (string, error) {
	f, err := pdf.Open(path)
	if err != nil {
		return "", err
	}

	var textBuilder strings.Builder
	for i := 1; i <= f.NumPage(); i++ {
		page := f.Page(i)
		if page.V.IsNull() {
			continue
		}
		content := page.Content()
		var lastX, lastY float64
		for i, text := range content.Text {
			if i == 0 || text.X != lastX || text.Y != lastY {
				textBuilder.WriteString(" ")
			}
			textBuilder.WriteString(text.S)
			lastX, lastY = text.X, text.Y
		}
	}
	return textBuilder.String(), nil
}
