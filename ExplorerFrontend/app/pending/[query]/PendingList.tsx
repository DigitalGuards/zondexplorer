"use client";

import React, { useEffect } from 'react';
import axios from 'axios';
import config from '../../../config';
import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { PendingTransaction, PendingTransactionsResponse } from '../tx/types';
import { formatAmount } from '../../lib/helpers';

interface PendingTransactionDisplay {
  hash: string;
  from: string;
  to: string | null;
  value: string;
  gasPrice: string;
  timestamp: number;
}

const transformPendingData = (data: PendingTransactionsResponse): PendingTransactionDisplay[] => {
  console.log('Raw response data:', data);
  const transformedTxs: PendingTransactionDisplay[] = [];
  
  try {
    if (data?.pending) {
      Object.entries(data.pending).forEach(([address, nonceTxs]) => {
        console.log(`Processing address ${address}:`, nonceTxs);
        if (typeof nonceTxs === 'object') {
          Object.entries(nonceTxs).forEach(([nonce, tx]) => {
            console.log(`Processing nonce ${nonce}:`, tx);
            if (tx && typeof tx === 'object' && 'hash' in tx) {
              try {
                const transaction: PendingTransactionDisplay = {
                  hash: tx.hash,
                  from: tx.from,
                  to: tx.to || null,
                  value: tx.value,
                  gasPrice: tx.gasPrice,
                  timestamp: Math.floor(Date.now() / 1000)
                };
                console.log('Transformed transaction:', transaction);
                transformedTxs.push(transaction);
              } catch (error) {
                console.error('Error transforming transaction:', error, tx);
              }
            }
          });
        }
      });
    }
  } catch (error) {
    console.error('Error in transformPendingData:', error);
  }
  
  console.log('All transformed transactions:', transformedTxs);
  return transformedTxs;
};

const fetchPendingTransactions = async () => {
  console.log('Fetching pending transactions...');
  const response = await axios.get(`${config.handlerUrl}/pending-transactions`);
  console.log('Raw API response:', response.data);
  
  const transformedTxs = transformPendingData(response.data);
  
  return {
    txs: transformedTxs,
    total: transformedTxs.length
  };
};

interface PendingListProps {
  initialData: {
    txs: PendingTransactionDisplay[];
    total: number;
  };
  currentPage: number;
}

