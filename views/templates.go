package templates

import (
	"html/template"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type cacheEntry struct {
	once      sync.Once
	templates *Templates
}

var (
	cacheMu sync.RWMutex
	cache   = map[string]*cacheEntry{}
)

// For returns the parsed template set for a group of file patterns, parsing it once per
// process instead of once per request.
//
// Handlers still assign the result to the shared echo.Renderer, which is a data race
// (two concurrent requests can swap each other's template set). Removing that properly
// means rendering through a per-request renderer instead of c.Render; until then this at
// least removes the per-request file I/O (60 files) from every page load.
func For(patterns ...string) *Templates {
	key := strings.Join(patterns, "|")

	cacheMu.RLock()
	entry, ok := cache[key]
	cacheMu.RUnlock()

	if !ok {
		cacheMu.Lock()
		entry, ok = cache[key]
		if !ok {
			entry = &cacheEntry{}
			cache[key] = entry
		}
		cacheMu.Unlock()
	}

	entry.once.Do(func() {
		entry.templates = ParseTemplates(patterns)
	})

	return entry.templates
}

// ParseTemplates parses a fresh set. Prefer For() so the files are read once.
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
