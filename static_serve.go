package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/HexmosTech/git-lrc/result"
)

//go:embed static/*
var staticFiles embed.FS

type JSONTemplateData = result.JSONTemplateData
type JSONFileData = result.JSONFileData
type JSONHunkData = result.JSONHunkData
type JSONLineData = result.JSONLineData
type JSONCommentData = result.JSONCommentData

// renderPreactHTML renders the Preact-based HTML with embedded JSON data
func renderPreactHTML(data *HTMLTemplateData) (string, error) {
	// Convert to JSON-serializable format
	jsonData := result.ConvertToJSONData(data)

	// Serialize to JSON
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return "", err
	}

	// Read the HTML template
	htmlBytes, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		return "", err
	}

	// Replace the placeholder with actual JSON data
	html := string(htmlBytes)
	html = strings.Replace(html, "{{.JSONData}}", string(jsonBytes), 1)

	// Update title if friendly name is present
	if data.FriendlyName != "" {
		html = strings.Replace(html, "<title>LiveReview Results</title>",
			"<title>LiveReview Results — "+data.FriendlyName+"</title>", 1)
	}

	return html, nil
}

// getStaticHandler returns an HTTP handler for serving static files
func getStaticHandler() http.Handler {
	// Get the static subdirectory
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(staticFS))
}

// serveStaticFile serves a specific static file
func serveStaticFile(w http.ResponseWriter, r *http.Request, filename string) error {
	content, err := staticFiles.ReadFile("static/" + filename)
	if err != nil {
		return err
	}

	// Set content type based on extension
	switch {
	case strings.HasSuffix(filename, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(filename, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case strings.HasSuffix(filename, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	if _, err := io.Copy(w, bytes.NewReader(content)); err != nil {
		return err
	}
	return nil
}