function TransactionCard({ transaction }: { transaction: PendingTransactionDisplay }) {
  const date = new Date(transaction.timestamp * 1000).toLocaleString('en-GB', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });

  const [formattedValue, valueUnit] = formatAmount(transaction.value);
  const [formattedGasPrice, gasPriceUnit] = formatAmount(transaction.gasPrice);

  return (
    <Link 
      href={`/pending/tx/${transaction.hash}`}
      className='relative overflow-hidden rounded-2xl 
                bg-gradient-to-br from-[#2d2d2d] to-[#1f1f1f]
                border border-[#3d3d3d] shadow-xl
                hover:border-[#ffa729] transition-all duration-300
                group mb-4 block'
    >
      <div className="flex items-center p-6">
        <div className="flex flex-col items-center w-48">
          <div className="mb-2">
            <div className="bg-yellow-500/20 text-yellow-500 px-3 py-1 rounded-lg text-sm">
              Pending
            </div>
          </div>
          <div className="text-center">
            <p className="text-lg font-semibold text-[#ffa729] mb-1">Transfer</p>
            <p className="text-sm text-gray-400">{date}</p>
          </div>
        </div>

        <div className="flex-1 px-8 border-l border-r border-[#3d3d3d]">
          <div className="mb-2">
            <p className="text-sm font-medium text-gray-400 mb-1">Transaction Hash</p>
            <p className="text-gray-300 hover:text-[#ffa729] transition-colors break-all font-mono">
              {transaction.hash}
            </p>
          </div>
          <div className="mt-4">
            <p className="text-sm font-medium text-gray-400 mb-1">From</p>
            <p className="text-gray-300 font-mono truncate">
              {transaction.from}
            </p>
            {transaction.to && (
              <div className="mt-2">
                <p className="text-sm font-medium text-gray-400 mb-1">To</p>
                <p className="text-gray-300 font-mono truncate">
                  {transaction.to}
                </p>
              </div>
            )}
          </div>
        </div>

        <div className="w-48 text-right">
          <p className="text-sm font-medium text-gray-400 mb-2">Amount</p>
          <p className="text-2xl font-semibold text-[#ffa729]">
            {formattedValue}
            <span className="text-sm text-gray-400 ml-2">{valueUnit}</span>
          </p>
          <p className="text-sm text-gray-400 mt-2">
            Gas Price: {formattedGasPrice} {gasPriceUnit}
          </p>
        </div>
      </div>
    </Link>
  );
}

export default function PendingList({ initialData, currentPage }: PendingListProps) {
  const { data, isLoading, isFetching, isError, error, refetch } = useQuery({
    queryKey: ['pending-transactions'],
    queryFn: fetchPendingTransactions,
    refetchInterval: 5000,
    retry: 2,
    refetchOnMount: true
  });

  useEffect(() => {
    console.log('Component mounted, forcing refetch...');
    refetch();
  }, [refetch]);

  useEffect(() => {
    console.log('Current data:', data);
  }, [data]);

  if (isLoading || isFetching) {
    return (
      <div className="container mx-auto px-4">
        <h1 className="text-2xl font-bold mb-6 text-[#ffa729]">Mempool</h1>
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div 
              key={i}
              className="rounded-xl bg-gradient-to-br from-[#2d2d2d] to-[#1f1f1f] border border-[#3d3d3d] p-6 animate-pulse"
            >
              <div className="flex flex-col md:flex-row items-center">
                <div className="w-48 flex flex-col items-center">
                  <div className="w-20 h-6 bg-gray-700 rounded-lg mb-2"></div>
                  <div className="h-4 w-24 bg-gray-700 rounded"></div>
                </div>
                <div className="flex-1 md:mx-8 space-y-2 border-l border-r border-[#3d3d3d] px-8">
                  <div className="h-4 w-32 bg-gray-700 rounded"></div>
                  <div className="h-6 w-full bg-gray-700 rounded"></div>
                  <div className="h-4 w-32 bg-gray-700 rounded mt-4"></div>
                  <div className="h-6 w-2/3 bg-gray-700 rounded"></div>
                </div>
                <div className="w-48 space-y-2">
                  <div className="h-4 w-20 bg-gray-700 rounded ml-auto"></div>
                  <div className="h-8 w-32 bg-gray-700 rounded ml-auto"></div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="container mx-auto px-4">
        <h1 className="text-2xl font-bold mb-6 text-[#ffa729]">Mempool</h1>
        <div className="bg-red-900/50 border border-red-500 text-red-200 px-6 py-4 rounded-xl">
          <p className="font-bold">Error:</p>
          <p>{error instanceof Error ? error.message : 'Failed to load pending transactions'}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4">
      <h1 className="text-2xl font-bold mb-6 text-[#ffa729]">Mempool</h1>
      {!data || data.txs.length === 0 ? (
        <div className="bg-gradient-to-r from-[#2d2d2d] to-[#1f1f1f] border border-[#3d3d3d] rounded-xl p-8 text-center">
          <div className="inline-block p-4 bg-yellow-500/20 rounded-full mb-4">
            <svg className="w-8 h-8 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
          </div>
          <h2 className="text-xl font-semibold text-[#ffa729] mb-2">No Pending Transactions</h2>
          <p className="text-gray-400 max-w-lg mx-auto">
            Showing unconfirmed transactions waiting to be included in a block. Updates every 5 seconds.
          </p>
        </div>
      ) : (
        <div className="mb-8">
          <div className="bg-gradient-to-r from-[#2d2d2d] to-[#1f1f1f] border border-[#3d3d3d] rounded-xl p-4 mb-6">
            <p className="text-gray-300">
              <span className="text-[#ffa729] font-semibold">{data.txs.length}</span> unconfirmed {data.txs.length === 1 ? 'transaction' : 'transactions'} waiting to be included in a block. Updates every 5 seconds.
            </p>
          </div>
          {data.txs.map(transaction => (
            <TransactionCard 
              key={transaction.hash} 
              transaction={transaction}
            />
          ))}
        </div>
      )}
    </div>
  );
}
