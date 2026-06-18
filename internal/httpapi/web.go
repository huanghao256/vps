package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed webdist
var embeddedWeb embed.FS

func (s *Server) serveWeb(w http.ResponseWriter, r *http.Request) {
	dist := filepath.Join("web", "dist")
	index := filepath.Join(dist, "index.html")
	if _, err := os.Stat(index); err == nil {
		http.FileServer(http.Dir(dist)).ServeHTTP(w, r)
		return
	}

	if webFS, err := fs.Sub(embeddedWeb, "webdist"); err == nil {
		if _, err := webFS.Open("index.html"); err == nil {
			http.FileServer(http.FS(webFS)).ServeHTTP(w, r)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(fallbackHTML))
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
