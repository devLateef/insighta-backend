'use client';
import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Suspense } from 'react';

function CallbackHandler() {
  const router = useRouter();
  const params = useSearchParams();

  useEffect(() => {
    const accessToken  = params.get('access_token');
    const refreshToken = params.get('refresh_token');
    const username     = params.get('username');
    const role         = params.get('role');
    const id           = params.get('id');

    if (!accessToken || !refreshToken) {
      router.replace('/login');
      return;
    }

    // Store tokens in localStorage for API calls
    localStorage.setItem('access_token',  accessToken);
    localStorage.setItem('refresh_token', refreshToken);
    localStorage.setItem('username',      username || '');
    localStorage.setItem('role',          role || '');
    localStorage.setItem('user_id',       id || '');
    // Store expiry times (access=3min, refresh=5min from now)
    localStorage.setItem('access_expires',  String(Date.now() + 3 * 60 * 1000));
    localStorage.setItem('refresh_expires', String(Date.now() + 5 * 60 * 1000));

    // Clean URL and go to dashboard
    router.replace('/dashboard');
  }, [params, router]);

  return (
    <div style={{
      minHeight: '100vh', display: 'flex', alignItems: 'center',
      justifyContent: 'center', fontFamily: 'sans-serif', color: '#555'
    }}>
      <p>Completing login...</p>
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense fallback={<div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>Loading...</div>}>
      <CallbackHandler />
    </Suspense>
  );
}
