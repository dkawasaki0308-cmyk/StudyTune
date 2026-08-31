# StudyTune

プログラミング学習リソース（記事）を**難易度で自動分類**し、初級・中級・上級で整理する
ナビゲーションサイト。学習リソースが公式ドキュメント・個人ブログ・参考書・動画に散在し、
「自分の段階で何を見るべきか」の判断自体が学習の障壁になっている、という課題に対して、
言語・技術ごとにレベル別で整理して「調律」する。

個人開発ポートフォリオ。この MVP はインターン選考用に、フルスコープ（下記）を
デプロイ可能な最小構成へ絞って構築したもの。

## デモ

- サイト: （デプロイ後に記載）
- Qiita API が制限中／ローカル実行時は、埋め込みサンプル記事へ自動フォールバックして
  同じスコアリングパイプラインを通す（`source: "sample"` をバナー表示）

## アーキテクチャ

```
┌─────────────────┐        ┌──────────────────────────────┐
│  Astro (SSR)    │  fetch │  Go serverless (Vercel)       │
│  Tailwind v4    │───────▶│  /api/resources  /api/feedback │
│  レベル/トピック  │  JSON  │                              │
│  絞り込み UI     │◀───────│  Qiita API v2 ─▶ scoring       │
└─────────────────┘        └──────────────────────────────┘
        フロント・API を疎結合（将来のモバイル展開の余地を残す）
```

| レイヤ | 技術 | 補足 |
|---|---|---|
| フロント | Astro 7 + TypeScript + Tailwind CSS v4 | SSR（`@astrojs/vercel`）。絞り込みはクエリパラメータ駆動で JS 最小 |
| API | Go（標準 `net/http`） | Vercel の Go サーバーレス関数。`api/index.go` 1 本に集約しパスで内部ディスパッチ |
| データ源 | Qiita API v2 | 保存するのはメタデータ + 自作要約のみ。本文の転載はしない |
| 難易度推定 | 複数シグナルの加重和 | `internal/scoring/`。重み・境界は設定値で外出し |

技術選定の理由は [`docs/architecture.md`](docs/architecture.md)、選考で説明する技術判断は
[`docs/es-notes.md`](docs/es-notes.md) に整理。

## ディレクトリ

```
api/index.go              Vercel の Go 関数エントリ（/api/* を内部ディスパッチ）
internal/
  api/                    エンドポイント本体（resources / feedback）
  qiita/                  Qiita API v2 クライアント
  scoring/                難易度スコアラー（7 シグナルの加重和）＋テスト
  catalog/                記事 → 公開用 Resource への組み立て（自作要約を含む）
  sample/                 Qiita 失敗時のフォールバック記事（//go:embed）
  httpx/                  JSON レスポンス共通ヘルパ
src/
  pages/index.astro       レベル/トピック別のリソース一覧（SSR）
  components/             ResourceCard / LevelBadge / FeedbackButtons
  lib/                    API 型・fetch ラッパ・ローカル用サンプル
```

## 開発

```bash
# 依存
npm install

# フロントのみ（Go 関数は動かない → サンプルにフォールバック）
npm run dev            # http://localhost:4321

# フル構成（Astro + Go 関数）
vercel dev             # 要 Vercel CLI ログイン

# Go
go test ./...
go build ./...
```

Qiita トークン（任意、無いと 60req/h）は `.env` に `QIITA_TOKEN=...`（`.env.example` 参照）。
Vercel では環境変数 `QIITA_TOKEN` を設定する。

## デプロイ

Vercel。`vercel.json` で `/api/(.*)` を `api/index.go` に rewrite している。
`go.mod` は Vercel の Go ランタイム互換のため `go 1.23` に固定。

## スコープ

**フルスコープ**（別途 100〜120h 想定）: Qiita 以外のソース追加（Zenn / はてブ / openBD / YouTube）、
PostgreSQL + pgx + sqlc による永続化、Elo 方式のユーザーレベル動的更新、
手動正解セットによる精度検証基盤。

**この MVP に含むもの**: Qiita 1 ソース、ルールベースの難易度推定、シグナル内訳の開示、
明示フィードバックの受け口、レベル/トピック絞り込み UI、Vercel 1 デプロイ。

**あえて落としたもの**: DB（Vercel 1 デプロイに収めるため）、動的レベル更新、凝った UI。
理由は [`docs/es-notes.md`](docs/es-notes.md)。
