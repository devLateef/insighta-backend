'use client';
import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import Nav from '@/components/Nav';
import { api, User, Profile } from '@/lib/api';
import { withCSRF } from '@/lib/csrf';

export default function ProfileDetailPage() {
  const router = useRouter();
  const { id } = useParams<{ id: string }>();
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    const load = async () => {
      try {
        const [meRes, profileRes] = await Promise.all([
          api.get('/auth/me'),
          api.get(`/api/profiles/${id}`),
        ]);
        setUser(meRes.data.data);
        setProfile(profileRes.data.data);
      } catch {
        router.push('/login');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [id, router]);

  const handleDelete = async () => {
    if (!confirm('Delete this profile?')) return;
    setDeleting(true);
    try {
      await api.delete(`/api/profiles/${id}`, { headers: withCSRF() });
      router.push('/profiles');
    } catch (err: any) {
      alert(err.response?.data?.message || 'Failed to delete');
      setDeleting(false);
    }
  };

  if (loading || !user) return <div className="loading">Loading...</div>;
  if (!profile) return <div className="loading">Profile not found.</div>;

  return (
    <>
      <Nav username={user.username} role={user.role} />
      <div className="container page">
        <button className="btn btn-ghost" style={{ marginBottom: 24 }} onClick={() => router.back()}>
          ← Back
        </button>

        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <h1 style={{ fontSize: 28, fontWeight: 700 }}>{profile.name}</h1>
            {user.role === 'admin' && (
              <button className="btn btn-danger" onClick={handleDelete} disabled={deleting}>
                {deleting ? 'Deleting...' : 'Delete'}
              </button>
            )}
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24, marginTop: 24 }}>
            <div>
              <p style={{ color: '#888', fontSize: 13, marginBottom: 4 }}>Gender</p>
              <p>
                <span className={`badge badge-${profile.gender}`}>{profile.gender}</span>
                &nbsp;
                <span style={{ color: '#888', fontSize: 13 }}>({(profile.gender_probability * 100).toFixed(1)}% confidence)</span>
              </p>
            </div>
            <div>
              <p style={{ color: '#888', fontSize: 13, marginBottom: 4 }}>Age</p>
              <p style={{ fontWeight: 600 }}>{profile.age} <span style={{ color: '#888', fontWeight: 400, fontSize: 13 }}>({profile.age_group})</span></p>
            </div>
            <div>
              <p style={{ color: '#888', fontSize: 13, marginBottom: 4 }}>Country</p>
              <p style={{ fontWeight: 600 }}>{profile.country_name} ({profile.country_id}) <span style={{ color: '#888', fontWeight: 400, fontSize: 13 }}>({(profile.country_probability * 100).toFixed(1)}% confidence)</span></p>
            </div>
            <div>
              <p style={{ color: '#888', fontSize: 13, marginBottom: 4 }}>Created</p>
              <p>{new Date(profile.created_at).toLocaleString()}</p>
            </div>
            <div style={{ gridColumn: '1 / -1' }}>
              <p style={{ color: '#888', fontSize: 13, marginBottom: 4 }}>ID</p>
              <p style={{ fontFamily: 'monospace', fontSize: 13 }}>{profile.id}</p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
