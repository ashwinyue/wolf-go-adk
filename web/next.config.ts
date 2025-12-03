import type { NextConfig } from "next";

const isGitHubPages = process.env.GITHUB_PAGES === 'true';

const nextConfig: NextConfig = {
  // 只在 GitHub Pages 启用静态导出
  ...(isGitHubPages && { output: 'export' }),
  basePath: isGitHubPages ? '/wolf-go-adk' : '',
  assetPrefix: isGitHubPages ? '/wolf-go-adk/' : '',
  images: {
    unoptimized: true,
  },
  trailingSlash: true,
};

export default nextConfig;
