'use client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Nav from '@/components/Nav';
import { api, User } from '@/lib/api';

export default function AccountPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.get('/auth/me')
      .then((r) => setUser(r.data.data))
      .catch(() => router.push('/login'))
      .finally(() => setLoading(false));
  }, [router]);

  const handleLogout = async () => {
    try {
      await api.post('/auth/logout', {
        refresh_token: typeof window !== 'undefined' ? localStorage.getItem('refresh_token') : null,
      });
    } catch { /* ignore */ }
    if (typeof window !== 'undefined') localStorage.clear();
    router.push('/login');
  };

  if (loading) return <div className="loading">Loading...</div>;
  if (!user) return null;

  return (
    <>
      <Nav username={user.username} role={user.role} />
      <div className="container page">
        <h1 className="page-title">Account</h1>
        <div className="card" style={{ maxWidth: 480 }}>
          {user.avatar_url && (
            <img
              src={user.avatar_url}
              alt={user.username}
              style={{ width: 80, height: 80, borderRadius: '50%', marginBottom: 20 }}
            />
          )}
          <table style={{ marginBottom: 24 }}>
            <tbody>
              <tr><td style={{ color: '#888', paddingRight: 24 }}>Username</td><td>@{user.username}</td></tr>
              <tr><td style={{ color: '#888' }}>Email</td><td>{user.email || '—'}</td></tr>
              <tr><td style={{ color: '#888' }}>Role</td><td><span className={`badge badge-${user.role}`}>{user.role}</span></td></tr>
              <tr><td style={{ color: '#888' }}>ID</td><td style={{ fontFamily: 'monospace', fontSize: 12 }}>{user.id}</td></tr>
            </tbody>
          </table>
          <button className="btn btn-danger" onClick={handleLogout}>
            Log out
          </button>
        </div>
      </div>
    </>
  );
}
