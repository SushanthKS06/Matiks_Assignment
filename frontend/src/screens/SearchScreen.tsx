import React, { useCallback } from 'react';
import { View, Text, FlatList, StyleSheet, TouchableOpacity, TextInput } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { useSearch } from '../hooks/useSearch';
import { UserRow } from '../components/UserRow';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { UserWithRank } from '../types';

interface SearchScreenProps {
    onNavigateBack: () => void;
}

export function SearchScreen({ onNavigateBack }: SearchScreenProps) {
    const { query, results, loading, error, searched, handleQueryChange, clearSearch } = useSearch();

    const renderItem = useCallback(({ item, index }: { item: UserWithRank; index: number }) => (
        <UserRow user={item} index={index} />
    ), []);

    const keyExtractor = useCallback((item: UserWithRank) => item.id, []);

    const renderEmpty = useCallback(() => {
        if (loading) {
            return <LoadingSpinner message="Searching..." />;
        }

        if (error) {
            return (
                <View style={styles.emptyContainer}>
                    <Text style={styles.emptyEmoji}>❌</Text>
                    <Text style={styles.emptyTitle}>Search failed</Text>
                    <Text style={styles.emptyMessage}>{error}</Text>
                </View>
            );
        }

        if (searched && results.length === 0) {
            return (
                <View style={styles.emptyContainer}>
                    <Text style={styles.emptyEmoji}>🔍</Text>
                    <Text style={styles.emptyTitle}>No results found</Text>
                    <Text style={styles.emptyMessage}>
                        No players matching "{query}"
                    </Text>
                </View>
            );
        }

        return (
            <View style={styles.emptyContainer}>
                <View style={styles.searchIconContainer}>
                    <LinearGradient
                        colors={['#8b5cf6', '#06b6d4']}
                        style={styles.searchIconGradient}
                    >
                        <Text style={styles.searchIconEmoji}>🔍</Text>
                    </LinearGradient>
                </View>
                <Text style={styles.emptyTitle}>Search for a player</Text>
                <Text style={styles.emptyMessage}>
                    Enter a username to find their global rank
                </Text>
            </View>
        );
    }, [loading, error, searched, results.length, query]);

    const renderHeader = useCallback(() => (
        <View style={styles.resultsHeader}>
            {results.length > 0 && (
                <View style={styles.resultsCountContainer}>
                    <Text style={styles.resultsEmoji}>🎯</Text>
                    <Text style={styles.resultsCount}>
                        Found {results.length} player{results.length !== 1 ? 's' : ''}
                    </Text>
                </View>
            )}
        </View>
    ), [results.length]);

    return (
        <View style={styles.container}>
            <LinearGradient
                colors={['#0a0a0f', '#12121a', '#0a0a0f']}
                style={StyleSheet.absoluteFill}
            />

            <View style={styles.header}>
                <LinearGradient
                    colors={['rgba(139, 92, 246, 0.1)', 'transparent']}
                    style={styles.headerGradient}
                />
                <View style={styles.headerContent}>
                    <TouchableOpacity onPress={onNavigateBack} style={styles.backButton}>
                        <Text style={styles.backButtonText}>← Back</Text>
                    </TouchableOpacity>
                    <Text style={styles.title}>Search Players</Text>
                    <View style={styles.placeholder} />
                </View>
            </View>

            <View style={styles.searchContainer}>
                <View style={styles.searchInputWrapper}>
                    <Text style={styles.searchInputIcon}>🔍</Text>
                    <TextInput
                        style={styles.searchInput}
                        value={query}
                        onChangeText={handleQueryChange}
                        placeholder="Search by username (e.g., rahul)"
                        placeholderTextColor="#71717a"
                    />
                    {query.length > 0 && (
                        <TouchableOpacity onPress={clearSearch} style={styles.clearButton}>
                            <Text style={styles.clearButtonText}>✕</Text>
                        </TouchableOpacity>
                    )}
                </View>
            </View>

            <FlatList
                data={results}
                renderItem={renderItem}
                keyExtractor={keyExtractor}
                ListHeaderComponent={renderHeader}
                ListEmptyComponent={renderEmpty}
                contentContainerStyle={results.length === 0 ? styles.emptyList : styles.listContent}
                initialNumToRender={20}
                maxToRenderPerBatch={20}
                windowSize={10}
            />
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#0a0a0f',
    },
    header: {
        paddingTop: 60,
        paddingBottom: 16,
        position: 'relative',
        overflow: 'hidden',
    },
    headerGradient: {
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: 150,
    },
    headerContent: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
        paddingHorizontal: 20,
    },
    backButton: {
        paddingVertical: 8,
        paddingRight: 16,
    },
    backButtonText: {
        color: '#8b5cf6',
        fontSize: 16,
        fontWeight: '600',
    },
    title: {
        fontSize: 20,
        fontWeight: '700',
        color: '#ffffff',
    },
    placeholder: {
        width: 60,
    },
    searchContainer: {
        paddingHorizontal: 20,
        paddingBottom: 16,
    },
    searchInputWrapper: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: 'rgba(255, 255, 255, 0.05)',
        borderRadius: 16,
        borderWidth: 1,
        borderColor: 'rgba(255, 255, 255, 0.08)',
        paddingHorizontal: 16,
    },
    searchInputIcon: {
        fontSize: 18,
        marginRight: 12,
    },
    searchInput: {
        flex: 1,
        paddingVertical: 16,
        fontSize: 16,
        color: '#ffffff',
    },
    clearButton: {
        padding: 8,
    },
    clearButtonText: {
        color: '#71717a',
        fontSize: 16,
    },
    listContent: {
        paddingHorizontal: 20,
    },
    resultsHeader: {
        paddingVertical: 12,
    },
    resultsCountContainer: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 8,
    },
    resultsEmoji: {
        fontSize: 14,
    },
    resultsCount: {
        color: '#a1a1aa',
        fontSize: 14,
        fontWeight: '500',
    },
    emptyList: {
        flex: 1,
    },
    emptyContainer: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 40,
    },
    searchIconContainer: {
        marginBottom: 20,
    },
    searchIconGradient: {
        width: 80,
        height: 80,
        borderRadius: 24,
        alignItems: 'center',
        justifyContent: 'center',
    },
    searchIconEmoji: {
        fontSize: 36,
    },
    emptyEmoji: {
        fontSize: 56,
    },
    emptyTitle: {
        fontSize: 20,
        fontWeight: '700',
        color: '#ffffff',
        marginTop: 16,
    },
    emptyMessage: {
        fontSize: 14,
        color: '#71717a',
        marginTop: 8,
        textAlign: 'center',
        lineHeight: 22,
    },
});
