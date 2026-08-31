package scoring

import (
	"testing"
	"time"

	"studytune/pkg/qiita"
)

func tag(names ...string) []qiita.Tag {
	ts := make([]qiita.Tag, len(names))
	for i, n := range names {
		ts[i] = qiita.Tag{Name: n}
	}
	return ts
}

// 平易な入門記事は初級に寄るはず。
func TestScore_BeginnerArticle(t *testing.T) {
	s := NewScorer()
	item := qiita.Item{
		Title:     "はじめての Go: Hello World まで",
		CreatedAt: time.Now(),
		Tags:      tag("Go", "初心者"),
		Body: `# はじめに
Go を初めて触る人向けの記事です。
## インストール
公式サイトからインストーラを落とします。
## Hello World
` + "```go\npackage main\nfunc main() { println(\"hi\") }\n```" + `
以上です。おつかれさまでした。`,
	}
	r := s.Score(item)
	if r.Level != "beginner" {
		t.Fatalf("expected beginner, got %s (difficulty=%.2f)", r.Level, r.Difficulty)
	}
}

// 前提知識・専門用語・一次資料が多い記事は上級に寄るはず。
func TestScore_AdvancedArticle(t *testing.T) {
	s := NewScorer()
	item := qiita.Item{
		Title:     "Go スケジューラの内部実装とエスケープ解析",
		CreatedAt: time.Now(),
		Tags:      tag("Go", "内部実装", "並行処理"),
		Body: `本記事では goroutine の基礎は理解している前提で進めます。
ゴルーチンのスケジューリングは GMP モデルに基づき、ミューテックスやアトミック操作、
メモリバリアの理解が必要です。エスケープ解析によりヒープとスタックの割り当てが決まります。
詳細は https://research.swtch.com/gomm と https://go.dev/ref/spec を参照。
システムコールと epoll の関係、デッドロックとレースコンディションについても触れます。
計算量はオーダー記法で議論します。`,
	}
	r := s.Score(item)
	if r.Level != "advanced" {
		t.Fatalf("expected advanced, got %s (difficulty=%.2f)", r.Level, r.Difficulty)
	}
}

// 難易度の順序関係が保たれること (相対順位が検証の主目的)。
func TestScore_Ordering(t *testing.T) {
	s := NewScorer()
	easy := s.Score(qiita.Item{
		Tags: tag("Go", "初心者"),
		Body: "# 入門\nはじめての人向け。\n## 手順\nインストールするだけ。\n",
	})
	hard := s.Score(qiita.Item{
		Tags: tag("Go", "内部実装"),
		Body: "goroutine は理解している前提。ゴルーチン ミューテックス アトミック メモリバリア " +
			"システムコール epoll デッドロック 計算量 オーダー記法。https://go.dev/ref/spec",
	})
	if !(easy.Difficulty < hard.Difficulty) {
		t.Fatalf("expected easy(%.2f) < hard(%.2f)", easy.Difficulty, hard.Difficulty)
	}
}

// トピックベクトルが認識タグごとに難易度を持つこと。
func TestScore_TopicVector(t *testing.T) {
	s := NewScorer()
	r := s.Score(qiita.Item{Tags: tag("Go", "Docker", "全く無関係なタグ"), Body: "短い本文"})
	if _, ok := r.Topics["go"]; !ok {
		t.Errorf("expected 'go' topic, got %v", r.Topics)
	}
	if _, ok := r.Topics["docker"]; !ok {
		t.Errorf("expected 'docker' topic, got %v", r.Topics)
	}
	if len(r.Topics) != 2 {
		t.Errorf("expected 2 recognized topics, got %v", r.Topics)
	}
}
