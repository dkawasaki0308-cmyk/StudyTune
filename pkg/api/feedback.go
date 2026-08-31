// Vercel サーバーレス関数: POST /api/feedback
//
// 「難しすぎた / 簡単すぎた / ちょうど良かった」の明示フィードバックを受け取る。
// これは唯一の教師信号であり、将来 Elo 方式の動的レベル更新の入力になる:
//
//	新レベル = 現レベル + K × (実際の結果 - 期待値)
//	期待値   = σ(ユーザーレベル - 記事難易度)
//
// MVP スコープでは永続化しない (DB を持たない)。受理して 202 を返し、
// 構造化ログにだけ残す。保存先が決まったらこの関数の中だけ差し替える。
package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"studytune/pkg/httpx"
)

type feedbackRequest struct {
	URL          string  `json:"url"`          // 対象記事の URL
	Kind         string  `json:"kind"`         // too-hard | too-easy | just-right
	ArticleLevel string  `json:"articleLevel"` // beginner | intermediate | advanced
	Difficulty   float64 `json:"difficulty"`   // 記事の推定難易度 0..5 (任意)
}

var validKind = map[string]bool{"too-hard": true, "too-easy": true, "just-right": true}

// Feedback は POST /api/feedback を処理する。
func Feedback(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		httpx.Error(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<10))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "read body failed")
		return
	}
	var req feedbackRequest
	if err := json.Unmarshal(body, &req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" || !strings.HasPrefix(req.URL, "http") {
		httpx.Error(w, http.StatusBadRequest, "url is required")
		return
	}
	if !validKind[req.Kind] {
		httpx.Error(w, http.StatusBadRequest, "kind must be too-hard|too-easy|just-right")
		return
	}

	// TODO(persistence): DB 導入時にここを INSERT に差し替える。
	// 教師信号として user_id ごとに蓄積し、Elo 更新のバッチで消費する。
	log.Printf("feedback kind=%s level=%s difficulty=%.2f url=%s",
		req.Kind, req.ArticleLevel, req.Difficulty, req.URL)

	httpx.JSON(w, http.StatusAccepted, map[string]any{
		"accepted": true,
		"note":     "MVP scope: 受理のみ。永続化とElo更新は後続。",
	})
}
