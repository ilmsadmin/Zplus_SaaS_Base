import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { ApolloProvider } from '@/components/providers/apollo-provider';
import { Toaster } from 'react-hot-toast';
import '@/styles/globals.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: {
    default: 'Zplus SaaS Platform',
    template: '%s | Zplus SaaS',
  },
  description: 'Multi-tenant SaaS platform with advanced features',
  keywords: ['SaaS', 'multi-tenant', 'platform', 'business'],
  authors: [{ name: 'Zplus Team' }],
  creator: 'Zplus',
  metadataBase: new URL('https://zplus.io'),
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://zplus.io',
    title: 'Zplus SaaS Platform',
    description: 'Multi-tenant SaaS platform with advanced features',
    siteName: 'Zplus',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Zplus SaaS Platform',
    description: 'Multi-tenant SaaS platform with advanced features',
    creator: '@zplus',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  verification: {
    google: 'your-google-verification-code',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="icon" href="/favicon.ico" />
        <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
        <link rel="manifest" href="/manifest.json" />
        <meta name="theme-color" content="#3b82f6" />
      </head>
      <body className={inter.className}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <ApolloProvider>
            <div className="relative min-h-screen bg-white dark:bg-gray-950 text-gray-900 dark:text-gray-100">
              {children}
              <Toaster
                position="top-right"
                toastOptions={{
                  duration: 4000,
                  style: {
                    background: 'rgb(255 255 255)',
                    color: 'rgb(17 24 39)',
                    border: '1px solid rgb(229 231 235)',
                  },
                  success: {
                    iconTheme: {
                      primary: 'var(--success-500)',
                      secondary: 'var(--background)',
                    },
                  },
                  error: {
                    iconTheme: {
                      primary: 'var(--error-500)',
                      secondary: 'var(--background)',
                    },
                  },
                }}
              />
            </div>
          </ApolloProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
