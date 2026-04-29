// Read the CSRF token from the csrf_token cookie (set by backend, not HttpOnly)
export function getCSRFToken(): string {
  if (typeof document === 'undefined') return '';
  const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}

// Attach CSRF token to mutating requests
export function withCSRF(headers: Record<string, string> = {}): Record<string, string> {
  const token = getCSRFToken();
  if (token) {
    headers['X-CSRF-Token'] = token;
  }
  return headers;
}
