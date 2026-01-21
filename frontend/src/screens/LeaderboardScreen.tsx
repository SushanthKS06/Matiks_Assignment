import React, { useCallback, useState } from 'react';
import { View, Text, FlatList, StyleSheet, RefreshControl, TouchableOpacity, Alert, Platform, Animated } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { useLeaderboard } from '../hooks/useLeaderboard';
import { UserRow } from '../components/UserRow';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { UserWithRank } from '../types';
import { api } from '../services/api';

interface LeaderboardScreenProps {
    onNavigateToSearch: () => void;
}

export function LeaderboardScreen({ onNavigateToSearch }: LeaderboardScreenProps) {
    const { users, loading, refreshing, error, hasMore, totalUsers, refresh, loadMore } = useLeaderboard(50);
    const [seeding, setSeeding] = useState(false);

    const handleSeedUsers = useCallback(async () => {
        if (seeding) return;
        setSeeding(true);
        try {
            const result = await api.seedUsers(10000);
            if (Platform.OS === 'web') {
                window.alert(`Successfully seeded ${result.users_added.toLocaleString()} users!`);
            } else {
                Alert.alert('Success', `Seeded ${result.users_added.toLocaleString()} users!`);
            }
            refresh();
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to seed users';
            if (Platform.OS === 'web') {
                window.alert(message);
            } else {
                Alert.alert('Error', message);
            }
        } finally {
            setSeeding(false);
        }
    }, [seeding, refresh]);

    const renderItem = useCallback(({ item, index }: { item: UserWithRank; index: number }) => (
        <UserRow user={item} index={index} />
    ), []);

    const keyExtractor = useCallback((item: UserWithRank) => item.id, []);

    const renderFooter = useCallback(() => {
        if (!hasMore) {
            return (
                <View style={styles.footer}>
                    <View style={styles.footerLine} />
                    <Text style={styles.footerText}>End of leaderboard</Text>
                    <View style={styles.footerLine} />
                </View>
            );
        }
        return <LoadingSpinner message="Loading more..." />;
    }, [hasMore]);

    const renderHeader = useCallback(() => (
        <View style={styles.header}>
            <LinearGradient
                colors={['rgba(139, 92, 246, 0.15)', 'transparent']}
                style={styles.headerGradient}
            />
            <View style={styles.headerTop}>
                <View style={styles.logoSection}>
                    <LinearGradient
                        colors={['#8b5cf6', '#06b6d4']}
                        style={styles.logoIcon}
                    >
                        <Text style={styles.logoEmoji}>🏆</Text>
                    </LinearGradient>
                    <View>
                        <Text style={styles.title}>Leaderboard</Text>
                        <Text style={styles.subtitle}>Real-time rankings</Text>
                    </View>
                </View>
                <View style={styles.liveIndicator}>
                    <View style={styles.liveDot} />
                    <Text style={styles.liveText}>Live</Text>
                </View>
            </View>

            <View style={styles.statsBar}>
                <View style={styles.statItem}>
                    <Text style={styles.statValue}>{totalUsers.toLocaleString()}</Text>
                    <Text style={styles.statLabel}>Players</Text>
                </View>
                <View style={styles.statDivider} />
                <View style={styles.statItem}>
                    <Text style={styles.statValue}>O(1)</Text>
                    <Text style={styles.statLabel}>Rank Lookup</Text>
                </View>
                <View style={styles.statDivider} />
                <View style={styles.statItem}>
                    <Text style={styles.statValue}>~100</Text>
                    <Text style={styles.statLabel}>Updates/sec</Text>
                </View>
            </View>

            <View style={styles.actionButtons}>
                <TouchableOpacity
                    onPress={handleSeedUsers}
                    style={[styles.seedButton, seeding && styles.seedButtonDisabled]}
                    disabled={seeding}
                >
                    <LinearGradient
                        colors={seeding ? ['#374151', '#374151'] : ['#10b981', '#059669']}
                        style={styles.seedButtonGradient}
                    >
                        <Text style={styles.seedButtonText}>
                            {seeding ? '⏳ Seeding...' : '⚡ Seed 10K Users'}
                        </Text>
                    </LinearGradient>
                </TouchableOpacity>
                <TouchableOpacity onPress={onNavigateToSearch} style={styles.searchButton}>
                    <Text style={styles.searchButtonText}>🔍 Search</Text>
                </TouchableOpacity>
            </View>

            <View style={styles.tableHeader}>
                <Text style={[styles.tableHeaderText, styles.rankHeader]}>RANK</Text>
                <Text style={[styles.tableHeaderText, styles.nameHeader]}>PLAYER</Text>
                <Text style={[styles.tableHeaderText, styles.ratingHeader]}>RATING</Text>
            </View>
        </View>
    ), [totalUsers, onNavigateToSearch, handleSeedUsers, seeding]);

    if (loading && users.length === 0) {
        return (
            <View style={styles.loadingContainer}>
                <LinearGradient
                    colors={['#0a0a0f', '#12121a']}
                    style={StyleSheet.absoluteFill}
                />
                <LoadingSpinner fullScreen message="Loading leaderboard..." />
            </View>
        );
    }

    if (error && users.length === 0) {
        return (
            <View style={styles.errorContainer}>
                <LinearGradient
                    colors={['#0a0a0f', '#12121a']}
                    style={StyleSheet.absoluteFill}
                />
                <Text style={styles.errorEmoji}>❌</Text>
                <Text style={styles.errorTitle}>Failed to load leaderboard</Text>
                <Text style={styles.errorMessage}>{error}</Text>
                <TouchableOpacity style={styles.retryButton} onPress={refresh}>
                    <LinearGradient
                        colors={['#8b5cf6', '#06b6d4']}
                        style={styles.retryButtonGradient}
                    >
                        <Text style={styles.retryButtonText}>Retry Connection</Text>
                    </LinearGradient>
                </TouchableOpacity>
            </View>
        );
    }

    return (
        <View style={styles.container}>
            <LinearGradient
                colors={['#0a0a0f', '#12121a', '#0a0a0f']}
                style={StyleSheet.absoluteFill}
            />
            <FlatList
                data={users}
                renderItem={renderItem}
                keyExtractor={keyExtractor}
                ListHeaderComponent={renderHeader}
                ListFooterComponent={renderFooter}
                onEndReached={loadMore}
                onEndReachedThreshold={0.5}
                refreshControl={
                    <RefreshControl
                        refreshing={refreshing}
                        onRefresh={refresh}
                        tintColor="#8b5cf6"
                        colors={['#8b5cf6']}
                    />
                }
                initialNumToRender={20}
                maxToRenderPerBatch={20}
                windowSize={10}
                removeClippedSubviews={true}
                getItemLayout={(_, index) => ({
                    length: 80,
                    offset: 80 * index,
                    index,
                })}
                contentContainerStyle={styles.listContent}
            />
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#0a0a0f',
    },
    loadingContainer: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
    },
    listContent: {
        paddingBottom: 20,
    },
    header: {
        padding: 24,
        paddingTop: 60,
        borderBottomWidth: 1,
        borderBottomColor: 'rgba(255, 255, 255, 0.05)',
        position: 'relative',
        overflow: 'hidden',
    },
    headerGradient: {
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: 200,
    },
    headerTop: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
    },
    logoSection: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 14,
    },
    logoIcon: {
        width: 52,
        height: 52,
        borderRadius: 16,
        alignItems: 'center',
        justifyContent: 'center',
    },
    logoEmoji: {
        fontSize: 26,
    },
    title: {
        fontSize: 28,
        fontWeight: '800',
        color: '#ffffff',
        letterSpacing: -0.5,
    },
    subtitle: {
        fontSize: 13,
        color: '#71717a',
        marginTop: 2,
    },
    liveIndicator: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 6,
        paddingHorizontal: 12,
        paddingVertical: 6,
        backgroundColor: 'rgba(16, 185, 129, 0.1)',
        borderRadius: 20,
        borderWidth: 1,
        borderColor: 'rgba(16, 185, 129, 0.2)',
    },
    liveDot: {
        width: 8,
        height: 8,
        borderRadius: 4,
        backgroundColor: '#10b981',
    },
    liveText: {
        fontSize: 12,
        fontWeight: '600',
        color: '#10b981',
    },
    statsBar: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'center',
        marginTop: 24,
        paddingVertical: 16,
        backgroundColor: 'rgba(255, 255, 255, 0.03)',
        borderRadius: 16,
        borderWidth: 1,
        borderColor: 'rgba(255, 255, 255, 0.05)',
    },
    statItem: {
        flex: 1,
        alignItems: 'center',
    },
    statValue: {
        fontSize: 20,
        fontWeight: '700',
        color: '#ffffff',
    },
    statLabel: {
        fontSize: 11,
        color: '#71717a',
        marginTop: 2,
        textTransform: 'uppercase',
        letterSpacing: 0.5,
    },
    statDivider: {
        width: 1,
        height: 32,
        backgroundColor: 'rgba(255, 255, 255, 0.1)',
    },
    actionButtons: {
        flexDirection: 'row',
        gap: 12,
        marginTop: 20,
    },
    seedButton: {
        flex: 1,
        borderRadius: 14,
        overflow: 'hidden',
    },
    seedButtonDisabled: {
        opacity: 0.6,
    },
    seedButtonGradient: {
        paddingVertical: 14,
        alignItems: 'center',
    },
    seedButtonText: {
        color: '#ffffff',
        fontSize: 15,
        fontWeight: '600',
    },
    searchButton: {
        flex: 1,
        backgroundColor: 'rgba(255, 255, 255, 0.05)',
        paddingVertical: 14,
        borderRadius: 14,
        alignItems: 'center',
        borderWidth: 1,
        borderColor: 'rgba(255, 255, 255, 0.1)',
    },
    searchButtonText: {
        color: '#a1a1aa',
        fontSize: 15,
        fontWeight: '600',
    },
    tableHeader: {
        flexDirection: 'row',
        marginTop: 24,
        paddingBottom: 12,
        borderBottomWidth: 1,
        borderBottomColor: 'rgba(255, 255, 255, 0.05)',
    },
    tableHeaderText: {
        fontSize: 11,
        fontWeight: '600',
        color: '#71717a',
        letterSpacing: 1,
    },
    rankHeader: {
        width: 80,
        textAlign: 'center',
    },
    nameHeader: {
        flex: 1,
        marginLeft: 12,
    },
    ratingHeader: {
        width: 90,
        textAlign: 'right',
    },
    footer: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
        gap: 12,
    },
    footerLine: {
        flex: 1,
        height: 1,
        backgroundColor: 'rgba(255, 255, 255, 0.05)',
    },
    footerText: {
        color: '#71717a',
        fontSize: 13,
    },
    errorContainer: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 24,
    },
    errorEmoji: {
        fontSize: 56,
    },
    errorTitle: {
        fontSize: 20,
        fontWeight: '700',
        color: '#ffffff',
        marginTop: 20,
    },
    errorMessage: {
        fontSize: 14,
        color: '#71717a',
        marginTop: 8,
        textAlign: 'center',
    },
    retryButton: {
        marginTop: 24,
        borderRadius: 12,
        overflow: 'hidden',
    },
    retryButtonGradient: {
        paddingHorizontal: 32,
        paddingVertical: 14,
    },
    retryButtonText: {
        color: '#ffffff',
        fontSize: 16,
        fontWeight: '600',
    },
});
