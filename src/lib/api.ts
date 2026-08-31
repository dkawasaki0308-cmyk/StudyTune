// フロントから Go API を呼ぶ薄いラッパ。SSR (index.astro) からサーバー側で使う。
import type { Level, ResourcesResponse } from './types';
import { SAMPLE_RESPONSE } from './sample';

export interface ResourceQuery {
  level?: Level | '';
  topic?: string;
  q?: string;
  per?: number;
}

/**
 * /api/resources を叩く。Go 関数が動いていない (ローカルの素の astro dev など) 場合は
 * 埋め込みサンプルにフォールバックしてページを必ず描画する。
 */
export async function fetchResources(
  origin: string,
  query: ResourceQuery = {},
): Promise<ResourcesResponse & { degraded?: boolean }> {
  const params = new URLSearchParams();
  if (query.level) params.set('level', query.level);
  if (query.topic) params.set('topic', query.topic);
  if (query.q) params.set('q', query.q);
  if (query.per) params.set('per', String(query.per));

  const url = `${origin}/api/resources${params.toString() ? `?${params}` : ''}`;

  try {
    const res = await fetch(url, { headers: { accept: 'application/json' } });
    if (!res.ok) throw new Error(`api ${res.status}`);
    return (await res.json()) as ResourcesResponse;
  } catch {
    return { ...filterSample(query), degraded: true };
  }
}

function filterSample(query: ResourceQuery): ResourcesResponse {
  let resources = SAMPLE_RESPONSE.resources;
  if (query.level) resources = resources.filter((r) => r.level === query.level);
  if (query.topic) resources = resources.filter((r) => query.topic! in r.topics);
  return { ...SAMPLE_RESPONSE, resources, count: resources.length };
}
