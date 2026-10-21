package templates

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

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
	err       error
}

var (
	cacheMu sync.RWMutex
	cache   = map[string]*cacheEntry{}

	// parseCount is how many template sets have actually been read from disk. It is
	// exported for the test that asserts the files are read once per process.
	parseCount atomic.Int64
)

// ParseCount reports how many template sets have been parsed so far.
func ParseCount() int64 {
	return parseCount.Load()
}

// Parse reads a fresh set and reports the error instead of panicking, so the caller can
// decide whether a bad template is fatal now (boot) or later (first request).
func Parse(patterns []string) (*Templates, error) {
	fileList := []string{}

	for _, pattern := range patterns {
		files := getFiles(pattern)

		if len(files) > 0 {
			fileList = append(fileList, files...)
		}
	}

	if len(fileList) == 0 {
		return nil, fmt.Errorf("no template files matched %v", patterns)
	}

	parsed, err := template.ParseFiles(fileList...)
	if err != nil {
		return nil, err
	}

	parseCount.Add(1)

	return &Templates{templates: parsed}, nil
}

// load returns the cached set for patterns, parsing it at most once per process.
func load(patterns []string) (*Templates, error) {
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
		entry.templates, entry.err = Parse(patterns)
	})

	return entry.templates, entry.err
}

// Preload parses every route group up front, so a template syntax error stops the
// process at boot instead of turning into a 500 on the first request that needs it.
func Preload() error {
	groups, err := Groups()
	if err != nil {
		return err
	}

	for _, patterns := range groups {
		if _, err := load(patterns); err != nil {
			return fmt.Errorf("templates %v: %w", patterns, err)
		}
	}

	return nil
}

// Groups lists the template set of every route group, derived from the views tree:
// the layout plus each section directory, and the auth shell plus each auth screen.
// It mirrors the per-controller viewPatterns without importing them. A directory that
// cannot be read is an error: silently dropping its group would let its syntax errors
// surface as a 500 on the first request, the failure mode Preload exists to remove.
func Groups() ([][]string, error) {
	groups := [][]string{{"views/layout/*.html"}}

	entries, err := os.ReadDir("views")
	if err != nil {
		return nil, fmt.Errorf("read views directory: %w", err)
	}

	for _, entry := range entries {
		// layout is already the fixed first group; assets holds no templates.
		if !entry.IsDir() || entry.Name() == "assets" || entry.Name() == "layout" {
			continue
		}

		dir := filepath.Join("views", entry.Name())

		if entry.Name() == "auth" {
			screens, err := os.ReadDir(dir)
			if err != nil {
				return nil, fmt.Errorf("read auth screens: %w", err)
			}

			for _, screen := range screens {
				if screen.IsDir() {
					groups = append(groups,
						[]string{"views/auth/*.html", filepath.Join(dir, screen.Name(), "*.html")})
				}
			}

			continue
		}

		groups = append(groups, []string{"views/layout/*.html", filepath.Join(dir, "*.html")})
	}

	return groups, nil
}

// For returns the parsed template set for a group of file patterns, parsing it once per
// process instead of once per request. Render with RenderGroup rather than assigning
// c.Echo().Renderer: the renderer is shared state.
func For(patterns ...string) *Templates {
	t, err := load(patterns)
	if err != nil {
		// Callers that reached this point cannot render anything; a panic is louder than a
		// blank page. Preload() turns the same error into a boot failure.
		panic(err)
	}

	return t
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

// ParseTemplates parses a fresh set and panics on error. Prefer For() so the files are
// read once, and Preload() at boot so a bad template fails there.
func ParseTemplates(patterns []string) *Templates {
	parsed, err := Parse(patterns)
	if err != nil {
		panic(err)
	}

	return parsed
}

func getFiles(pattern string) []string {
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Println(err.Error())
	}

	return files
}
