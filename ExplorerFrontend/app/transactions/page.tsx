import React from 'react';
import TransactionsList from './[query]/TransactionsList';
import { Transaction } from './[query]/types';

export const dynamic = 'force-dynamic';

interface TransactionsResponse {
  txs: Transaction[];
  total: number;
}

async function getTransactions(page: string): Promise<TransactionsResponse> {
  const handlerUrl = process.env.NEXT_PUBLIC_HANDLER_URL || 'http://localhost:8080';
  const response = await fetch(`${handlerUrl}/txs?page=${page}`, {
    next: { revalidate: 10 }
  });
  
  if (!response.ok) {
    throw new Error('Failed to fetch transactions');
  }

  return response.json();
}

export default async function Page({
  searchParams,
}: {
  searchParams: { page?: string };
}) {
  // Get page from searchParams or default to '1'
  const page = searchParams.page || '1';
  
  // Fetch data
  const data = await getTransactions(page);
  
  // Parse page number
  const currentPage = parseInt(page);

  return (
    <TransactionsList 
      initialData={data} 
      currentPage={currentPage} 
    />
  );
}