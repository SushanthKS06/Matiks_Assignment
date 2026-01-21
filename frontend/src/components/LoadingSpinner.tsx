import React from 'react';
import { View, Text, StyleSheet, ActivityIndicator } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';

interface LoadingSpinnerProps {
    message?: string;
    fullScreen?: boolean;
}

export function LoadingSpinner({ message = 'Loading...', fullScreen = false }: LoadingSpinnerProps) {
    const content = (
        <View style={styles.container}>
            <View style={styles.spinnerContainer}>
                <ActivityIndicator size="large" color="#8b5cf6" />
            </View>
            <Text style={styles.message}>{message}</Text>
        </View>
    );

    if (fullScreen) {
        return (
            <View style={styles.fullScreen}>
                <LinearGradient
                    colors={['#0a0a0f', '#12121a']}
                    style={StyleSheet.absoluteFill}
                />
                {content}
            </View>
        );
    }

    return content;
}

const styles = StyleSheet.create({
    container: {
        alignItems: 'center',
        justifyContent: 'center',
        padding: 32,
    },
    fullScreen: {
        flex: 1,
        alignItems: 'center',
        justifyContent: 'center',
    },
    spinnerContainer: {
        width: 64,
        height: 64,
        borderRadius: 20,
        backgroundColor: 'rgba(139, 92, 246, 0.1)',
        alignItems: 'center',
        justifyContent: 'center',
        marginBottom: 16,
    },
    message: {
        fontSize: 14,
        color: '#71717a',
        fontWeight: '500',
    },
});
