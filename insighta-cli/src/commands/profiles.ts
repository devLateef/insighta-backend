import * as fs from 'fs';
import * as path from 'path';
import chalk from 'chalk';
import ora from 'ora';
import { getClient, handleApiError } from '../lib/api';
import { requireAuth } from '../lib/config';
import {
  printProfilesTable,
  printProfileDetail,
  printPagination,
  Profile,
} from '../lib/table';

// ─── insighta profiles list ───────────────────────────────────────────────────

export interface ListOptions {
  gender?: string;
  country?: string;
  ageGroup?: string;
  minAge?: string;
  maxAge?: string;
  sortBy?: string;
  order?: string;
  page?: string;
  limit?: string;
}

export async function listProfiles(opts: ListOptions): Promise<void> {
  requireAuth();
  const spinner = ora('Fetching profiles...').start();

  try {
    const params: Record<string, string> = {};
    if (opts.gender)   params.gender = opts.gender;
    if (opts.country)  params.country_id = opts.country.toUpperCase();
    if (opts.ageGroup) params.age_group = opts.ageGroup;
    if (opts.minAge)   params.min_age = opts.minAge;
    if (opts.maxAge)   params.max_age = opts.maxAge;
    if (opts.sortBy)   params.sort_by = opts.sortBy;
    if (opts.order)    params.order = opts.order;
    params.page  = opts.page  || '1';
    params.limit = opts.limit || '10';

    const res = await getClient().get('/api/profiles', { params });
    spinner.stop();

    const { data, page, limit, total, total_pages } = res.data;
    printProfilesTable(data as Profile[]);
    printPagination(page, limit, total, total_pages);
  } catch (err) {
    spinner.stop();
    handleApiError(err);
  }
}

// ─── insighta profiles get <id> ───────────────────────────────────────────────

export async function getProfile(id: string): Promise<void> {
  requireAuth();
  const spinner = ora(`Fetching profile ${id}...`).start();

  try {
    const res = await getClient().get(`/api/profiles/${id}`);
    spinner.stop();
    printProfileDetail(res.data.data as Profile);
  } catch (err) {
    spinner.stop();
    handleApiError(err);
  }
}

// ─── insighta profiles search ─────────────────────────────────────────────────

export async function searchProfiles(query: string, opts: { page?: string; limit?: string }): Promise<void> {
  requireAuth();
  const spinner = ora(`Searching: "${query}"...`).start();

  try {
    const params: Record<string, string> = { q: query };
    params.page  = opts.page  || '1';
    params.limit = opts.limit || '10';

    const res = await getClient().get('/api/profiles/search', { params });
    spinner.stop();

    const { data, page, limit, total, total_pages } = res.data;
    printProfilesTable(data as Profile[]);
    printPagination(page, limit, total, total_pages);
  } catch (err) {
    spinner.stop();
    handleApiError(err);
  }
}

// ─── insighta profiles create ─────────────────────────────────────────────────

export async function createProfile(name: string): Promise<void> {
  requireAuth();
  const spinner = ora(`Creating profile for "${name}"...`).start();

  try {
    const res = await getClient().post('/api/profiles', { name });
    spinner.stop();
    console.log(chalk.green('Profile created:'));
    printProfileDetail(res.data.data as Profile);
  } catch (err) {
    spinner.stop();
    handleApiError(err);
  }
}

// ─── insighta profiles export ─────────────────────────────────────────────────

export interface ExportOptions {
  format?: string;
  gender?: string;
  country?: string;
  ageGroup?: string;
  minAge?: string;
  maxAge?: string;
}

export async function exportProfiles(opts: ExportOptions): Promise<void> {
  requireAuth();
  const format = opts.format || 'csv';
  const spinner = ora(`Exporting profiles as ${format}...`).start();

  try {
    const params: Record<string, string> = { format };
    if (opts.gender)   params.gender = opts.gender;
    if (opts.country)  params.country_id = opts.country.toUpperCase();
    if (opts.ageGroup) params.age_group = opts.ageGroup;
    if (opts.minAge)   params.min_age = opts.minAge;
    if (opts.maxAge)   params.max_age = opts.maxAge;

    const res = await getClient().get('/api/profiles/export', {
      params,
      responseType: 'text',
    });

    spinner.stop();

    // Save to current working directory
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
    const filename = `profiles_${timestamp}.${format}`;
    const filepath = path.join(process.cwd(), filename);
    fs.writeFileSync(filepath, res.data as string, 'utf-8');

    console.log(chalk.green(`✓ Exported to: ${filepath}`));
  } catch (err) {
    spinner.stop();
    handleApiError(err);
  }
}
