'use client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Nav from '@/components/Nav';
import ProfilesTable from '@/components/ProfilesTable';
import { api, User, Profile, ProfilesResponse } from '@/lib/api';

export default function SearchPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [query, setQuery] = useState('');
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    api.get('/auth/me').then((r) => setUser(r.data.data)).catch(() => router.push('/login'));
  }, [router]);

  const handleSearch = async (p = 1) => {
    if (!query.trim()) return;
    setLoading(true);
    setError('');
    setSearched(true);
    try {
      const res = await api.get<ProfilesResponse>('/api/profiles/search', {
        params: { q: query, page: p, limit: 10 },
      });
      setProfiles(res.data.data);
      setTotal(res.data.total);
      setTotalPages(res.data.total_pages);
      setPage(p);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Search failed');
      setProfiles([]);
    } finally {
      setLoading(false);
    }
  };

  if (!user) return <div className="loading">Loading...</div>;

  return (
    <>
      <Nav username={user.username} role={user.role} />
      <div className="container page">
        <h1 className="page-title">Natural Language Search</h1>

        <div className="card" style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', gap: 12 }}>
            <input
              style={{ flex: 1, padding: '10px 14px', border: '1px solid #ddd', borderRadius: 8, fontSize: 15 }}
              placeholder='e.g. "young males from nigeria" or "adult females above 30"'
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch(1)}
            />
            <button className="btn btn-primary" onClick={() => handleSearch(1)} disabled={loading}>
              {loading ? 'Searching...' : 'Search'}
            </button>
          </div>
          {error && <p className="error">{error}</p>}
          <p style={{ marginTop: 12, fontSize: 13, color: '#888' }}>
            Try: &quot;young males from nigeria&quot; · &quot;adult females above 30&quot; · &quot;seniors from ghana&quot;
          </p>
        </div>

        {searched && (
          <div className="card">
            {loading ? (
              <div className="loading">Searching...</div>
            ) : (
              <>
                <p style={{ marginBottom: 16, color: '#555', fontSize: 14 }}>
                  {total} result{total !== 1 ? 's' : ''} for &quot;{query}&quot;
                </p>
                <ProfilesTable profiles={profiles} onRowClick={(id) => router.push(`/profiles/${id}`)} />
                {totalPages > 1 && (
                  <div className="pagination">
                    <button onClick={() => handleSearch(page - 1)} disabled={page === 1}>‹</button>
                    <span style={{ fontSize: 14, color: '#555' }}>Page {page} of {totalPages}</span>
                    <button onClick={() => handleSearch(page + 1)} disabled={page >= totalPages}>›</button>
                  </div>
                )}
              </>
            )}
          </div>
        )}
      </div>
    </>
  );
}
