'use client';

import { useState, useCallback, useEffect, ChangeEvent, FormEvent } from 'react';
import { useRouter } from 'next/navigation';

function onlyNumbers(str: string): boolean {
  return /^[0-9]+$/.test(str);
}

// Validates if the input is a valid address (either 0x or Z prefixed)
function isValidAddress(address: string): boolean {
  // Check for Z-prefixed address (Z + 40 hex chars)
  if (address.startsWith('Z') && address.length === 41) {
    // Check if the rest of the string is valid hex
    return /^Z[0-9a-fA-F]{40}$/.test(address);
  }
  
  // Check for 0x-prefixed address (0x + 40 hex chars)
  if (address.startsWith('0x') && address.length === 42) {
    return /^0x[0-9a-fA-F]{40}$/.test(address);
  }
  
  return false;
}

export default function SearchBar(): JSX.Element {
  const [searchValue, setSearchValue] = useState<string>('');
  const [error, setError] = useState<string>('');
  const router = useRouter();

  function handleInputChange(event: ChangeEvent<HTMLInputElement>): void {
    setSearchValue(event.target.value);
    setError('');
  }

  const navigateHandler = useCallback((): void => {
    let newPath: string;
    if (onlyNumbers(searchValue)) {
      newPath = "/block/" + searchValue;
    } else if (searchValue.length === 66) {
      newPath = "/tx/" + searchValue;
    } else if (isValidAddress(searchValue)) {
      newPath = "/address/" + searchValue;
    } else {
      setError('Invalid input!');
      return;
    }
    router.push(newPath);
  }, [searchValue, router]);

  useEffect(() => {
    const listener = (event: KeyboardEvent): void => {
      if (event.code === "Enter" || event.code === "NumpadEnter") {
        event.preventDefault();
        navigateHandler();
      }
    };
    window.addEventListener("keydown", listener);
    return () => {
      window.removeEventListener("keydown", listener);
    };
  }, [navigateHandler]);

  return (
    <div className="relative w-full">
      <div className="relative bg-gradient-to-br from-[#2d2d2d] to-[#1f1f1f] rounded-2xl p-3 sm:p-6 
                    shadow-xl border border-[#3d3d3d] hover:border-[#4d4d4d] transition-colors">
        <form
          onSubmit={(e: FormEvent<HTMLFormElement>) => {
            e.preventDefault();
            navigateHandler();
          }}
          className="flex flex-col sm:flex-row gap-3 sm:gap-6">
          <input
            type="text"
            placeholder="Search by Address (Zxx) / Txn Hash / Block.."
            className="flex-1 py-3 sm:py-4 px-4 sm:px-6 text-sm sm:text-base text-gray-300 
                     bg-[#1a1a1a] rounded-xl
                     border border-[#3d3d3d]
                     outline-none shadow-lg
                     focus:ring-2 focus:ring-[#ffa729] focus:border-transparent
                     placeholder-gray-500 transition-all duration-300
                     hover:border-[#4d4d4d]"
            value={searchValue}
            onChange={handleInputChange}
          />
          <button
            type="submit"
            className="px-8 sm:px-10 py-3 sm:py-4 bg-[#ffa729] text-white text-sm sm:text-base
                     rounded-xl shadow-lg font-medium whitespace-nowrap
                     hover:bg-[#ff9709] hover:shadow-2xl hover:scale-105 
                     active:scale-95 transition-all duration-300 
                     sm:w-auto w-full"
          >
            Search
          </button>
        </form>
        {error && (
          <div className="mt-3 sm:mt-4">
            <div className="p-3 sm:p-4 mb-3 sm:mb-4 text-xs sm:text-sm text-red-400 rounded-xl bg-gradient-to-br from-[#2d2d2d] to-[#1f1f1f] border border-red-400 shadow-lg" role="alert">
              <span className="font-medium">{error}</span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
