import { useState, useEffect, useCallback, useRef } from 'react';
import { api } from '../services/api';
import { UserWithRank, LeaderboardResponse } from '../types';

export function useLeaderboard(pageSize: number = 50) {
    const [users, setUsers] = useState<UserWithRank[]>([]);
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [hasMore, setHasMore] = useState(true);
    const [totalUsers, setTotalUsers] = useState(0);
    const offsetRef = useRef(0);
    const loadingMoreRef = useRef(false);

    const fetchLeaderboard = useCallback(async (offset: number = 0, append: boolean = false) => {
        try {
            if (!append) {
                setLoading(true);
            }
            setError(null);

            const response: LeaderboardResponse = await api.getLeaderboard(pageSize, offset);

            if (append) {
                setUsers(prev => [...prev, ...response.users]);
            } else {
                setUsers(response.users);
            }

            setHasMore(response.has_more);
            setTotalUsers(response.total_users);
            offsetRef.current = offset + response.users.length;
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch leaderboard');
        } finally {
            setLoading(false);
            setRefreshing(false);
            loadingMoreRef.current = false;
        }
    }, [pageSize]);

    const refresh = useCallback(() => {
        setRefreshing(true);
        offsetRef.current = 0;
        fetchLeaderboard(0, false);
    }, [fetchLeaderboard]);

    const loadMore = useCallback(() => {
        if (!hasMore || loadingMoreRef.current || loading) return;
        loadingMoreRef.current = true;
        fetchLeaderboard(offsetRef.current, true);
    }, [hasMore, loading, fetchLeaderboard]);

    useEffect(() => {
        fetchLeaderboard(0, false);
    }, [fetchLeaderboard]);

    return {
        users,
        loading,
        refreshing,
        error,
        hasMore,
        totalUsers,
        refresh,
        loadMore,
    };
}
