import { LeaderboardResponse, SearchResponse, UserWithRank, HealthResponse } from '../types';

const getApiBaseUrl = (): string => {
    // Check for environment variable (Expo)
    if (typeof process !== 'undefined' && process.env?.EXPO_PUBLIC_API_URL) {
        return process.env.EXPO_PUBLIC_API_URL;
    }
    // Check for window-based config (web)
    if (typeof window !== 'undefined' && (window as any).__API_URL__) {
        return (window as any).__API_URL__;
    }
    // Default to localhost
    return 'http://localhost:8080/api';
};

const DEFAULT_TIMEOUT = 10000; // 10 seconds

class ApiService {
    private baseUrl: string;
    private timeout: number;

    constructor(baseUrl?: string, timeout: number = DEFAULT_TIMEOUT) {
        this.baseUrl = baseUrl || getApiBaseUrl();
        this.timeout = timeout;
    }

    private async fetchWithTimeout(url: string, options: RequestInit = {}): Promise<Response> {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        try {
            const response = await fetch(url, {
                ...options,
                headers: {
                    ...options.headers,
                    'ngrok-skip-browser-warning': 'true',
                },
                signal: controller.signal,
            });
            clearTimeout(timeoutId);
            return response;
        } catch (error) {
            clearTimeout(timeoutId);
            if (error instanceof Error && error.name === 'AbortError') {
                throw new Error('Request timed out. Please try again.');
            }
            throw error;
        }
    }

    async getLeaderboard(limit: number = 50, offset: number = 0): Promise<LeaderboardResponse> {
        const response = await this.fetchWithTimeout(
            `${this.baseUrl}/leaderboard?limit=${limit}&offset=${offset}`
        );
        if (!response.ok) {
            throw new Error(`Failed to fetch leaderboard: ${response.statusText}`);
        }
        return response.json();
    }

    async searchUsers(query: string): Promise<SearchResponse> {
        const sanitizedQuery = encodeURIComponent(query.trim());
        const response = await this.fetchWithTimeout(
            `${this.baseUrl}/search?q=${sanitizedQuery}`
        );
        if (!response.ok) {
            throw new Error(`Failed to search users: ${response.statusText}`);
        }
        return response.json();
    }

    async getUser(id: string): Promise<UserWithRank> {
        const response = await this.fetchWithTimeout(`${this.baseUrl}/users/${id}`);
        if (!response.ok) {
            throw new Error(`Failed to get user: ${response.statusText}`);
        }
        return response.json();
    }

    async seedUsers(count: number = 10000): Promise<{ message: string; users_added: number }> {
        const response = await this.fetchWithTimeout(`${this.baseUrl}/seed?count=${count}`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error(`Failed to seed users: ${response.statusText}`);
        }
        return response.json();
    }

    async getHealth(): Promise<HealthResponse> {
        const response = await this.fetchWithTimeout(`${this.baseUrl}/health`);
        if (!response.ok) {
            throw new Error(`Failed to get health: ${response.statusText}`);
        }
        return response.json();
    }

    async startSimulator(): Promise<{ message: string; running: boolean }> {
        const response = await this.fetchWithTimeout(`${this.baseUrl}/simulator/start`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error(`Failed to start simulator: ${response.statusText}`);
        }
        return response.json();
    }

    async stopSimulator(): Promise<{ message: string; running: boolean }> {
        const response = await this.fetchWithTimeout(`${this.baseUrl}/simulator/stop`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error(`Failed to stop simulator: ${response.statusText}`);
        }
        return response.json();
    }
}

export const api = new ApiService();
export default api;
