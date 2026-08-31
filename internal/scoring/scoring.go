// Package scoring は記事の「難易度」を複数シグナルの加重和で推定する。
//
// 設計方針 (studytune-context より):
//   - 難易度と品質は別軸。いいね数・ブクマ数は品質であって難易度ではない。
//   - 難易度はトピックごとのベクトルで持つ。単一スカラーは破綻する。
//   - 各シグナルを 0..1 に正規化 → 加重和 → 0..5 にマップ。
//   - 重みは設定値として外出しし、チューニング可能にする。
//   - 正解データが無いので、シグナルの内訳を必ず返す (説明可能性 = 唯一の検証手段)。
package scoring

import (
	"math"
	"regexp"
	"strings"

	"studytune/internal/qiita"
)

// Weights は各シグナルの寄与度。合計が 1 になる必要はない (内部で正規化する)。
// Explanatoriness だけは「丁寧なほど易しい」ため負に効かせる。
type Weights struct {
	Tag             float64 `json:"tag"`
	Prerequisite    float64 `json:"prerequisite"`
	Jargon          float64 `json:"jargon"`
	Explanatoriness float64 `json:"explanatoriness"` // 負の重みとして適用
	ExternalRefs    float64 `json:"externalRefs"`
	CodeRatio       float64 `json:"codeRatio"`
	Length          float64 `json:"length"`
}

// DefaultWeights はチューニング前の初期値。config で上書きできる想定。
var DefaultWeights = Weights{
	Tag:             0.22,
	Prerequisite:    0.20,
	Jargon:          0.20,
	Explanatoriness: 0.10,
	ExternalRefs:    0.13,
	CodeRatio:       0.12,
	Length:          0.03,
}

// LevelThresholds は難易度 (0..5) をレベル区分に落とす境界。
type LevelThresholds struct {
	Intermediate float64 // これ未満は beginner
	Advanced     float64 // これ未満は intermediate、以上は advanced
}

// DefaultThresholds は初期境界。
//
// 加重和は全シグナルが同時に 1.0 になることがほぼ無いため、実データの難易度は
// おおむね 1.0〜4.0 に収まる。3 区分が実際に使えるよう境界を低めに置く。
// 手動正解セットとの相関を取りながらチューニングする前提。
var DefaultThresholds = LevelThresholds{Intermediate: 1.8, Advanced: 3.2}

// Signal は 1 シグナルの評価結果 (UI/検証用の内訳)。
type Signal struct {
	Name         string  `json:"name"`
	Raw          string  `json:"raw"`          // 人が読める根拠 (例: "専門用語 7 種")
	Value        float64 `json:"value"`        // 0..1 に正規化した値
	Weight       float64 `json:"weight"`       // 適用した重み (符号付き)
	Contribution float64 `json:"contribution"` // 0..5 スケールでの寄与ポイント (符号付き)
}

// Result は 1 記事のスコアリング結果。
type Result struct {
	Difficulty float64            `json:"difficulty"` // 0..5
	Level      string             `json:"level"`      // beginner | intermediate | advanced
	LevelLabel string             `json:"levelLabel"` // 初級 | 中級 | 上級
	Topics     map[string]float64 `json:"topics"`     // トピック -> 難易度 0..5
	Signals    []Signal           `json:"signals"`
}

// Scorer は重みと境界を保持してスコアリングを行う。
type Scorer struct {
	Weights    Weights
	Thresholds LevelThresholds

	prereqRe *regexp.Regexp
	headingRe *regexp.Regexp
	imageRe  *regexp.Regexp
	fenceRe  *regexp.Regexp
	linkRe   *regexp.Regexp
}

// NewScorer は既定設定の Scorer を返す。
func NewScorer() *Scorer {
	return NewScorerWith(DefaultWeights, DefaultThresholds)
}

// NewScorerWith は重みと境界を指定して Scorer を返す。
func NewScorerWith(w Weights, t LevelThresholds) *Scorer {
	return &Scorer{
		Weights:    w,
		Thresholds: t,
		prereqRe:   regexp.MustCompile("(?i)" + strings.Join(prerequisitePatterns, "|")),
		headingRe:  regexp.MustCompile(`(?m)^#{1,6}\s`),
		imageRe:    regexp.MustCompile(`!\[[^\]]*\]\([^)]+\)`),
		fenceRe:    regexp.MustCompile("(?s)```.*?```"),
		linkRe:     regexp.MustCompile(`https?://[^\s)]+`),
	}
}

// Score は 1 記事を評価する。item.Body は計算にのみ使い、Result には含めない。
func (s *Scorer) Score(item qiita.Item) Result {
	body := item.Body
	lower := strings.ToLower(body)

	sigs := []Signal{
		s.tagSignal(item.Tags),
		s.prerequisiteSignal(body),
		s.jargonSignal(lower),
		s.explanatorinessSignal(body),
		s.externalRefsSignal(lower),
		s.codeRatioSignal(body),
		s.lengthSignal(body),
	}

	// 正の重みの合計で正規化 (Explanatoriness は負なので除外)。
	posWeightSum := s.Weights.Tag + s.Weights.Prerequisite + s.Weights.Jargon +
		s.Weights.ExternalRefs + s.Weights.CodeRatio + s.Weights.Length
	if posWeightSum == 0 {
		posWeightSum = 1
	}

	var acc float64
	for i := range sigs {
		acc += sigs[i].Value * sigs[i].Weight
		// contribution は 0..5 スケールに直して埋め戻す
		sigs[i].Contribution = round2(sigs[i].Value * sigs[i].Weight / posWeightSum * 5)
	}

	difficulty := clamp(acc/posWeightSum, 0, 1) * 5
	difficulty = round2(difficulty)

	level, label := s.classify(difficulty)

	return Result{
		Difficulty: difficulty,
		Level:      level,
		LevelLabel: label,
		Topics:     topicVector(item.Tags, difficulty),
		Signals:    sigs,
	}
}

