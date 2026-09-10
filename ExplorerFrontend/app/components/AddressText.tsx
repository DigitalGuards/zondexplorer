'use client';

import { displayAddress } from '../lib/preferences';
import { usePreferences } from './PreferencesProvider';

export default function AddressText({
  address,
  leading = 8,
  trailing = 6,
}: {
  address: string;
  leading?: number;
  trailing?: number;
}) {
  const { preferences } = usePreferences();
  return (
    <span title={address} data-explorer-address={address.toLowerCase()}>
      {displayAddress(address, preferences.addressDisplay, leading, trailing)}
    </span>
  );
}
