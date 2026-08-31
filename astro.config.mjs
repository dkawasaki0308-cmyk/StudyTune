// @ts-check
import { defineConfig } from 'astro/config';
import vercel from '@astrojs/vercel';
import tailwindcss from '@tailwindcss/vite';

// 疎結合構成: Astro フロント + Vercel 上の Go サーバーレス関数 (/api/*.go)
// Astro は SSR で /api/resources (Go) を叩き、レベル別に整理済みのリソースを描画する
export default defineConfig({
  output: 'server',
  adapter: vercel(),
  vite: {
    plugins: [tailwindcss()],
  },
});
