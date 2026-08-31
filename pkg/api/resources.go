// Package api は各 HTTP エンドポイントのハンドラ本体。
// Vercel のエントリポイント (api/index.go) から呼ばれる。
package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"studytune/pkg/catalog"
	"studytune/pkg/httpx"
	"studytune/pkg/qiita"
	"studytune/pkg/sample"
	"studytune/pkg/scoring"
)

// Resources は GET /api/resources を処理する。
//
// Qiita API v2 から記事を取得 → scoring で難易度を推定 → レベル/トピックで絞り込み → JSON。
// Qiita 取得に失敗したら埋め込みサンプルにフォールバックする (デモを止めない)。
//
// クエリ:
//
//	q      検索クエリ (既定: "tag:Go")
//	level  beginner | intermediate | advanced
//	topic  go | docker | sql ... (scoring の topicAliases のトピック名)
//	per    取得件数 1..50 (既定: 20)
func Resources(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		httpx.Error(w, http.StatusMethodNotAllowed, "GET only")
		return
	}

	q := r.URL.Query()
	query := q.Get("q")
	if query == "" {
		query = "tag:Go"
	}
	per, _ := strconv.Atoi(q.Get("per"))
	if per <= 0 || per > 50 {
		per = 20
	}
	level := q.Get("level")
	topic := q.Get("topic")

	// Vercel Hobby の関数上限 10s に収める。
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	scorer := scoring.NewScorer()

	items, err := qiita.New().Search(ctx, qiita.SearchOptions{Query: query, PerPage: per})
	source := "qiita"
	if err != nil || len(items) == 0 {
		items = sample.Items() // フォールバック (Qiita レート制限/ネットワーク断)
		source = "sample"
	}

	all := catalog.Build(items, scorer)
	filtered := catalog.Filter(all, level, topic)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"source":     source,
		"query":      query,
		"count":      len(filtered),
		"total":      len(all),
		"weights":    scorer.Weights,
		"thresholds": scorer.Thresholds,
		"resources":  filtered,
	})
}
