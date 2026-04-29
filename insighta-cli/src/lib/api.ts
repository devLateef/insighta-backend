import axios, { AxiosInstance, AxiosError } from 'axios';
import { getApiUrl, loadCredentials, saveCredentials, clearCredentials } from './config';

let _client: AxiosInstance | null = null;

export function getClient(): AxiosInstance {
  if (_client) return _client;

  _client = axios.create({
    baseURL: getApiUrl(),
    headers: {
      'X-API-Version': '1',
      'Content-Type': 'application/json',
    },
    timeout: 15000,
  });

  // Attach access token to every request
  _client.interceptors.request.use((config) => {
    const creds = loadCredentials();
    if (creds?.access_token) {
      config.headers['Authorization'] = `Bearer ${creds.access_token}`;
    }
    return config;
  });

  // Auto-refresh on 401
  _client.interceptors.response.use(
    (res) => res,
    async (error: AxiosError) => {
      const original = error.config as any;
      if (error.response?.status === 401 && !original._retry) {
        original._retry = true;
        const creds = loadCredentials();
        if (!creds?.refresh_token) {
          clearCredentials();
          console.error('\nSession expired. Please run: insighta login');
          process.exit(1);
        }
        try {
          const res = await axios.post(`${getApiUrl()}/auth/refresh`, {
            refresh_token: creds.refresh_token,
          });
          const { access_token, refresh_token } = res.data;
          saveCredentials({ ...creds, access_token, refresh_token });
          original.headers['Authorization'] = `Bearer ${access_token}`;
          return _client!(original);
        } catch {
          clearCredentials();
          console.error('\nSession expired. Please run: insighta login');
          process.exit(1);
        }
      }
      return Promise.reject(error);
    }
  );

  return _client;
}

export function handleApiError(error: unknown): never {
  if (axios.isAxiosError(error)) {
    const msg = (error.response?.data as any)?.message || error.message;
    console.error(`Error: ${msg}`);
  } else {
    console.error('Unexpected error:', error);
  }
  process.exit(1);
}
