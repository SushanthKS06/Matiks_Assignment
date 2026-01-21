import { useState, useCallback, useMemo } from 'react';
import { api } from '../services/api';
import { UserWithRank } from '../types';
import { debounce } from '../utils/debounce';

export function useSearch() {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState<UserWithRank[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [searched, setSearched] = useState(false);

    const performSearch = useCallback(async (searchQuery: string) => {
        if (!searchQuery.trim()) {
            setResults([]);
            setSearched(false);
            return;
        }

        try {
            setLoading(true);
            setError(null);
            setSearched(true);

            const response = await api.searchUsers(searchQuery.trim());
            setResults(response.users);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to search');
            setResults([]);
        } finally {
            setLoading(false);
        }
    }, []);

    const debouncedSearch = useMemo(
        () => debounce(performSearch, 300),
        [performSearch]
    );

    const handleQueryChange = useCallback((text: string) => {
        setQuery(text);
        debouncedSearch(text);
    }, [debouncedSearch]);

    const clearSearch = useCallback(() => {
        setQuery('');
        setResults([]);
        setSearched(false);
        setError(null);
    }, []);

    return {
        query,
        results,
        loading,
        error,
        searched,
        handleQueryChange,
        clearSearch,
    };
}
