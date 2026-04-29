import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const api = axios.create({
  baseURL: API_URL,
  headers: { 'X-API-Version': '1' },
});

// Attach token from localStorage on every request
api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
  }
  return config;
});

let isRefreshing = false;
let refreshQueue: Array<(token: string) => void> = [];

// Auto-refresh on 401
api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config;

    if (error.response?.status === 401 && !original._retry) {
      original._retry = true;

      if (isRefreshing) {
        // Queue this request until refresh completes
        return new Promise((resolve) => {
          refreshQueue.push((token: string) => {
            original.headers['Authorization'] = `Bearer ${token}`;
            resolve(api(original));
          });
        });
      }

      isRefreshing = true;

      try {
        const refreshToken = typeof window !== 'undefined'
          ? localStorage.getItem('refresh_token')
          : null;

        if (!refreshToken) throw new Error('no refresh token');

        const res = await axios.post(`${API_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });

        const { access_token, refresh_token } = res.data;
        localStorage.setItem('access_token', access_token);
        localStorage.setItem('refresh_token', refresh_token);

        // Flush queued requests
        refreshQueue.forEach((cb) => cb(access_token));
        refreshQueue = [];

        original.headers['Authorization'] = `Bearer ${access_token}`;
        return api(original);
      } catch {
        // Refresh failed — clear and redirect to login
        refreshQueue = [];
        if (typeof window !== 'undefined') {
          localStorage.clear();
          window.location.href = '/login';
        }
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

export interface Profile {
  id: string;
  name: string;
  gender: string;
  gender_probability: number;
  age: number;
  age_group: string;
  country_id: string;
  country_name: string;
  country_probability: number;
  created_at: string;
}

export interface ProfilesResponse {
  status: string;
  page: number;
  limit: number;
  total: number;
  total_pages: number;
  data: Profile[];
}

export interface User {
  id: string;
  username: string;
  email: string;
  avatar_url: string;
  role: string;
}
