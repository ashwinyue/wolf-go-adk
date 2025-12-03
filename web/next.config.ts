import type { NextConfig } from "next";

const isProd = process.env.NODE_ENV === 'production';

const nextConfig: NextConfig = {
  // 只在生产环境启用静态导出
  ...(isProd && { output: 'export' }),
  basePath: isProd ? '/wolf-go' : '',
  assetPrefix: isProd ? '/wolf-go/' : '',
  images: {
    unoptimized: true,
  },
  trailingSlash: true,
};

export default nextConfig;
