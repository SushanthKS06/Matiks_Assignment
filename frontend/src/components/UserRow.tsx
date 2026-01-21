import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { UserWithRank } from '../types';

interface UserRowProps {
    user: UserWithRank;
    index: number;
}

export function UserRow({ user, index }: UserRowProps) {
    const getRatingPercentage = (rating: number) => {
        return ((rating - 100) / (5000 - 100)) * 100;
    };

    const getInitials = (username: string) => {
        return username.substring(0, 2).toUpperCase();
    };

    const renderRankBadge = () => {
        if (user.rank === 1) {
            return (
                <LinearGradient colors={['#ffd700', '#ff8c00']} style={styles.rankBadgeGold}>
                    <Text style={styles.crownIcon}>👑</Text>
                    <Text style={styles.rankTextDark}>1</Text>
                </LinearGradient>
            );
        } else if (user.rank === 2) {
            return (
                <LinearGradient colors={['#e8e8e8', '#a8a8a8']} style={styles.rankBadgeSilver}>
                    <Text style={styles.rankTextDark}>2</Text>
                </LinearGradient>
            );
        } else if (user.rank === 3) {
            return (
                <LinearGradient colors={['#cd7f32', '#8b4513']} style={styles.rankBadgeBronze}>
                    <Text style={styles.rankTextLight}>3</Text>
                </LinearGradient>
            );
        } else {
            return (
                <View style={styles.rankBadgeDefault}>
                    <Text style={styles.rankTextMuted}>#{user.rank}</Text>
                </View>
            );
        }
    };

    const isTopThree = user.rank <= 3;

    return (
        <View style={[styles.container, isTopThree && styles.containerTopThree]}>
            {isTopThree && (
                <LinearGradient
                    colors={['rgba(139, 92, 246, 0.1)', 'transparent']}
                    start={{ x: 0, y: 0.5 }}
                    end={{ x: 1, y: 0.5 }}
                    style={styles.topThreeGradient}
                />
            )}
            <View style={styles.rankContainer}>
                {renderRankBadge()}
            </View>

            <View style={styles.userInfo}>
                <LinearGradient
                    colors={['#8b5cf6', '#06b6d4']}
                    style={styles.avatar}
                >
                    <Text style={styles.avatarText}>{getInitials(user.username)}</Text>
                </LinearGradient>
                <View style={styles.userDetails}>
                    <Text style={styles.username}>{user.username}</Text>
                    <Text style={styles.userId}>{user.id.slice(0, 8)}...</Text>
                </View>
            </View>

            <View style={styles.ratingContainer}>
                <Text style={styles.ratingValue}>{user.rating.toLocaleString()}</Text>
                <View style={styles.ratingBar}>
                    <LinearGradient
                        colors={['#8b5cf6', '#06b6d4']}
                        start={{ x: 0, y: 0 }}
                        end={{ x: 1, y: 0 }}
                        style={[styles.ratingFill, { width: `${getRatingPercentage(user.rating)}%` }]}
                    />
                </View>
            </View>
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flexDirection: 'row',
        alignItems: 'center',
        paddingVertical: 14,
        paddingHorizontal: 20,
        borderBottomWidth: 1,
        borderBottomColor: 'rgba(255, 255, 255, 0.03)',
        position: 'relative',
        overflow: 'hidden',
    },
    containerTopThree: {
        borderBottomColor: 'rgba(139, 92, 246, 0.1)',
    },
    topThreeGradient: {
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
    },
    rankContainer: {
        width: 56,
        alignItems: 'center',
    },
    rankBadgeGold: {
        width: 44,
        height: 44,
        borderRadius: 14,
        alignItems: 'center',
        justifyContent: 'center',
        shadowColor: '#ffd700',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.4,
        shadowRadius: 8,
        elevation: 8,
        position: 'relative',
    },
    rankBadgeSilver: {
        width: 44,
        height: 44,
        borderRadius: 14,
        alignItems: 'center',
        justifyContent: 'center',
        shadowColor: '#c0c0c0',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.3,
        shadowRadius: 8,
        elevation: 6,
    },
    rankBadgeBronze: {
        width: 44,
        height: 44,
        borderRadius: 14,
        alignItems: 'center',
        justifyContent: 'center',
        shadowColor: '#cd7f32',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.3,
        shadowRadius: 8,
        elevation: 6,
    },
    rankBadgeDefault: {
        width: 44,
        height: 44,
        borderRadius: 14,
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'rgba(255, 255, 255, 0.03)',
        borderWidth: 1,
        borderColor: 'rgba(255, 255, 255, 0.08)',
    },
    crownIcon: {
        position: 'absolute',
        top: -10,
        fontSize: 12,
    },
    rankTextDark: {
        fontSize: 18,
        fontWeight: '800',
        color: '#1a1a1a',
    },
    rankTextLight: {
        fontSize: 18,
        fontWeight: '800',
        color: '#ffffff',
    },
    rankTextMuted: {
        fontSize: 13,
        fontWeight: '600',
        color: '#71717a',
    },
    userInfo: {
        flex: 1,
        flexDirection: 'row',
        alignItems: 'center',
        gap: 14,
        marginLeft: 8,
    },
    avatar: {
        width: 42,
        height: 42,
        borderRadius: 12,
        alignItems: 'center',
        justifyContent: 'center',
    },
    avatarText: {
        fontSize: 15,
        fontWeight: '700',
        color: '#ffffff',
    },
    userDetails: {
        flex: 1,
    },
    username: {
        fontSize: 15,
        fontWeight: '600',
        color: '#ffffff',
        marginBottom: 2,
    },
    userId: {
        fontSize: 11,
        color: '#71717a',
        fontFamily: 'monospace',
    },
    ratingContainer: {
        alignItems: 'flex-end',
        minWidth: 80,
    },
    ratingValue: {
        fontSize: 18,
        fontWeight: '700',
        color: '#8b5cf6',
    },
    ratingBar: {
        width: 70,
        height: 4,
        backgroundColor: 'rgba(255, 255, 255, 0.05)',
        borderRadius: 2,
        marginTop: 6,
        overflow: 'hidden',
    },
    ratingFill: {
        height: '100%',
        borderRadius: 2,
    },
});
