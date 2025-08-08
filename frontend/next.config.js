/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    // typedRoutes was moved to stable in Next.js 15
    optimizePackageImports: ['lucide-react', '@headlessui/react'],
  },
  images: {
    domains: ['localhost', 'zplus.io'],
    formats: ['image/webp', 'image/avif'],
  },
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY,
  },
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
        ],
      },
    ]
  },
  async rewrites() {
    return [
      {
        source: '/api/graphql',
        destination: `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/graphql`,
      },
    ]
  },
}

module.exports = nextConfig
