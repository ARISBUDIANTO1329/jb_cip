'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      router.push('/dashboard');
    }
  }, [router]);

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: '#0f0f23',
        color: '#e0e0e0',
      }}
    >
      <div style={{ textAlign: 'center' }}>
        <h1 style={{ fontSize: '48px', fontWeight: 'bold', marginBottom: '16px', color: '#ff6b6b' }}>
          CIP
        </h1>
        <h2 style={{ fontSize: '24px', fontWeight: '300', marginBottom: '8px', color: '#888' }}>
          Creator Intelligence Platform
        </h2>
        <p style={{ fontSize: '16px', color: '#666', marginBottom: '32px' }}>
          YouTube Channel Intelligence
        </p>
        <button
          onClick={() => router.push('/login')}
          style={{
            padding: '12px 32px',
            backgroundColor: '#ff6b6b',
            color: 'white',
            border: 'none',
            borderRadius: '8px',
            fontSize: '16px',
            fontWeight: '600',
            cursor: 'pointer',
            transition: 'background-color 0.15s ease',
          }}
          onMouseOver={(e) => (e.currentTarget.style.backgroundColor = '#e55a5a')}
          onMouseOut={(e) => (e.currentTarget.style.backgroundColor = '#ff6b6b')}
        >
          Login
        </button>
      </div>
    </div>
  );
}
