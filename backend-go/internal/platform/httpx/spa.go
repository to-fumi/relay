package httpx

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// SPAHandler は dir の静的ファイルを配信し、存在しないパスには index.html を返す。
func SPAHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+r.URL.Path)))
		if info, err := os.Stat(p); err != nil || info.IsDir() {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		files.ServeHTTP(w, r)
	})
}
