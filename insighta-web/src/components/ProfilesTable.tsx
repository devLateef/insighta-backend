'use client';
import { Profile } from '@/lib/api';

interface Props {
  profiles: Profile[];
  onRowClick?: (id: string) => void;
}

export default function ProfilesTable({ profiles, onRowClick }: Props) {
  if (profiles.length === 0) {
    return <p style={{ color: '#888', padding: '24px 0' }}>No profiles found.</p>;
  }

  return (
    <table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Gender</th>
          <th>Age</th>
          <th>Age Group</th>
          <th>Country</th>
          <th>G. Prob</th>
          <th>C. Prob</th>
        </tr>
      </thead>
      <tbody>
        {profiles.map((p) => (
          <tr
            key={p.id}
            style={{ cursor: onRowClick ? 'pointer' : 'default' }}
            onClick={() => onRowClick?.(p.id)}
          >
            <td style={{ fontWeight: 500 }}>{p.name}</td>
            <td>
              <span className={`badge badge-${p.gender}`}>{p.gender}</span>
            </td>
            <td>{p.age}</td>
            <td>{p.age_group}</td>
            <td>{p.country_name} ({p.country_id})</td>
            <td>{(p.gender_probability * 100).toFixed(0)}%</td>
            <td>{(p.country_probability * 100).toFixed(0)}%</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
