import Table from 'cli-table3';
import chalk from 'chalk';

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

export function printProfilesTable(profiles: Profile[]): void {
  if (profiles.length === 0) {
    console.log(chalk.yellow('No profiles found.'));
    return;
  }

  const table = new Table({
    head: [
      chalk.cyan('Name'),
      chalk.cyan('Gender'),
      chalk.cyan('Age'),
      chalk.cyan('Age Group'),
      chalk.cyan('Country'),
      chalk.cyan('G.Prob'),
      chalk.cyan('C.Prob'),
    ],
    colWidths: [20, 8, 6, 12, 20, 8, 8],
    style: { head: [], border: [] },
  });

  for (const p of profiles) {
    table.push([
      p.name,
      p.gender,
      String(p.age),
      p.age_group,
      `${p.country_name} (${p.country_id})`,
      p.gender_probability.toFixed(2),
      p.country_probability.toFixed(2),
    ]);
  }

  console.log(table.toString());
}

export function printProfileDetail(p: Profile): void {
  const table = new Table({ style: { head: [], border: [] } });
  table.push(
    { ID: p.id },
    { Name: p.name },
    { Gender: `${p.gender} (${p.gender_probability.toFixed(2)})` },
    { Age: `${p.age} (${p.age_group})` },
    { Country: `${p.country_name} (${p.country_id}) — prob: ${p.country_probability.toFixed(2)}` },
    { 'Created At': new Date(p.created_at).toLocaleString() }
  );
  console.log(table.toString());
}

export function printPagination(page: number, limit: number, total: number, totalPages: number): void {
  console.log(
    chalk.dim(`\nPage ${page}/${totalPages} · ${total} total results · ${limit} per page`)
  );
}
