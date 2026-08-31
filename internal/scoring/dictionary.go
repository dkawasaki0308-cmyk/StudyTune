package scoring

// beginnerTags はタグ名 (小文字) が難易度を下げるもの。
var beginnerTags = map[string]bool{
	"初心者":       true,
	"初学者":       true,
	"入門":        true,
	"beginner":  true,
	"初心者向け":     true,
	"プログラミング初心者": true,
	"新人プログラマ応援": true,
	"備忘録":       true,
	"個人開発":      true,
	"ポエム":       true,
}

// advancedTags はタグ名 (小文字) が難易度を上げるもの。
var advancedTags = map[string]bool{
	"内部実装":            true,
	"low-level":       true,
	"lowlevel":        true,
	"コンパイラ":           true,
	"os自作":            true,
	"アルゴリズム":          true,
	"パフォーマンスチューニング":    true,
	"設計":              true,
	"アーキテクチャ":         true,
	"分散システム":          true,
	"ポインタ":            true,
	"メモリ管理":           true,
	"並行処理":            true,
	"golang-internal": true,
}

// jargonTerms は専門用語辞書。本文中の出現で難易度が上がる。
// 日本語/英語の表記ゆれを両方入れる。
var jargonTerms = []string{
	"ゴルーチン", "goroutine", "ミューテックス", "mutex", "セマフォ", "semaphore",
	"チャネル", "排他制御", "デッドロック", "deadlock", "レースコンディション", "race condition",
	"アトミック", "atomic", "メモリバリア", "memory barrier", "キャッシュライン",
	"ヒープ", "スタックフレーム", "エスケープ解析", "escape analysis", "ガベージコレクタ", "gc",
	"システムコール", "syscall", "カーネル", "epoll", "kqueue", "ファイルディスクリプタ",
	"抽象構文木", "ast", "字句解析", "構文解析", "パーサ", "コンパイラ",
	"計算量", "オーダー記法", "償却", "amortized", "b-tree", "btree", "赤黒木",
	"トランザクション分離レベル", "mvcc", "wal", "インデックススキャン", "クエリプランナ", "explain analyze",
	"冪等", "idempotent", "結合度", "凝集度", "共変", "反変", "モナド",
	"tcp", "輻輳制御", "スループット", "レイテンシ", "rtt", "handshake",
	"型パラメータ", "ジェネリクス", "型推論", "単一化", "部分型",
}

// prerequisitePatterns は「前提知識の明示」を検出する正規表現ソース。
var prerequisitePatterns = []string{
	`前提知識`,
	`前提として`,
	`を(?:理解|把握|習得)している(?:こと|前提|必要)`,
	`を知っている(?:こと|前提)`,
	`(?:基礎|基本)は(?:分かって|理解して|習得して)いる`,
	`触ったことがある(?:こと|前提)`,
	`本記事では(?:.{0,10})?説明しません`,
	`詳細は割愛`,
	`assume(?:s)? (?:you|the reader) (?:know|are familiar)`,
}

// specRefHosts は「重厚な一次資料」を指すホスト/文字列。多いほど難易度が上がる。
var specRefHosts = []string{
	"datatracker.ietf.org", "rfc-editor.org", "www.rfc-editor.org",
	"tools.ietf.org", "/rfc",
	"arxiv.org", "dl.acm.org", "ieeexplore.ieee.org", "dblp.org",
	"go.dev/ref/spec", "go.dev/blog", "research.swtch.com",
	"man7.org", "pubs.opengroup.org", "w3.org/TR", "tc39.es",
	"postgresql.org/docs", "sqlite.org/", "kernel.org",
}

// topicAliases はタグ名 (小文字) を正規のトピック名に寄せる。
var topicAliases = map[string]string{
	"go": "go", "golang": "go",
	"docker": "docker", "docker-compose": "docker", "コンテナ": "docker",
	"kubernetes": "kubernetes", "k8s": "kubernetes",
	"postgresql": "sql", "postgres": "sql", "mysql": "sql", "sql": "sql", "rdb": "sql", "データベース": "sql",
	"react": "react", "nextjs": "react", "next.js": "react",
	"typescript": "typescript", "ts": "typescript",
	"javascript": "javascript", "js": "javascript",
	"python": "python", "rust": "rust", "c++": "cpp", "cpp": "cpp",
	"gin": "go", "grpc": "grpc", "protobuf": "grpc",
	"astro": "astro", "tailwindcss": "css", "css": "css",
	"aws": "aws", "gcp": "gcp", "terraform": "iac",
	"redis": "redis", "elasticsearch": "search", "meilisearch": "search",
	"linux": "linux", "network": "network", "ネットワーク": "network",
	"algorithm": "algorithm", "アルゴリズム": "algorithm", "競技プログラミング": "algorithm",
}
