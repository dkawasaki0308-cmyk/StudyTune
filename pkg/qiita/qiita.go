// Package qiita は Qiita API v2 から記事を取得する薄いクライアント。
//
// 著作権上の制約: 本文 (Body) はスコアリング計算にのみ使用し、
// レスポンスやストレージには載せない。外部に出すのはメタデータのみ
// (タイトル・URL・タグ・著者・いいね数・公開日 + 自作の要約/スコア)。
package qiita

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const apiBase = "https://qiita.com/api/v2"

// Tag は Qiita 記事に付与されたタグ。
type Tag struct {
	Name string `json:"name"`
}

// User は記事の著者。
type User struct {
	ID string `json:"id"`
}

// Item は Qiita API v2 の記事 1 件。Body は計算専用 (JSON 出力しない)。
type Item struct {
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	LikesCount int       `json:"likes_count"`
	CreatedAt  time.Time `json:"created_at"`
	Tags       []Tag     `json:"tags"`
	User       User      `json:"user"`
	Body       string    `json:"body"` // Markdown 全文。スコアリング後に破棄する。
}

// Client は Qiita API のクライアント。
type Client struct {
	http  *http.Client
	token string
}

// New は環境変数 QIITA_TOKEN を読んでクライアントを生成する。
// トークンが無くても動くが 60req/h に制限される (あれば 1000req/h)。
func New() *Client {
	return &Client{
		http:  &http.Client{Timeout: 10 * time.Second},
		token: os.Getenv("QIITA_TOKEN"),
	}
}

// SearchOptions は記事検索のパラメータ。
type SearchOptions struct {
	Query   string // 例: "tag:Go stocks:>10"
	PerPage int    // 1..100
	Page    int    // 1..
}

// Search は条件に合う記事を取得する。
func (c *Client) Search(ctx context.Context, opt SearchOptions) ([]Item, error) {
	if opt.PerPage <= 0 || opt.PerPage > 100 {
		opt.PerPage = 20
	}
	if opt.Page <= 0 {
		opt.Page = 1
	}

	q := url.Values{}
	q.Set("query", opt.Query)
	q.Set("per_page", strconv.Itoa(opt.PerPage))
	q.Set("page", strconv.Itoa(opt.Page))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/items?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qiita request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qiita api status %d", res.StatusCode)
	}

	var items []Item
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("qiita decode: %w", err)
	}
	return items, nil
}
