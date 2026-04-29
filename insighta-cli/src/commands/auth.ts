import * as http from 'http';
import * as crypto from 'crypto';
import * as url from 'url';
import chalk from 'chalk';
import ora from 'ora';
import open from 'open';
import { getApiUrl, saveCredentials, clearCredentials, loadCredentials } from '../lib/config';
import { getClient, handleApiError } from '../lib/api';

// ─── State helper ─────────────────────────────────────────────────────────────

function generateState(): string {
  return crypto.randomBytes(32).toString('base64url');
}

// ─── insighta login ───────────────────────────────────────────────────────────
//
// Flow:
//  1. CLI starts a local HTTP server on a random port
//  2. CLI opens browser → backend /auth/github?state=...&redirect_uri=http://localhost:PORT/callback
//  3. Backend stores the redirect_uri in a cookie, then redirects to GitHub
//  4. User authorizes on GitHub
//  5. GitHub redirects to backend /auth/github/callback?code=...&state=...
//  6. Backend exchanges code, issues tokens, then redirects to CLI local server with tokens in query
//  7. CLI local server captures tokens, saves credentials, closes server

export async function login(): Promise<void> {
  const state = generateState();
  const port = await getFreePort();
  const localCallbackUrl = `http://localhost:${port}/callback`;

  const spinner = ora('Opening browser for GitHub login...').start();

  const credentials = await new Promise<{
    access_token: string;
    refresh_token: string;
    username: string;
    role: string;
    id: string;
  }>((resolve, reject) => {
    const server = http.createServer((req, res) => {
      const parsed = url.parse(req.url || '', true);
      if (parsed.pathname !== '/callback') return;

      const { access_token, refresh_token, username, role, id, error } = parsed.query as Record<string, string>;

      // Send a nice response to the browser
      res.writeHead(200, { 'Content-Type': 'text/html' });
      if (error) {
        res.end(`<html><body style="font-family:sans-serif;text-align:center;padding:40px">
          <h2>❌ Login failed</h2><p>${error}</p><p>You can close this tab.</p>
        </body></html>`);
        server.close();
        reject(new Error(error));
        return;
      }

      res.end(`<html><body style="font-family:sans-serif;text-align:center;padding:40px">
        <h2>✅ Logged in as @${username}!</h2>
        <p>You can close this tab and return to the terminal.</p>
      </body></html>`);

      server.close();
      resolve({ access_token, refresh_token, username, role, id });
    });

    server.listen(port, () => {
      // Build the URL that opens the backend OAuth flow
      const authUrl = new URL(`${getApiUrl()}/auth/github`);
      authUrl.searchParams.set('state', state);
      authUrl.searchParams.set('redirect_uri', localCallbackUrl);

      spinner.text = 'Waiting for GitHub authorization...';

      open(authUrl.toString()).catch(() => {
        spinner.info(`Open this URL in your browser:\n${authUrl.toString()}`);
      });
    });

    server.on('error', reject);

    // Timeout after 5 minutes
    setTimeout(() => {
      server.close();
      reject(new Error('Login timed out after 5 minutes'));
    }, 5 * 60 * 1000);
  });

  saveCredentials(credentials);
  spinner.succeed(chalk.green(`Logged in as @${credentials.username} (${credentials.role})`));
}

// ─── insighta logout ──────────────────────────────────────────────────────────

export async function logout(): Promise<void> {
  const creds = loadCredentials();
  if (!creds) {
    console.log(chalk.yellow('Not logged in.'));
    return;
  }

  const spinner = ora('Logging out...').start();
  try {
    await getClient().post('/auth/logout', { refresh_token: creds.refresh_token });
  } catch {
    // Ignore errors — clear locally regardless
  }

  clearCredentials();
  spinner.succeed(chalk.green('Logged out successfully.'));
}

// ─── insighta whoami ──────────────────────────────────────────────────────────

export async function whoami(): Promise<void> {
  const spinner = ora('Fetching user info...').start();
  try {
    const res = await getClient().get('/auth/me');
    spinner.stop();
    const u = res.data.data;
    console.log(chalk.bold('Logged in as:'));
    console.log(`  Username : ${chalk.cyan('@' + u.username)}`);
    console.log(`  Role     : ${chalk.yellow(u.role)}`);
    console.log(`  Email    : ${u.email || '(not set)'}`);
    console.log(`  ID       : ${u.id}`);
  } catch (err) {
    spinner.stop();
    handleApiError(err);
  }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function getFreePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = http.createServer();
    srv.listen(0, () => {
      const addr = srv.address() as { port: number };
      srv.close(() => resolve(addr.port));
    });
    srv.on('error', reject);
  });
}
