// Vercel サーバーレス関数のエントリポイント。
//
// Vercel Go ランタイムは /api 直下の各 .go ファイルを 1 関数として個別ビルドする。
// ここでは 1 関数にまとめ、パスで内部ディスパッチする (ローカルの `go build ./...` も
// 素直に通る)。実処理は internal/api に分離。
//
// ルーティング (vercel.json の rewrite で /api/(.*) をここへ集約):
//
//	GET  /api/resources  -> api.Resources
//	POST /api/feedback   -> api.Feedback
package handler

import (
	"net/http"
	"strings"

	"studytune/internal/api"
	"studytune/internal/httpx"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api"), "/")
	switch path {
	case "resources":
		api.Resources(w, r)
	case "feedback":
		api.Feedback(w, r)
	case "", "health":
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	default:
		httpx.Error(w, http.StatusNotFound, "unknown endpoint: "+path)
	}
}
