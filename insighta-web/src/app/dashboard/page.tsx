'use client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Nav from '@/components/Nav';
import { api, User, ProfilesResponse } from '@/lib/api';

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [stats, setStats] = useState({ total: 0, male: 0, female: 0 });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if we have a token at all before making requests
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.replace('/login');
      return;
    }

    const load = async () => {
      try {
        // Sequential: get user first, then stats
        // The axios interceptor handles 401 → refresh automatically
        const meRes = await api.get('/auth/me');
        setUser(meRes.data.data);

        // Sequential to avoid exhausting Neon free tier connection pool
        const allRes = await api.get<ProfilesResponse>('/api/profiles', { params: { limit: 1 } });
        const maleRes = await api.get<ProfilesResponse>('/api/profiles', { params: { gender: 'male', limit: 1 } });
        const femaleRes = await api.get<ProfilesResponse>('/api/profiles', { params: { gender: 'female', limit: 1 } });

        setStats({
          total: allRes.data.total,
          male: maleRes.data.total,
          female: femaleRes.data.total,
        });
      } catch {
        // Only redirect if we truly can't authenticate
        // The interceptor already tried to refresh before we get here
        router.replace('/login');
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
