'use client';
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Nav from '@/components/Nav';
import ProfilesTable from '@/components/ProfilesTable';
import { api, User, Profile, ProfilesResponse } from '@/lib/api';
import { withCSRF } from '@/lib/csrf';

export default function ProfilesPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState('');

  // Filters
  const [gender, setGender] = useState('');
  const [countryId, setCountryId] = useState('');
  const [ageGroup, setAgeGroup] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [order, setOrder] = useState('desc');

  const [exporting, setExporting] = useState(false);

  const handleExport = async () => {
    setExporting(true);
    try {
      const params: Record<string, string> = { format: 'csv' };
      if (gender)    params.gender = gender;
      if (countryId) params.country_id = countryId.toUpperCase();
      if (ageGroup)  params.age_group = ageGroup;

      const token = localStorage.getItem('access_token');
      const query = new URLSearchParams(params).toString();
      const url = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/profiles/export?${query}`;

      // Fetch with auth header then trigger browser download
      const res = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'X-API-Version': '1',
        },
      });

      if (!res.ok) throw new Error('Export failed');

      const blob = await res.blob();
      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = `profiles_${new Date().toISOString().slice(0, 10)}.csv`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(downloadUrl);
    } catch (err: any) {
      alert(err.message || 'Export failed');
    } finally {
      setExporting(false);
    }
  };

  const limit = 10;

  useEffect(() => {
    api.get('/auth/me').then((r) => setUser(r.data.data)).catch(() => router.push('/login'));
  }, [router]);

  useEffect(() => {
    fetchProfiles();
  }, [page, gender, countryId, ageGroup, sortBy, order]);

  const fetchProfiles = async () => {
    setLoading(true);
    try {
      const params: Record<string, string | number> = { page, limit, sort_by: sortBy, order };
      if (gender)    params.gender = gender;
      if (countryId) params.country_id = countryId.toUpperCase();
      if (ageGroup)  params.age_group = ageGroup;

      const res = await api.get<ProfilesResponse>('/api/profiles', { params });
      setProfiles(res.data.data);
      setTotal(res.data.total);
      setTotalPages(res.data.total_pages);
    } catch {
      router.push('/login');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!newName.trim()) return;
    setCreating(true);
    try {
      await api.post('/api/profiles', { name: newName }, { headers: withCSRF() });
      setNewName('');
      fetchProfiles();
    } catch (err: any) {
      alert(err.response?.data?.message || 'Failed to create profile');
    } finally {
      setCreating(false);
    }
  };

  if (!user) return <div className="loading">Loading...</div>;

  return (
    <>
      <Nav username={user.username} role={user.role} />
      <div className="container page">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
          <h1 className="page-title" style={{ margin: 0 }}>Profiles</h1>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span style={{ color: '#888', fontSize: 14 }}>{total.toLocaleString()} total</span>
            <button
              className="btn btn-ghost"
              onClick={handleExport}
              disabled={exporting}
              style={{ fontSize: 13 }}
            >
              {exporting ? 'Exporting...' : '⬇ Export CSV'}
            </button>
          </div>
        </div>

        {/* Filters */}
        <div className="filters">
          <select value={gender} onChange={(e) => { setGender(e.target.value); setPage(1); }}>
            <option value="">All genders</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
          </select>
          <input
            placeholder="Country code (e.g. NG)"
            value={countryId}
            onChange={(e) => { setCountryId(e.target.value); setPage(1); }}
            style={{ width: 160 }}
          />
          <select value={ageGroup} onChange={(e) => { setAgeGroup(e.target.value); setPage(1); }}>
            <option value="">All age groups</option>
            <option value="child">Child</option>
            <option value="teenager">Teenager</option>
            <option value="adult">Adult</option>
            <option value="senior">Senior</option>
          </select>
          <select value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
            <option value="created_at">Sort: Date</option>
            <option value="age">Sort: Age</option>
            <option value="gender_probability">Sort: Gender Prob</option>
          </select>
          <select value={order} onChange={(e) => setOrder(e.target.value)}>
            <option value="desc">Desc</option>
            <option value="asc">Asc</option>
          </select>
        </div>

        {/* Create profile (admin only) */}
        {user.role === 'admin' && (
          <div className="card" style={{ marginBottom: 24, display: 'flex', gap: 12, alignItems: 'center' }}>
            <input
              placeholder="Full name to create profile..."
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              style={{ flex: 1, padding: '8px 12px', border: '1px solid #ddd', borderRadius: 8, fontSize: 14 }}
              onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            />
            <button className="btn btn-primary" onClick={handleCreate} disabled={creating}>
              {creating ? 'Creating...' : '+ Create Profile'}
            </button>
          </div>
        )}

        {/* Table */}
        <div className="card">
          {loading ? (
            <div className="loading">Loading profiles...</div>
          ) : (
            <ProfilesTable profiles={profiles} onRowClick={(id) => router.push(`/profiles/${id}`)} />
          )}
        </div>

        {/* Pagination */}
        <div className="pagination">
          <button onClick={() => setPage(1)} disabled={page === 1}>«</button>
          <button onClick={() => setPage(p => p - 1)} disabled={page === 1}>‹</button>
          <span style={{ fontSize: 14, color: '#555' }}>Page {page} of {totalPages}</span>
          <button onClick={() => setPage(p => p + 1)} disabled={page >= totalPages}>›</button>
          <button onClick={() => setPage(totalPages)} disabled={page >= totalPages}>»</button>
        </div>
      </div>
    </>
  );
}
