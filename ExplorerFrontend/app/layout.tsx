import './globals.css'
import Sidebar from "./components/Sidebar"
import Script from 'next/script'
import Providers from './providers'
import { Metadata } from 'next';
import { sharedMetadata } from './lib/seo/metaData';
import Footer from './components/Footer';

export const metadata: Metadata = {
  ...sharedMetadata,
  title: 'ZondScan | QRL Zond EVM Blockchain Explorer (Quantum-Safe)',
  description:
    'ZondScan is an explorer for QRL Zond – the first EVM-compatible blockchain secured with post-quantum cryptography. Track real-time transactions, smart contracts, and validator activity on a next-gen proof-of-stake network.',
  openGraph: {
    ...sharedMetadata.openGraph,
    title: 'ZondScan | QRL Zond EVM Blockchain Explorer (Quantum-Safe)',
    description:
      'Discover QRL Zond with ZondScan – the secure, quantum-resistant, and EVM-compatible explorer for smart contracts, blocks, and transactions. Built for the future of decentralized apps.',
    url: 'https://zondscan.com',
  },
  twitter: {
    ...sharedMetadata.twitter,
    title: 'ZondScan | QRL Zond EVM Blockchain Explorer (Quantum-Safe)',
    description:
      'Explore QRL Zond with ZondScan – the first post-quantum EVM-compatible blockchain. View live data on smart contracts, PoS validators, and secure Web3 apps.',
  },
  
  
};


export const viewport = {
  themeColor: '#1a1a1a',
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
              "name": "QRL Zond Explorer",
              "description": "QRL Zond Web3 EVM Compatible Blockchain Explorer - Explore transactions, blocks, smart contracts, and validators on the Quantum Resistant Ledger Proof-of-Stake network",
              "url": "https://zondscan.com",
              "applicationCategory": "Blockchain Explorer",
              "operatingSystem": "All",
              "browserRequirements": "Requires JavaScript",
              "offers": {
                "@type": "Offer",
                "price": "0",
                "priceCurrency": "USD"
              },
              "relatedApplication": {
                "@type": "SoftwareApplication",
                "name": "QRL Zond Web Wallet",
                "url": "https://qrlwallet.com",
                "applicationCategory": "Blockchain Wallet",
                "operatingSystem": "All"
              },
              "hasPart": [
                {
                  "@type": "WebPage",
                  "name": "Latest Transactions",
                  "description": "View recent Transactions",
                  "url": "https://zondscan.com/transactions/1"
                },
                {
                  "@type": "WebPage",
                  "name": "Pending Transactions",
                  "description": "View pending transactions",
                  "url": "https://zondscan.com/pending/1"
                },
                {
                  "@type": "WebPage",
                  "name": "Latest Blocks",
                  "description": "View all Blocks",
                  "url": "https://zondscan.com/blocks/1"
                },
                {
                  "@type": "WebPage",
                  "name": "Smart Contracts",
                  "description": "Explore QRL contracts",
                  "url": "https://zondscan.com/contracts"
                },
                {
                  "@type": "WebPage",
                  "name": "Validators",
                  "description": "Network Validators",
                  "url": "https://zondscan.com/validators"
                },
                {
                  "@type": "WebPage",
                  "name": "Balance Checker",
                  "description": "Check Account balance",
                  "url": "https://zondscan.com/checker"
                },
                {
                  "@type": "WebPage",
                  "name": "Unit Converter",
                  "description": "Convert QRL currencies",
                  "url": "https://zondscan.com/converter"
                },
                {
                  "@type": "WebPage",
                  "name": "Richlist",
                  "description": "Top QRL holders",
                  "url": "https://zondscan.com/richlist"
                }
              ]
            }
          `}
        </Script>
      </head>
      <body className="min-h-screen bg-[#1a1a1a] text-gray-300">
        <Providers>
          <div className="flex min-h-screen">
            <Sidebar />
            <main className="flex-1 lg:ml-64 min-h-screen relative transition-all duration-300 mt-[72px] lg:mt-4">
              <div className="relative">
                {children}
                <Footer />
              </div>
            </main>
          </div>
          
        </Providers>
      </body>
    </html>
  )
}
