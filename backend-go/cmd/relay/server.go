package main

import (
	"relay/internal/api"
	handoffhttp "relay/internal/handoff/adapter/http"
	identityhttp "relay/internal/identity/adapter/http"
)

// 埋め込むフィールド名を分けるための別名。
type (
	identityHandler = identityhttp.Handler
	handoffHandler  = handoffhttp.Handler
)

// server は各モジュールのハンドラーを埋め込み、api.StrictServerInterface を満たす。
// API を追加したときに実装漏れがあれば、下の型アサーションでコンパイルエラーになる。
type server struct {
	*identityHandler
	*handoffHandler
}

var _ api.StrictServerInterface = server{}