func (s *Scorer) classify(d float64) (string, string) {
	switch {
	case d < s.Thresholds.Intermediate:
		return "beginner", "初級"
	case d < s.Thresholds.Advanced:
		return "intermediate", "中級"
	default:
		return "advanced", "上級"
	}
}

// --- 各シグナル (すべて 0..1 を返す) ---

func (s *Scorer) tagSignal(tags []qiita.Tag) Signal {
	v := 0.5
	var hits []string
	for _, t := range tags {
		n := strings.ToLower(t.Name)
		if beginnerTags[n] {
			v -= 0.25
			hits = append(hits, "−"+t.Name)
		}
		if advancedTags[n] {
			v += 0.25
			hits = append(hits, "+"+t.Name)
		}
	}
	raw := "中立"
	if len(hits) > 0 {
		raw = strings.Join(hits, " ")
	}
	return Signal{Name: "tag", Raw: raw, Value: clamp(v, 0, 1), Weight: s.Weights.Tag}
}

func (s *Scorer) prerequisiteSignal(body string) Signal {
	m := s.prereqRe.FindAllString(body, -1)
	n := len(m)
	var v float64
	switch {
	case n == 0:
		v = 0
	case n == 1:
		v = 0.6
	default:
		v = 1
	}
	return Signal{Name: "prerequisite", Raw: countRaw(n, "箇所で前提知識に言及"), Value: v, Weight: s.Weights.Prerequisite}
}

func (s *Scorer) jargonSignal(lowerBody string) Signal {
	distinct := 0
	total := 0
	for _, term := range jargonTerms {
		c := strings.Count(lowerBody, term)
		if c > 0 {
			distinct++
			total += c
		}
	}
	// 種類数を主、密度を従で見る。8 種類で頭打ち。
	byKinds := float64(distinct) / 8.0
	runes := float64(len([]rune(lowerBody)))
	density := 0.0
	if runes > 0 {
		density = float64(total) / runes * 1000 // 1000 文字あたりの出現数
	}
	byDensity := density / 6.0
	v := clamp(0.7*byKinds+0.3*byDensity, 0, 1)
	return Signal{Name: "jargon", Raw: countRaw(distinct, "種の専門用語"), Value: v, Weight: s.Weights.Jargon}
}

func (s *Scorer) explanatorinessSignal(body string) Signal {
	headings := len(s.headingRe.FindAllString(body, -1))
	images := len(s.imageRe.FindAllString(body, -1))
	// 見出し多め + 図多めなら「丁寧」。12 で頭打ち。
	v := clamp(float64(headings+2*images)/12.0, 0, 1)
	return Signal{
		Name:   "explanatoriness",
		Raw:    countRaw(headings, "見出し") + " / " + countRaw(images, "図"),
		Value:  v,
		Weight: -s.Weights.Explanatoriness, // 丁寧なほど易しい
	}
}

func (s *Scorer) externalRefsSignal(lowerBody string) Signal {
	hits := 0
	for _, h := range specRefHosts {
		hits += strings.Count(lowerBody, h)
	}
	v := clamp(float64(hits)/3.0, 0, 1)
	return Signal{Name: "externalRefs", Raw: countRaw(hits, "件の一次資料リンク (RFC/論文/仕様)"), Value: v, Weight: s.Weights.ExternalRefs}
}

func (s *Scorer) codeRatioSignal(body string) Signal {
	fences := s.fenceRe.FindAllString(body, -1)
	codeLines := 0
	for _, f := range fences {
		codeLines += strings.Count(f, "\n")
	}
	totalLines := strings.Count(body, "\n") + 1
	ratio := 0.0
	if totalLines > 0 {
		ratio = float64(codeLines) / float64(totalLines)
	}
	// コード 40% 超で頭打ち。写経系は難易度が上がりやすい。
	v := clamp(ratio/0.4, 0, 1)
	return Signal{Name: "codeRatio", Raw: pctRaw(ratio) + " がコード", Value: v, Weight: s.Weights.CodeRatio}
}

func (s *Scorer) lengthSignal(body string) Signal {
	chars := len([]rune(body))
	// 対数スケール。3000 字で ~0.5、12000 字で ~1.0。単独では効かせない (重み小)。
	v := 0.0
	if chars > 0 {
		v = clamp(math.Log10(float64(chars)/300.0)/1.6, 0, 1)
	}
	return Signal{Name: "length", Raw: countRaw(chars, "文字"), Value: v, Weight: s.Weights.Length}
}

// --- トピックベクトル ---

// topicVector は記事の全体難易度を、認識できたトピックタグそれぞれに割り当てる。
// v1 の割り切り: 記事単位の難易度をトピックに横流しする。
// 将来は「そのトピックへの言及量」で重み付けする。
func topicVector(tags []qiita.Tag, difficulty float64) map[string]float64 {
	out := map[string]float64{}
	for _, t := range tags {
		if canon, ok := topicAliases[strings.ToLower(t.Name)]; ok {
			out[canon] = difficulty
		}
	}
	return out
}

// --- helpers ---

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func countRaw(n int, unit string) string {
	return itoa(n) + " " + unit
}

func pctRaw(ratio float64) string {
	return itoa(int(math.Round(ratio*100))) + "%"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
