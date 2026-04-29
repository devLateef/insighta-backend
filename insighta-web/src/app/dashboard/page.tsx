'use client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Nav from '@/components/Nav';
import { api, User, ProfilesResponse } from '@/lib/api';

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [stats, setStats] = useState({ total: 0, male: 0, female: 0, countries: 0 });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const [meRes, allRes, maleRes, femaleRes] = await Promise.all([
          api.get('/auth/me'),
          api.get<ProfilesResponse>('/api/profiles', { params: { limit: 1 } }),
          api.get<ProfilesResponse>('/api/profiles', { params: { gender: 'male', limit: 1 } }),
          api.get<ProfilesResponse>('/api/profiles', { params: { gender: 'female', limit: 1 } }),
        ]);

        setUser(meRes.data.data);
        setStats({
          total: allRes.data.total,
          male: maleRes.data.total,
          female: femaleRes.data.total,
          countries: 0, // approximate
        });
      } catch {
        router.push('/login');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [router]);

  if (loading) return <div className="loading">Loading...</div>;
  if (!user) return null;

  return (
    <>
      <Nav username={user.username} role={user.role} />
      <div className="container page">
        <h1 className="page-title">Dashboard</h1>
        <div className="stat-grid">
          <div className="stat-card">
            <div className="value">{stats.total.toLocaleString()}</div>
            <div className="label">Total Profiles</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.male.toLocaleString()}</div>
            <div className="label">Male Profiles</div>
          </div>
          <div className="stat-card">
            <div className="value">{stats.female.toLocaleString()}</div>
            <div className="label">Female Profiles</div>
          </div>
          <div className="stat-card">
            <div className="value">{user.role}</div>
            <div className="label">Your Role</div>
          </div>
        </div>

        <div className="card">
          <h2 style={{ marginBottom: 16, fontSize: 18 }}>Quick Actions</h2>
          <div style={{ display: 'flex', gap: 12 }}>
            <button className="btn btn-primary" onClick={() => router.push('/profiles')}>
              Browse Profiles
            </button>
            <button className="btn btn-ghost" onClick={() => router.push('/search')}>
              Search Profiles
            </button>
          </div>
        </div>
      </div>
    </>
  );
}
