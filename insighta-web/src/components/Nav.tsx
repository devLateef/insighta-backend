'use client';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';

export default function Nav({ username, role }: { username: string; role: string }) {
  const router = useRouter();

  const handleLogout = async () => {
    try {
      await api.post('/auth/logout', {
        refresh_token: typeof window !== 'undefined' ? localStorage.getItem('refresh_token') : null,
      });
    } catch { /* ignore */ }
    if (typeof window !== 'undefined') localStorage.clear();
    router.push('/login');
  };

  return (
    <nav className="nav">
      <span className="nav-brand">Insighta Labs+</span>
      <div className="nav-links">
        <Link href="/dashboard">Dashboard</Link>
        <Link href="/profiles">Profiles</Link>
        <Link href="/search">Search</Link>
        <Link href="/account">Account</Link>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
        <span style={{ fontSize: 13, color: '#888' }}>
          @{username} &nbsp;
          <span className={`badge badge-${role}`}>{role}</span>
        </span>
        <button className="btn btn-ghost" style={{ padding: '6px 14px' }} onClick={handleLogout}>
          Logout
        </button>
      </div>
    </nav>
  );
}
