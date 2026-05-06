import { useEffect, useState } from 'react';

/**
 * Returns `value` debounced by `delay` ms. Useful for search inputs where you
 * want the latest character to settle before firing a query.
 *
 *   const q = useDebouncedValue(searchQuery, 300);
 *   useLibraryQuery({ q });
 */
export function useDebouncedValue<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const id = window.setTimeout(() => setDebounced(value), delay);
    return () => window.clearTimeout(id);
  }, [value, delay]);

  return debounced;
}
