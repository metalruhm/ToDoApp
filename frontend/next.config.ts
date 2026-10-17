import type { NextConfig } from "next";

// /api/... を API サーバーへ転送する先（ビルド時に固定される）
const apiOrigin = process.env.API_ORIGIN ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  /* config options here */
  output: "standalone", // Docker 用に、実行に必要な最小限のファイルだけを出力する
  experimental: {
    agentFeedback: true,
  },
  cacheComponents: true,
  partialPrefetching: true,
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
  // 本番: NEXT_PUBLIC_API_URL=/api にすると、ブラウザは画面と同じオリジンの /api を呼び、
  // ここで API へ中継される（別サイト扱いの Cookie / CORS の問題を避けられる）。
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiOrigin}/:path*` }];
  },
};

export default nextConfig;
