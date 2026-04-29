# Insighta CLI

A globally installable CLI tool for the Insighta Labs Profile Intelligence Platform.

## Installation

```bash
npm install
npm run build
npm install -g .
```

After installation, `insighta` is available globally from any directory.

## Configuration

Set the API URL (defaults to `http://localhost:8080`):

```bash
export INSIGHTA_API_URL=https://your-backend-url.com
```

Credentials are stored at `~/.insighta/credentials.json`.

## Commands

### Auth

```bash
insighta login          # Log in via GitHub OAuth (PKCE)
insighta logout         # Log out and clear credentials
insighta whoami         # Show current user info
```

### Profiles

```bash
# List profiles
insighta profiles list
insighta profiles list --gender male
insighta profiles list --country NG --age-group adult
insighta profiles list --min-age 25 --max-age 40
insighta profiles list --sort-by age --order desc
insighta profiles list --page 2 --limit 20

# Get single profile
insighta profiles get <id>

# Natural language search
insighta profiles search "young males from nigeria"
insighta profiles search "adult females from kenya" --page 2

# Create profile (admin only)
insighta profiles create --name "Harriet Tubman"

# Export to CSV
insighta profiles export --format csv
insighta profiles export --format csv --gender male --country NG
```

## Token Handling

- Access tokens expire in 3 minutes
- On a 401 response, the CLI automatically uses the refresh token to get a new pair
- If the refresh token is also expired, the CLI prompts you to run `insighta login` again
- Tokens are stored at `~/.insighta/credentials.json` with `600` permissions
