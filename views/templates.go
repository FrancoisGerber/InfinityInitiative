package templates

import (
	"html/template"
	"io"
	"log"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func ParseTemplates(patterns []string) *Templates {
	fileList := []string{}

	for _, pattern := range patterns {
		files := getFiles(pattern)

		if len(files) > 0 {
			fileList = append(fileList, files...)
		}
	}

	return &Templates{
		templates: template.Must(template.ParseFiles(fileList...)),
	}
}

func getFiles(pattern string) []string {
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Println(err.Error())
	}

	return files
}
