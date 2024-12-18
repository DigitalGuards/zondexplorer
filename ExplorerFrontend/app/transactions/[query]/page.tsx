import { Suspense } from 'react';
import { notFound } from 'next/navigation';
import TransactionsClient from './transactions-client';
import type { TransactionsResponse } from './types';

async function getTransactions(page: string): Promise<TransactionsResponse> {
  const handlerUrl = process.env.NEXT_PUBLIC_HANDLER_URL || 'http://localhost:8080';
  
  try {
    const response = await fetch(`${handlerUrl}/txs?page=${page}`, {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      },
      next: {
        revalidate: 0
      }
    });

    if (!response.ok) {
      if (response.status === 404) {
        notFound();
      }
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  } catch (error) {
    console.error('Error fetching transactions:', error);
    throw error;
  }
}

function LoadingUI(): JSX.Element {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-lg">Loading transactions...</div>
    </div>
  );
}

interface PageProps {
    params: Promise<{ query: string }>;
}

export default async function Page({ params }: PageProps): Promise<JSX.Element> {
  const resolvedParams = await params;
  const pageNumber = resolvedParams.query || '1';

  try {
    const data = await getTransactions(pageNumber);

    return (
      <main>
        <h1 className="sr-only">Transactions - Page {pageNumber}</h1>
        <Suspense fallback={<LoadingUI />}>
          <TransactionsClient 
            initialData={data} 
            pageNumber={pageNumber} 
          />
        </Suspense>
      </main>
    );
  } catch (error) {
    return (
      <div role="alert" className="p-4">
        <h1 className="text-xl font-bold mb-2">Error</h1>
        <p>Failed to load transactions. Please try again later.</p>
        {process.env.NODE_ENV === 'development' && (
          <pre className="mt-2 text-sm text-red-500">
            {error instanceof Error ? error.message : 'Unknown error'}
          </pre>
        )}
      </div>
    );
  }
}
