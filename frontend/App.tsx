import React, { useState, useEffect, useCallback } from 'react';
import {
  StatusBar,
  View,
  Text,
  FlatList,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  RefreshControl,
  SafeAreaView
} from 'react-native';

interface UserWithRank {
  id: string;
  username: string;
  rating: number;
  rank: number;
}

interface LeaderboardResponse {
  users: UserWithRank[];
  total_users: number;
  has_more: boolean;
}

interface SearchResponse {
  users: UserWithRank[];
  count: number;
}

// HARDCODED API URL for debugging - App.tsx ignores src/services/api.ts
const API_URL = 'https://siderographic-shay-frivolously.ngrok-free.dev/api';
// const API_URL = process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8080/api';

type Screen = 'leaderboard' | 'search';

export default function App() {
  const [screen, setScreen] = useState<Screen>('leaderboard');
  const [users, setUsers] = useState<UserWithRank[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [totalUsers, setTotalUsers] = useState(0);
  const [offset, setOffset] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<UserWithRank[]>([]);
  const [searching, setSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [seeding, setSeeding] = useState(false);

  const fetchLeaderboard = useCallback(async (newOffset: number = 0, append: boolean = false) => {
    try {
      if (!append) setLoading(true);
      setError(null);

      const response = await fetch(`https://siderographic-shay-frivolously.ngrok-free.dev/api/leaderboard?limit=50&offset=${newOffset}`, {
        headers: {
          'ngrok-skip-browser-warning': 'true',
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        }
      });
      if (!response.ok) throw new Error('Failed to fetch');

      const data: LeaderboardResponse = await response.json();

      if (append) {
        setUsers(prev => {
          const existingIds = new Set(prev.map(u => u.id));
          const newUsers = data.users.filter(u => !existingIds.has(u.id));
          return [...prev, ...newUsers];
        });
      } else {
        setUsers(data.users);
      }

      setHasMore(data.has_more);
      setTotalUsers(data.total_users);
      setOffset(newOffset + data.users.length);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error loading data');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  const handleSearch = useCallback(async (query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }

    try {
      setSearching(true);
      const response = await fetch(`${API_URL}/search?q=${encodeURIComponent(query)}`, {
        headers: {
          'ngrok-skip-browser-warning': 'true',
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        }
      });
      if (!response.ok) throw new Error('Search failed');

      const data: SearchResponse = await response.json();
      setSearchResults(data.users);
    } catch (err) {
      console.error('Search error:', err);
    } finally {
      setSearching(false);
    }
  }, []);

  const handleSeedUsers = useCallback(async () => {
    if (seeding) return;
    setSeeding(true);
    try {
      const response = await fetch(`${API_URL}/seed?count=10000`, {
        method: 'POST',
        headers: {
          'ngrok-skip-browser-warning': 'true',
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        }
      });
      if (!response.ok) throw new Error('Failed to seed users');
      const data = await response.json();
      alert(`Successfully seeded ${data.users_added.toLocaleString()} users!`);
      setOffset(0);
      fetchLeaderboard(0, false);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to seed users');
    } finally {
      setSeeding(false);
    }
  }, [seeding, fetchLeaderboard]);

  useEffect(() => {
    fetchLeaderboard(0);
  }, [fetchLeaderboard]);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (screen === 'search') {
        handleSearch(searchQuery);
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery, screen, handleSearch]);

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    setOffset(0);
    fetchLeaderboard(0, false);
  }, [fetchLeaderboard]);

  const loadMore = useCallback(() => {
    if (hasMore && !loading) {
      fetchLeaderboard(offset, true);
    }
  }, [hasMore, loading, offset, fetchLeaderboard]);

  const renderUser = useCallback(({ item, index }: { item: UserWithRank; index: number }) => {
    const isEven = index % 2 === 0;
    const badge = item.rank === 1 ? '🥇' : item.rank === 2 ? '🥈' : item.rank === 3 ? '🥉' : null;

    return (
      <View style={[styles.userRow, isEven && styles.evenRow]}>
        <View style={styles.rankContainer}>
          {badge ? (
            <Text style={styles.badge}>{badge}</Text>
          ) : (
            <Text style={styles.rank}>#{item.rank}</Text>
          )}
        </View>
        <View style={styles.userInfo}>
          <Text style={styles.username}>{item.username}</Text>
          <Text style={styles.userId}>{item.id.slice(0, 8)}...</Text>
        </View>
        <View style={styles.ratingContainer}>
          <Text style={styles.rating}>{item.rating}</Text>
          <Text style={styles.ratingLabel}>RATING</Text>
        </View>
      </View>
    );
  }, []);

  if (loading && users.length === 0) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#10B981" />
        <Text style={styles.loadingText}>Loading leaderboard...</Text>
      </View>
    );
  }

  if (error && users.length === 0) {
    return (
      <View style={styles.errorContainer}>
        <Text style={styles.errorEmoji}>❌</Text>
        <Text style={styles.errorTitle}>Failed to load</Text>
        <Text style={styles.errorMessage}>{error}</Text>
        <TouchableOpacity style={styles.retryButton} onPress={() => fetchLeaderboard(0)}>
          <Text style={styles.retryText}>RETRY V5 - FINAL SURGICAL FIX</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (screen === 'search') {
    return (
      <SafeAreaView style={styles.container}>
        <StatusBar barStyle="light-content" backgroundColor="#1F2937" />
        <View style={styles.searchHeader}>
          <TouchableOpacity onPress={() => setScreen('leaderboard')} style={styles.backButton}>
            <Text style={styles.backText}>← Back</Text>
          </TouchableOpacity>
          <Text style={styles.searchTitle}>Search Players</Text>
          <View style={{ width: 60 }} />
        </View>

        <View style={styles.searchBar}>
          <Text style={styles.searchIcon}>🔍</Text>
          <TextInput
            style={styles.searchInput}
            value={searchQuery}
            onChangeText={setSearchQuery}
            placeholder="Search by username (e.g., rahul)"
            placeholderTextColor="#6B7280"
            autoCapitalize="none"
            autoCorrect={false}
          />
          {searchQuery.length > 0 && (
            <TouchableOpacity onPress={() => setSearchQuery('')}>
              <Text style={styles.clearText}>✕</Text>
            </TouchableOpacity>
          )}
        </View>

        {searching ? (
          <View style={styles.searchingContainer}>
            <ActivityIndicator size="small" color="#10B981" />
            <Text style={styles.searchingText}>Searching...</Text>
          </View>
        ) : searchResults.length > 0 ? (
          <>
            <Text style={styles.resultsCount}>Found {searchResults.length} players</Text>
            <FlatList
              data={searchResults}
              renderItem={renderUser}
              keyExtractor={(item, index) => `search-${item.id}-${index}`}
            />
          </>
        ) : searchQuery.length > 0 ? (
          <View style={styles.noResults}>
            <Text style={styles.noResultsEmoji}>🔍</Text>
            <Text style={styles.noResultsTitle}>No results found</Text>
            <Text style={styles.noResultsMessage}>No users matching "{searchQuery}"</Text>
          </View>
        ) : (
          <View style={styles.noResults}>
            <Text style={styles.noResultsEmoji}>👤</Text>
            <Text style={styles.noResultsTitle}>Search for a player</Text>
            <Text style={styles.noResultsMessage}>Enter a username to find their global rank</Text>
          </View>
        )}
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <StatusBar barStyle="light-content" backgroundColor="#1F2937" />
      <View style={styles.header}>
        <View style={styles.headerRow}>
          <Text style={styles.title}>🏆 Leaderboard</Text>
          <View style={styles.headerButtons}>
            <TouchableOpacity
              onPress={handleSeedUsers}
              style={[styles.seedButton, seeding && styles.seedButtonDisabled]}
              disabled={seeding}
            >
              <Text style={styles.seedButtonText}>
                {seeding ? '⏳...' : '🌱 Seed 10k'}
              </Text>
            </TouchableOpacity>
            <TouchableOpacity onPress={() => setScreen('search')} style={styles.searchButton}>
              <Text style={styles.searchButtonText}>🔍</Text>
            </TouchableOpacity>
          </View>
        </View>
        <Text style={styles.subtitle}>{totalUsers.toLocaleString()} players competing</Text>
        <View style={styles.tableHeader}>
          <Text style={[styles.headerCell, { width: 70, textAlign: 'center' }]}>RANK</Text>
          <Text style={[styles.headerCell, { flex: 1, marginLeft: 12 }]}>PLAYER</Text>
          <Text style={[styles.headerCell, { width: 80, textAlign: 'center' }]}>RATING</Text>
        </View>
      </View>

      <FlatList
        data={users}
        renderItem={renderUser}
        keyExtractor={(item, index) => `${item.id}-${index}`}
        onEndReached={loadMore}
        onEndReachedThreshold={0.5}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor="#10B981"
            colors={['#10B981']}
          />
        }
        ListFooterComponent={() => (
          hasMore ? (
            <View style={styles.footer}>
              <ActivityIndicator size="small" color="#10B981" />
              <Text style={styles.footerText}>Loading more...</Text>
            </View>
          ) : (
            <View style={styles.footer}>
              <Text style={styles.footerText}>End of leaderboard</Text>
            </View>
          )
        )}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#111827',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#111827',
  },
  loadingText: {
    marginTop: 16,
    fontSize: 16,
    color: '#9CA3AF',
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#111827',
    padding: 20,
  },
  errorEmoji: {
    fontSize: 48,
  },
  errorTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#F9FAFB',
    marginTop: 16,
  },
  errorMessage: {
    fontSize: 14,
    color: '#9CA3AF',
    marginTop: 8,
  },
  retryButton: {
    marginTop: 20,
    backgroundColor: '#10B981',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  retryText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  header: {
    padding: 20,
    paddingTop: 48,
    backgroundColor: '#1F2937',
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  headerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: 8,
  },
  title: {
    fontSize: 22,
    fontWeight: '800',
    color: '#F9FAFB',
    flexShrink: 1,
  },
  headerButtons: {
    flexDirection: 'row',
    gap: 6,
    flexShrink: 0,
  },
  seedButton: {
    backgroundColor: '#059669',
    paddingHorizontal: 10,
    paddingVertical: 8,
    borderRadius: 8,
  },
  seedButtonDisabled: {
    backgroundColor: '#6B7280',
  },
  seedButtonText: {
    color: '#F9FAFB',
    fontSize: 12,
    fontWeight: '600',
  },
  searchButton: {
    backgroundColor: '#374151',
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 8,
  },
  searchButtonText: {
    color: '#F9FAFB',
    fontSize: 16,
    fontWeight: '600',
  },
  subtitle: {
    fontSize: 14,
    color: '#9CA3AF',
    marginTop: 8,
  },
  tableHeader: {
    flexDirection: 'row',
    marginTop: 20,
    paddingBottom: 10,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  headerCell: {
    fontSize: 12,
    fontWeight: '600',
    color: '#6B7280',
    letterSpacing: 1,
  },
  userRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    backgroundColor: '#1F2937',
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  evenRow: {
    backgroundColor: '#111827',
  },
  rankContainer: {
    width: 60,
    alignItems: 'center',
  },
  rank: {
    fontSize: 16,
    fontWeight: '700',
    color: '#6B7280',
  },
  badge: {
    fontSize: 24,
  },
  userInfo: {
    flex: 1,
    marginLeft: 12,
  },
  username: {
    fontSize: 16,
    fontWeight: '600',
    color: '#F9FAFB',
  },
  userId: {
    fontSize: 12,
    color: '#9CA3AF',
    marginTop: 2,
  },
  ratingContainer: {
    alignItems: 'center',
    backgroundColor: '#374151',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
  },
  rating: {
    fontSize: 18,
    fontWeight: '700',
    color: '#10B981',
  },
  ratingLabel: {
    fontSize: 10,
    color: '#9CA3AF',
    marginTop: 2,
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  footerText: {
    color: '#6B7280',
    fontSize: 14,
    marginLeft: 8,
  },
  searchHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingTop: 48,
    paddingBottom: 8,
    backgroundColor: '#1F2937',
  },
  backButton: {
    paddingVertical: 8,
    paddingRight: 16,
  },
  backText: {
    color: '#10B981',
    fontSize: 16,
    fontWeight: '600',
  },
  searchTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#F9FAFB',
  },
  searchBar: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1F2937',
    borderRadius: 12,
    paddingHorizontal: 16,
    marginHorizontal: 16,
    marginVertical: 12,
    borderWidth: 1,
    borderColor: '#374151',
  },
  searchIcon: {
    fontSize: 18,
    marginRight: 12,
  },
  searchInput: {
    flex: 1,
    paddingVertical: 14,
    fontSize: 16,
    color: '#F9FAFB',
  },
  clearText: {
    fontSize: 16,
    color: '#6B7280',
    padding: 8,
  },
  searchingContainer: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  searchingText: {
    marginLeft: 12,
    fontSize: 14,
    color: '#9CA3AF',
  },
  resultsCount: {
    color: '#9CA3AF',
    fontSize: 14,
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  noResults: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 40,
  },
  noResultsEmoji: {
    fontSize: 48,
  },
  noResultsTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#F9FAFB',
    marginTop: 16,
  },
  noResultsMessage: {
    fontSize: 14,
    color: '#9CA3AF',
    marginTop: 8,
    textAlign: 'center',
  },
});
