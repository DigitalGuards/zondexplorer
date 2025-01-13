import { notFound } from 'next/navigation';
import TransactionView from './transaction-view';
import type { TransactionDetails } from './types';
import { decodeBase64ToHexadecimal } from '../../lib/server-helpers';
import config from '../../../config';

interface PageProps {
  params: Promise<{ query: string }>;
}

function isEmptyTransaction(txData: any): boolean {
  return !txData.hash && 
         !txData.from && 
         !txData.to && 
         (!txData.value || txData.value === 0) &&
         (!txData.blockNumber || txData.blockNumber === 0);
}

async function getTransaction(txHash: string): Promise<TransactionDetails> {
  // Validate transaction hash format
  const hashRegex = /^0x[0-9a-fA-F]{64}$/;
  if (!hashRegex.test(txHash)) {
    throw new Error('Invalid transaction hash format');
  }

  console.log('Fetching transaction:', txHash);
  const response = await fetch(`${config.handlerUrl}/tx/${txHash}`, {
    method: 'GET',
    headers: { Accept: 'application/json' },
    next: { revalidate: 60 }, // Cache for 60 seconds
  });

  if (!response.ok) {
    if (response.status === 404) {
      notFound();
    }
    throw new Error('Failed to fetch transaction details');
  }

  const data = await response.json();
  console.log('Raw API response:', JSON.stringify(data, null, 2));

  // Check if we have a valid transaction response
  if (!data.response || isEmptyTransaction(data.response)) {
    console.log('Invalid or missing transaction data');
    throw new Error('Transaction not found');
  }

  const txData = data.response;
  console.log('Transaction data:', JSON.stringify(txData, null, 2));

  // Helper function to handle address decoding
  const decodeAddress = (address: string | null | undefined): string => {
    console.log('Decoding address:', address);
    if (!address) return '';
    // Check if it's already a hex address (starts with 0x)
    if (address.startsWith('0x')) {
      console.log('Address is already hex:', address);
      return address;
    }
    // Otherwise decode from base64
    try {
      const decoded = decodeBase64ToHexadecimal(address);
      console.log('Decoded address:', decoded);
      return decoded; // decodeBase64ToHexadecimal already adds '0x' prefix
    } catch (error) {
      console.error('Error decoding address:', error);
      return address;
    }
  };

  // Helper function to convert number to hex string with proper scaling
  const formatBigValue = (value: string | number | null | undefined): string => {
    if (value === null || value === undefined) return '0x0';
    
    try {
      // Convert to string and then BigInt
      const numStr = value.toString();
      const num = BigInt(numStr);
      
      // Return hex string
      return '0x' + num.toString(16);
    } catch (error) {
      console.error('Error converting value:', error);
      return '0x0';
    }
  };

  // Helper function to ensure integer values
  const ensureInteger = (value: any): number => {
    if (typeof value === 'number') {
      return Math.floor(value);
    }
    if (typeof value === 'string') {
      return Math.floor(parseFloat(value));
    }
    return 0;
  };

  const transaction = {
    hash: txData.hash || txHash,
    blockNumber: ensureInteger(txData.blockNumber) || '',
    from: decodeAddress(txData.from),
    to: decodeAddress(txData.to),
    value: formatBigValue(txData.value),
    timestamp: ensureInteger(txData.blockTimestamp || txData.timestamp) || 0,
    gasUsed: formatBigValue(txData.gasUsed),
    gasPrice: formatBigValue(txData.gasPrice),
    nonce: ensureInteger(txData.nonce),
    latestBlock: ensureInteger(data.latestBlock)
  };

  console.log('Processed transaction:', JSON.stringify(transaction, null, 2));

  return transaction;
}

export default async function TransactionPage({ params }: PageProps): Promise<JSX.Element> {
  let resolvedParams;
  let txHash = '';
  
  try {
    resolvedParams = await params;
    txHash = resolvedParams.query;
    console.log('Transaction hash from params:', txHash);

    // Validate transaction hash format
    const hashRegex = /^0x[0-9a-fA-F]{64}$/;
    if (!hashRegex.test(txHash)) {
      return (
        <div className="container mx-auto px-4">
          <div className="bg-red-900/20 border border-red-500/50 rounded-xl p-6 shadow-lg mt-6">
            <h2 className="text-red-500 font-semibold mb-2">Invalid Transaction Hash</h2>
            <p className="text-gray-300">
              The provided transaction hash is not in the correct format. 
              Transaction hashes should start with &apos;0x&apos; followed by 64 hexadecimal characters.
            </p>
          </div>
        </div>
      );
    }

    // Check if transaction is in mempool
    const pendingResponse = await fetch(`${config.handlerUrl}/pending-transaction/${txHash}`);
    if (pendingResponse.ok && pendingResponse.status === 200) {
      const pendingData = await pendingResponse.json();
      if (pendingData?.transaction) {
        // If found in mempool, show pending message with link
        return (
          <div className="container mx-auto px-4">
            <div className="bg-yellow-900/20 border border-yellow-500/50 rounded-xl p-6 shadow-lg mt-6">
              <h2 className="text-yellow-500 font-semibold mb-2">Transaction Pending</h2>
              <p className="text-gray-300 mb-4">
                This transaction is still pending and has not been mined yet.
              </p>
              <a 
                href={`/pending/tx/${txHash}`}
                className="inline-block bg-yellow-500/20 text-yellow-500 px-4 py-2 rounded-lg hover:bg-yellow-500/30 transition-colors"
              >
                View Pending Transaction →
              </a>
            </div>
          </div>
        );
      }
    }

    // Not in mempool, try to get mined transaction
    const transaction = await getTransaction(txHash);
    return <TransactionView transaction={transaction} />;
  } catch (error) {
    console.error('Error in TransactionPage:', error);
    return (
      <div className="container mx-auto px-4">
        <div className="bg-red-900/20 border border-red-500/50 rounded-xl p-6 shadow-lg mt-6">
          <h2 className="text-red-500 font-semibold mb-2">Transaction Not Found</h2>
          <p className="text-gray-300">
            The transaction could not be found. This could mean:
          </p>
          <ul className="list-disc ml-6 mt-2 text-gray-300">
            <li>The transaction hash is incorrect</li>
            <li>The transaction has not been mined yet</li>
            <li>The transaction was dropped from the network</li>
          </ul>
        </div>
      </div>
    );
  }
}
