// Package httpx は HTTP の共通処理。
package httpx

import (
	"encoding/json"
	"net/http"
)

// WriteProblem は RFC 9457 の Problem Details を書き込む。
func WriteProblem(w http.ResponseWriter, status int, title string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"title": title, "status": status})
}
