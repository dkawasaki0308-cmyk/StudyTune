// Package sample は Qiita API が使えないとき (レート制限・ネットワーク断) の
// フォールバック用サンプル記事を埋め込みで提供する。
//
// サンプルは実際の記事本文ではなく、パイプライン動作確認用の代表的な要約テキスト。
// これらもリアルタイムに scoring を通すので、難易度推定の挙動をそのまま見られる。
package sample

import (
	_ "embed"
	"encoding/json"

	"studytune/internal/qiita"
)

//go:embed items.json
var itemsJSON []byte

// Items は埋め込み済みのサンプル記事を返す。
func Items() []qiita.Item {
	var items []qiita.Item
	_ = json.Unmarshal(itemsJSON, &items)
	return items
}
