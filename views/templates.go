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

// RenderGroup renders name with the cached template set for patterns and writes it
// straight to the response.
//
// Handlers used to assign c.Echo().Renderer per request, which mutates shared state: two
// concurrent requests from different route groups could render each other's templates.
// Rendering through this helper keeps the parsed sets cached and the renderer untouched.
func RenderGroup(c echo.Context, status int, name string, data any, patterns ...string) error {
	t := For(patterns...)

	if !c.Response().Committed {
		// HTML is per-session: a stale copy (an old login page without the CSRF hook, for
		// instance) fails in ways the user cannot diagnose.
		c.Response().Header().Set("Cache-Control", "no-store, must-revalidate")
		c.Response().Header().Set("Pragma", "no-cache")
		c.Response().WriteHeader(status)
	}

	return t.Render(c.Response(), name, data, c)
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
