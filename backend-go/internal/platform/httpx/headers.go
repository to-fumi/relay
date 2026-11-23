package httpx

import "net/http"

// SecurityHeaders は、どの応答にも付けるセキュリティ用のヘッダーを足す。
//   - X-Frame-Options：ほかのサイトの iframe に入れさせない（クリックジャッキング）
//   - X-Content-Type-Options：Content-Type と違う種類として読ませない
//   - Referrer-Policy：ほかのサイトへ移るときに、URL のパスやクエリを送らない
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
