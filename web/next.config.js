/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  poweredByHeader: false,
  reactStrictMode: true,
  allowedDevOrigins: ['127.0.0.1'],
  env: {
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL,
    NEXT_PUBLIC_API_RECIPES_URL: process.env.NEXT_PUBLIC_API_RECIPES_URL,
    NEXT_PUBLIC_API_USERS_URL: process.env.NEXT_PUBLIC_API_USERS_URL,
    NEXT_PUBLIC_API_FAVORITES_URL: process.env.NEXT_PUBLIC_API_FAVORITES_URL,
    NEXT_PUBLIC_API_RECIPES_PORT: process.env.NEXT_PUBLIC_API_RECIPES_PORT || '8081',
    NEXT_PUBLIC_API_USERS_PORT: process.env.NEXT_PUBLIC_API_USERS_PORT || '8082',
    NEXT_PUBLIC_API_FAVORITES_PORT: process.env.NEXT_PUBLIC_API_FAVORITES_PORT || '8080',
  },
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=()' },
        ],
      },
    ]
  },
}

module.exports = nextConfig
