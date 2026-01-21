import React from 'react';
import { View, TextInput, TouchableOpacity, Text, StyleSheet } from 'react-native';

interface SearchBarProps {
    value: string;
    onChangeText: (text: string) => void;
    onClear: () => void;
    placeholder?: string;
}

export function SearchBar({ value, onChangeText, onClear, placeholder = 'Search by username...' }: SearchBarProps) {
    return (
        <View style={styles.container}>
            <View style={styles.searchIcon}>
                <Text style={styles.iconText}>🔍</Text>
            </View>
            <TextInput
                style={styles.input}
                value={value}
                onChangeText={onChangeText}
                placeholder={placeholder}
                placeholderTextColor="#71717a"
                autoCapitalize="none"
                autoCorrect={false}
            />
            {value.length > 0 && (
                <TouchableOpacity onPress={onClear} style={styles.clearButton}>
                    <Text style={styles.clearText}>✕</Text>
                </TouchableOpacity>
            )}
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: 'rgba(255, 255, 255, 0.05)',
        borderRadius: 16,
        paddingHorizontal: 16,
        marginHorizontal: 20,
        marginVertical: 12,
        borderWidth: 1,
        borderColor: 'rgba(255, 255, 255, 0.08)',
    },
    searchIcon: {
        marginRight: 12,
    },
    iconText: {
        fontSize: 18,
    },
    input: {
        flex: 1,
        paddingVertical: 16,
        fontSize: 16,
        color: '#ffffff',
    },
    clearButton: {
        padding: 8,
    },
    clearText: {
        fontSize: 16,
        color: '#71717a',
    },
});
