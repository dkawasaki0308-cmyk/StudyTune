// Package catalog は Qiita 記事をスコアリングし、外部公開用のリソース一覧に組み立てる。
//
// 著作権制約: 公開するのはメタデータ + 自作要約 + 出典 + 元記事リンクのみ。
// 本文の全文保存・転載はしない。
package catalog

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"studytune/internal/qiita"
	"studytune/internal/scoring"
)

// Resource は UI に返す 1 リソース。
type Resource struct {
	Title       string             `json:"title"`
	URL         string             `json:"url"`
	Source      string             `json:"source"` // "Qiita" 等の出典
	Author      string             `json:"author"`
	Likes       int                `json:"likes"` // 品質シグナル (難易度とは別軸)
	PublishedAt time.Time          `json:"publishedAt"`
	Tags        []string           `json:"tags"`
	Summary     string             `json:"summary"` // 自作の抽出要約 (LLM 要約への差し替え余地あり)
	Difficulty  float64            `json:"difficulty"`
	Level       string             `json:"level"`
	LevelLabel  string             `json:"levelLabel"`
	Topics      map[string]float64 `json:"topics"`
	Signals     []scoring.Signal   `json:"signals"`
}

// Build は取得済み記事をスコアリングして Resource 一覧に変換する。
// 難易度の降順で返す。
func Build(items []qiita.Item, s *scoring.Scorer) []Resource {
	out := make([]Resource, 0, len(items))
	for _, it := range items {
		r := s.Score(it)
		out = append(out, Resource{
			Title:       strings.TrimSpace(it.Title),
			URL:         it.URL,
			Source:      "Qiita",
			Author:      it.User.ID,
			Likes:       it.LikesCount,
			PublishedAt: it.CreatedAt,
			Tags:        tagNames(it.Tags),
			Summary:     summarize(it.Body),
			Difficulty:  r.Difficulty,
			Level:       r.Level,
			LevelLabel:  r.LevelLabel,
			Topics:      r.Topics,
			Signals:     r.Signals,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Difficulty > out[j].Difficulty })
	return out
}

// Filter は level / topic で絞り込む。空文字は「指定なし」。
func Filter(rs []Resource, level, topic string) []Resource {
	if level == "" && topic == "" {
		return rs
	}
	out := make([]Resource, 0, len(rs))
	for _, r := range rs {
		if level != "" && r.Level != level {
			continue
		}
		if topic != "" {
			if _, ok := r.Topics[topic]; !ok {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func tagNames(tags []qiita.Tag) []string {
	ns := make([]string, len(tags))
	for i, t := range tags {
		ns[i] = t.Name
	}
	return ns
}

var (
	mdFence    = regexp.MustCompile("(?s)```.*?```")
	mdInline   = regexp.MustCompile("`[^`]*`")
	mdImage    = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	mdLink     = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	mdHeading  = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	mdLeading  = regexp.MustCompile(`^[\s>*#|・\-]+`) // 行頭の箇条書き/引用/見出し記号
	mdBold     = regexp.MustCompile(`\*\*|__`)
	wsRun      = regexp.MustCompile(`\s+`)
	sentSplit  = regexp.MustCompile(`(?:。|．|\.\s|\n)`)
)

// summarize は本文 Markdown から最初のまとまった説明文を抽出する。
// あくまで暫定。将来は LLM による 2〜3 文の自作要約に差し替える。
func summarize(body string) string {
	t := mdFence.ReplaceAllString(body, " ")
	t = mdImage.ReplaceAllString(t, " ")
	t = mdLink.ReplaceAllString(t, "$1")
	t = mdInline.ReplaceAllString(t, " ")
	t = mdHeading.ReplaceAllString(t, "")

	for _, raw := range sentSplit.Split(t, -1) {
		line := mdLeading.ReplaceAllString(raw, "")
		line = mdBold.ReplaceAllString(line, "")
		line = strings.TrimSpace(wsRun.ReplaceAllString(line, " "))
		if len([]rune(line)) < 16 {
			continue // 見出し残り・箇条書きの断片を捨てる
		}
		return truncateRunes(line, 120)
	}
	return ""
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
