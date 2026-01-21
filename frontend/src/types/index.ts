export interface User {
  id: string;
  username: string;
  rating: number;
}

export interface UserWithRank extends User {
  rank: number;
}

export interface LeaderboardResponse {
  users: UserWithRank[];
  total_users: number;
  page: number;
  page_size: number;
  has_more: boolean;
}

export interface SearchResponse {
  users: UserWithRank[];
  query: string;
  count: number;
}

export interface HealthResponse {
  status: string;
  total_users: number;
}
