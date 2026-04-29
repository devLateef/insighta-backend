#!/usr/bin/env node
import { Command } from 'commander';
import { login, logout, whoami } from './commands/auth';
import {
  listProfiles,
  getProfile,
  searchProfiles,
  createProfile,
  exportProfiles,
} from './commands/profiles';

const program = new Command();

program
  .name('insighta')
  .description('Insighta Labs CLI — Profile Intelligence Platform')
  .version('1.0.0');

// ─── Auth commands ────────────────────────────────────────────────────────────

program
  .command('login')
  .description('Log in via GitHub OAuth (PKCE)')
  .action(async () => {
    try {
      await login();
    } catch (err: any) {
      console.error('Login failed:', err.message);
      process.exit(1);
    }
  });

program
  .command('logout')
  .description('Log out and clear stored credentials')
  .action(async () => {
    await logout();
  });

program
  .command('whoami')
  .description('Show current logged-in user')
  .action(async () => {
    await whoami();
  });

// ─── Profiles commands ────────────────────────────────────────────────────────

const profiles = program.command('profiles').description('Manage profiles');

profiles
  .command('list')
  .description('List profiles with optional filters')
  .option('--gender <gender>', 'Filter by gender (male/female)')
  .option('--country <code>', 'Filter by ISO country code (e.g. NG)')
  .option('--age-group <group>', 'Filter by age group (child/teenager/adult/senior)')
  .option('--min-age <age>', 'Minimum age')
  .option('--max-age <age>', 'Maximum age')
  .option('--sort-by <field>', 'Sort by: age, created_at, gender_probability', 'created_at')
  .option('--order <dir>', 'Sort direction: asc or desc', 'desc')
  .option('--page <n>', 'Page number', '1')
  .option('--limit <n>', 'Results per page (max 50)', '10')
  .action(async (opts) => {
    await listProfiles({
      gender:   opts.gender,
      country:  opts.country,
      ageGroup: opts.ageGroup,
      minAge:   opts.minAge,
      maxAge:   opts.maxAge,
      sortBy:   opts.sortBy,
      order:    opts.order,
      page:     opts.page,
      limit:    opts.limit,
    });
  });

profiles
  .command('get <id>')
  .description('Get a single profile by ID')
  .action(async (id: string) => {
    await getProfile(id);
  });

profiles
  .command('search <query>')
  .description('Natural language search (e.g. "young males from nigeria")')
  .option('--page <n>', 'Page number', '1')
  .option('--limit <n>', 'Results per page', '10')
  .action(async (query: string, opts: { page: string; limit: string }) => {
    await searchProfiles(query, { page: opts.page, limit: opts.limit });
  });

profiles
  .command('create')
  .description('Create a new profile (admin only)')
  .requiredOption('--name <name>', 'Full name of the person')
  .action(async (opts: { name: string }) => {
    await createProfile(opts.name);
  });

profiles
  .command('export')
  .description('Export profiles to CSV')
  .option('--format <fmt>', 'Export format (csv)', 'csv')
  .option('--gender <gender>', 'Filter by gender')
  .option('--country <code>', 'Filter by country code')
  .option('--age-group <group>', 'Filter by age group')
  .option('--min-age <age>', 'Minimum age')
  .option('--max-age <age>', 'Maximum age')
  .action(async (opts) => {
    await exportProfiles({
      format:   opts.format,
      gender:   opts.gender,
      country:  opts.country,
      ageGroup: opts.ageGroup,
      minAge:   opts.minAge,
      maxAge:   opts.maxAge,
    });
  });

program.parse(process.argv);
