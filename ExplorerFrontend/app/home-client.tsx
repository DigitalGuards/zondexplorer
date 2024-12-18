'use client';

import * as React from 'react';
import axios from "axios";
import { formatNumber, formatNumberWithCommas, toFixed } from "./lib/helpers";
import config from "../config.js"
import SearchBar from "./components/SearchBar"

interface StatsData {
  value: string;
  isLoading: boolean;
  error: boolean;
}

interface DashboardData {
  walletCount: StatsData;
  volume: StatsData;
  blockHeight: StatsData;
  totalTransactions: StatsData;
  marketCap: StatsData;
  currentPrice: StatsData;
  syncing: boolean;
  dataInitialized: boolean;
}

export default function HomeClient() {
  const [data, setData] = React.useState<DashboardData>({
    walletCount: { value: "0", isLoading: true, error: false },
    volume: { value: "0", isLoading: true, error: false },
    blockHeight: { value: "0", isLoading: true, error: false },
    totalTransactions: { value: "0", isLoading: true, error: false },
    marketCap: { value: "0", isLoading: true, error: false },
    currentPrice: { value: "0", isLoading: true, error: false },
    syncing: true,
    dataInitialized: false
  });

  React.useEffect(() => {
    async function fetchData() {
      try {
        if (config.handlerUrl) {
          const [overviewResponse, latestBlockResponse, txsResponse] = await Promise.allSettled([
            axios.get(config.handlerUrl + "/overview"),
            axios.get(config.handlerUrl + "/latestblock"),
            axios.get(config.handlerUrl + "/txs?page=1")
          ]);

          setData(prevData => {
            const newData = { ...prevData };

            // Handle overview response
            if (overviewResponse.status === 'fulfilled') {
              newData.walletCount.value = overviewResponse.value.data.countwallets?.toString() || "0";
              newData.volume.value = overviewResponse.value.data.volume?.toString() || "0";
              newData.marketCap.value = overviewResponse.value.data.marketcap?.toString() || "0";
              newData.currentPrice.value = overviewResponse.value.data.currentPrice?.toString() || "0";
              newData.syncing = overviewResponse.value.data.status?.syncing ?? true;
              newData.dataInitialized = overviewResponse.value.data.status?.dataInitialized ?? false;
            } else {
              newData.walletCount.error = true;
              newData.volume.error = true;
              newData.marketCap.error = true;
              newData.currentPrice.error = true;
            }

            // Handle latest block response
            if (latestBlockResponse.status === 'fulfilled') {
              newData.blockHeight.value = latestBlockResponse.value.data.response?.[0]?.result?.number?.toString() || "0";
            } else {
              newData.blockHeight.error = true;
            }

            // Handle transactions response
            if (txsResponse.status === 'fulfilled') {
              newData.totalTransactions.value = txsResponse.value.data.total?.toString() || "0";
            } else {
              newData.totalTransactions.error = true;
            }

            // Update loading states
            newData.walletCount.isLoading = false;
            newData.volume.isLoading = false;
            newData.blockHeight.isLoading = false;
            newData.totalTransactions.isLoading = false;
            newData.marketCap.isLoading = false;
            newData.currentPrice.isLoading = false;

            return newData;
          });
        }
      } catch (error) {
        console.error("Failed to fetch overview data:", error);
      }
    }

    fetchData();
  }, []);

  const blockchainStats = [
    {
      data: formatNumberWithCommas(data.walletCount.value),
      title: "Network Bagholder Count",
      loading: data.walletCount.isLoading,
      error: data.walletCount.error,
      icon: (
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mb-2 text-[#ffa729]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
      )
    },
    {
      data: toFixed(parseFloat(data.volume.value)) + " QRL",
      title: "Daily Transactions Volume",
      loading: data.volume.isLoading,
      error: data.volume.error,
      icon: (
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mb-2 text-[#ffa729]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      )
    },
    {
      data: formatNumberWithCommas(data.blockHeight.value),
      title: "Block Height",
      loading: data.blockHeight.isLoading,
      error: data.blockHeight.error,
      icon: (
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mb-2 text-[#ffa729]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
      )
    },
    {
      data: formatNumberWithCommas(data.totalTransactions.value),
      title: "Total Transactions",
      loading: data.totalTransactions.isLoading,
      error: data.totalTransactions.error,
      icon: (
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mb-2 text-[#ffa729]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
        </svg>
      )
    }
  ];

  const financialStats = [
    {
      data: "$" + formatNumberWithCommas(data.marketCap.value),
      title: "Market Cap",
      loading: data.marketCap.isLoading,
      error: data.marketCap.error,
      icon: (
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mb-2 text-[#ffa729]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      )
    },
    {
      data: "$" + parseFloat(data.currentPrice.value).toFixed(4),
      title: "Current Price",
      loading: data.currentPrice.isLoading,
      error: data.currentPrice.error,
      icon: (
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 mb-2 text-[#ffa729]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      )
    }
  ];

  const StatCard = ({ item }: { item: any }) => (
    <div className={`relative overflow-hidden rounded-2xl 
                   bg-gradient-to-br from-[#2d2d2d] to-[#1f1f1f]
                   border border-[#3d3d3d] shadow-xl
                   hover:border-[#ffa729] transition-all duration-300
                   group ${!data.dataInitialized ? 'opacity-50' : ''}`}>
      <div className="relative p-6 text-center min-h-[160px] flex flex-col justify-center">
        {item.loading ? (
          <div className="flex flex-col items-center justify-center space-y-2">
            <div className="w-32 h-8 bg-gray-700/50 rounded animate-pulse"></div>
            <div className="w-24 h-4 bg-gray-700/50 rounded animate-pulse"></div>
          </div>
        ) : item.error ? (
          <div className="flex flex-col items-center justify-center text-red-400">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm">Failed to load data</span>
          </div>
        ) : (
          <>
            <div className="flex justify-center">{item.icon}</div>
            <h4 className="text-3xl font-bold mb-3 text-[#ffa729] 
                        group-hover:scale-110 transition-transform duration-300">
              {item.data}
            </h4>
            <p className="text-sm text-gray-300 font-medium">
              {item.title}
            </p>
          </>
        )}
      </div>
    </div>
  );

  return (
    <div className="min-h-screen">
      <div className="max-w-[1200px] mx-auto p-8">
        <div className="mb-10">
          <SearchBar />
        </div>
        
        {!data.dataInitialized && (
          <div className="mb-8 p-4 rounded-lg bg-yellow-500/10 border border-yellow-500/20 text-yellow-500">
            <div className="flex items-center">
              <svg className="animate-spin -ml-1 mr-3 h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>Initializing explorer data... This may take a few minutes.</span>
            </div>
          </div>
        )}
        
        <div className="space-y-8">
          {/* Blockchain Stats */}
          <div>
            <h2 className="text-xl font-bold mb-4 text-[#ffa729]">Blockchain Statistics</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {blockchainStats.map((item, idx) => (
                <StatCard key={idx} item={item} />
              ))}
            </div>
          </div>

          {/* Financial Stats */}
          <div>
            <h2 className="text-xl font-bold mb-4 text-[#ffa729]">Financial Statistics</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {financialStats.map((item, idx) => (
                <StatCard key={idx} item={item} />
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
