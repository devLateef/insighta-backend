import Conf from 'conf';
import * as os from 'os';
import * as path from 'path';
import * as fs from 'fs';

export interface Credentials {
  access_token: string;
  refresh_token: string;
  username: string;
  role: string;
  id: string;
}

// Store credentials at ~/.insighta/credentials.json per spec
const CREDS_DIR = path.join(os.homedir(), '.insighta');
const CREDS_FILE = path.join(CREDS_DIR, 'credentials.json');

export const appConf = new Conf<{ api_url: string }>({
  projectName: 'insighta',
  defaults: {
    api_url: process.env.INSIGHTA_API_URL || 'http://localhost:8080',
  },
});

export function getApiUrl(): string {
  return appConf.get('api_url');
}

export function saveCredentials(creds: Credentials): void {
  if (!fs.existsSync(CREDS_DIR)) {
    fs.mkdirSync(CREDS_DIR, { recursive: true });
  }
  fs.writeFileSync(CREDS_FILE, JSON.stringify(creds, null, 2), { mode: 0o600 });
}

export function loadCredentials(): Credentials | null {
  try {
    if (!fs.existsSync(CREDS_FILE)) return null;
    const raw = fs.readFileSync(CREDS_FILE, 'utf-8');
    return JSON.parse(raw) as Credentials;
  } catch {
    return null;
  }
}

export function clearCredentials(): void {
  try {
    if (fs.existsSync(CREDS_FILE)) {
      fs.unlinkSync(CREDS_FILE);
    }
  } catch {
    // ignore
  }
}

export function requireAuth(): Credentials {
  const creds = loadCredentials();
  if (!creds) {
    console.error('Not logged in. Run: insighta login');
    process.exit(1);
  }
  return creds;
}
