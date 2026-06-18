package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed webdist
var embeddedWeb embed.FS

func (s *server) serveWeb(w http.ResponseWriter, r *http.Request) {
	dist := filepath.Join("web", "dist")
	index := filepath.Join(dist, "index.html")
	if _, err := os.Stat(index); err == nil {
		serveLocalSPA(w, r, dist, index)
		return
	}

	if webFS, err := fs.Sub(embeddedWeb, "webdist"); err == nil {
		if _, err := webFS.Open("index.html"); err == nil {
			serveEmbeddedSPA(w, r, webFS)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(fallbackHTML))
}

func serveLocalSPA(w http.ResponseWriter, r *http.Request, dist, index string) {
	target := strings.TrimPrefix(filepath.Clean(r.URL.Path), string(filepath.Separator))
	if target != "." && target != "" {
		path := filepath.Join(dist, target)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
	}
	http.ServeFile(w, r, index)
}

func serveEmbeddedSPA(w http.ResponseWriter, r *http.Request, webFS fs.FS) {
	target := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if target != "" {
		if file, err := webFS.Open(target); err == nil {
			defer file.Close()
			if info, err := file.Stat(); err == nil && !info.IsDir() {
				http.FileServer(http.FS(webFS)).ServeHTTP(w, r)
				return
			}
		}
	}

	index, err := fs.ReadFile(webFS, "index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(index)
}

const fallbackHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>VPS Inspector</title>
    <style>
      body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f7f8fa; color: #15191f; }
      main { max-width: 760px; margin: 12vh auto; padding: 0 24px; }
      h1 { font-size: 40px; margin: 0 0 12px; }
      p { color: #59616d; line-height: 1.6; }
      button { border: 0; background: #1967d2; color: white; padding: 10px 14px; border-radius: 6px; cursor: pointer; }
      pre { background: #101418; color: #d8e2ee; padding: 16px; border-radius: 8px; overflow: auto; }
    </style>
  </head>
  <body>
    <main>
      <h1>VPS Inspector</h1>
      <p>The backend is running. Build the React frontend from <code>web/</code> for the full interface.</p>
      <button id="run">Run inspection</button>
      <pre id="out">Ready.</pre>
    </main>
    <script>
      document.getElementById('run').onclick = async () => {
        const out = document.getElementById('out');
        out.textContent = 'Running...';
        const res = await fetch('/api/v1/runs', { method: 'POST', body: '{}' });
        out.textContent = JSON.stringify(await res.json(), null, 2);
      };
    </script>
  </body>
</html>`
