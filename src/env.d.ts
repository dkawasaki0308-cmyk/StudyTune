/// <reference path="../.astro/types.d.ts" />
/// <reference types="astro/client" />

interface ImportMetaEnv {
  /** Qiita API v2 のアクセストークン (任意。無いと 60req/h) */
  readonly QIITA_TOKEN?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
