// Go API (/api/resources) のレスポンスに対応する型。
// internal/catalog/catalog.go と internal/scoring/scoring.go の JSON タグに一致させる。
// 将来 gRPC/Protobuf か Hono RPC で型同期する余地あり（今は手動同期）。

export type Level = 'beginner' | 'intermediate' | 'advanced';

export interface Signal {
  name: string;
  raw: string;
  value: number; // 0..1
  weight: number; // 符号付き
  contribution: number; // 0..5 スケールでの寄与（符号付き）
}

export interface Resource {
  title: string;
  url: string;
  source: string;
  author: string;
  likes: number;
  publishedAt: string;
  tags: string[];
  summary: string;
  difficulty: number; // 0..5
  level: Level;
  levelLabel: string; // 初級 | 中級 | 上級
  topics: Record<string, number>;
  signals: Signal[];
}

export interface ResourcesResponse {
  source: 'qiita' | 'sample';
  query: string;
  count: number;
  total: number;
  weights: Record<string, number>;
  thresholds: { Intermediate: number; Advanced: number };
  resources: Resource[];
}

export const LEVELS: { key: Level; label: string }[] = [
  { key: 'beginner', label: '初級' },
  { key: 'intermediate', label: '中級' },
  { key: 'advanced', label: '上級' },
];
