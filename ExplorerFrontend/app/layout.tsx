import './globals.css'
import Sidebar from "./components/Sidebar"
import Script from 'next/script'
import Providers from './providers'

export const viewport = {
  themeColor: '#1a1a1a',
}

export const metadata = {
  metadataBase: new URL('https://zondscan.com'),
  title: 'QRL Zond Explorer',
  description: 'QRL ZOND Web3/EVM Compatible Blockchain Explorer - Explore transactions, blocks, smart contracts, and validators on the Quantum Resistant Ledger Proof-of-Stake network',
  keywords: 'QRL, ZOND, blockchain explorer, Web3, EVM, quantum resistant, cryptocurrency, blockchain, smart contracts, proof of stake, validators, transactions, blocks',
  openGraph: {
    title: 'QRL Zond Explorer',
    description: 'QRL ZOND Web3/EVM Compatible Blockchain Explorer - Explore transactions, blocks, smart contracts, and validators on the Quantum Resistant Ledger Proof-of-Stake network',
    type: 'website',
    locale: 'en_US',
    url: 'https://zondscan.com',
    siteName: 'QRL Zond Explorer',
    images: [
      {
        url: '/QRL.png',
        width: 512,
        height: 512,
        alt: 'QRL Logo',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'QRL Zond Explorer',
    description: 'QRL ZOND Web3/EVM Compatible Blockchain Explorer - Explore transactions, blocks, smart contracts, and validators on the Quantum Resistant Ledger Proof-of-Stake network',
    images: ['/QRL.png'],
    creator: '@QRLedger',
    site: '@QRLedger',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
    },
  },
  icons: {
    icon: '/favicon.ico',
    apple: '/QRL.png',
  },
  alternates: {
    canonical: 'https://zondscan.com',
  },
}

interface RootLayoutProps {
  children: React.ReactNode
}

export default function RootLayout({ children }: RootLayoutProps) {
  return (
    <html lang="en" className="dark">
      <head>
        <Script id="schema-org" type="application/ld+json" strategy="beforeInteractive">
          {`
            {
              "@context": "https://schema.org",
              "@type": "WebApplication",
              "name": "QRL ZOND Explorer",
              "description": "QRL ZOND Web3/EVM Compatible Blockchain Explorer - Explore transactions, blocks, smart contracts, and validators on the Quantum Resistant Ledger Proof-of-Stake network",
              "url": "https://zondscan.com",
              "applicationCategory": "Blockchain Explorer",
              "operatingSystem": "All",
              "browserRequirements": "Requires JavaScript",
              "offers": {
                "@type": "Offer",
                "price": "0",
                "priceCurrency": "USD"
              }
            }
          `}
        </Script>
      </head>
      <body className="min-h-screen bg-[#1a1a1a] text-gray-300">
        <Providers>
          <div className="flex min-h-screen">
            <Sidebar />
            <main className="flex-1 ml-64 min-h-screen relative">
              <div className="relative">
                {children}
              </div>
            </main>
          </div>
        </Providers>
      </body>
    </html>
  )
}
