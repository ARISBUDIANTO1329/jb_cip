'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/lib/auth/context';

const navItems = [
  { href: '/dashboard', label: 'Overview', icon: '📊' },
  { href: '/dashboard/channels', label: 'Channels', icon: '📺' },
  { href: '/dashboard/videos', label: 'Videos', icon: '🎬' },
  { href: '/dashboard/analytics', label: 'Analytics', icon: '📈' },
  { href: '/dashboard/sync', label: 'Sync', icon: '🔄' },
];

export function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);

  return (
    <aside
      style={{
        width: collapsed ? '60px' : '240px',
        minHeight: '100vh',
        backgroundColor: '#1a1a2e',
        color: '#e0e0e0',
        transition: 'width 0.2s ease',
        display: 'flex',
        flexDirection: 'column',
        borderRight: '1px solid #2a2a4a',
      }}
    >
      <div
        style={{
          padding: '20px',
          borderBottom: '1px solid #2a2a4a',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        {!collapsed && (
          <span style={{ fontWeight: 'bold', fontSize: '18px', color: '#ff6b6b' }}>
            CIP
          </span>
        )}
        <button
          onClick={() => setCollapsed(!collapsed)}
          style={{
            background: 'none',
            border: 'none',
            color: '#888',
            cursor: 'pointer',
            fontSize: '16px',
          }}
        >
          {collapsed ? '→' : '←'}
        </button>
      </div>

      <nav style={{ flex: 1, padding: '10px 0' }}>
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            style={{
              display: 'flex',
              alignItems: 'center',
              padding: '12px 20px',
              color: pathname === item.href ? '#ff6b6b' : '#888',
              textDecoration: 'none',
              backgroundColor: pathname === item.href ? '#2a2a4a' : 'transparent',
              transition: 'background-color 0.15s ease',
            }}
          >
            <span style={{ marginRight: collapsed ? '0' : '12px', fontSize: '18px' }}>
              {item.icon}
            </span>
            {!collapsed && <span>{item.label}</span>}
          </Link>
        ))}
      </nav>
    </aside>
  );
}

export function Topbar() {
  const { user, logout } = useAuth();

  return (
    <header
      style={{
        height: '60px',
        backgroundColor: '#16213e',
        borderBottom: '1px solid #2a2a4a',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 24px',
        color: '#e0e0e0',
      }}
    >
      <div style={{ fontSize: '14px', color: '#888' }}>
        Creator Intelligence Platform
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        {user && (
          <>
            <span style={{ fontSize: '14px' }}>{user.name || user.email}</span>
            <button
              onClick={logout}
              style={{
                padding: '6px 12px',
                backgroundColor: 'transparent',
                border: '1px solid #444',
                color: '#888',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '12px',
              }}
            >
              Logout
            </button>
          </>
        )}
      </div>
    </header>
  );
}

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', minHeight: '100vh', backgroundColor: '#0f0f23' }}>
      <Sidebar />
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <Topbar />
        <main style={{ flex: 1, padding: '24px', overflowY: 'auto' }}>
          {children}
        </main>
      </div>
    </div>
  );
}

export function PageContainer({ children, title }: { children: React.ReactNode; title?: string }) {
  return (
    <div>
      {title && (
        <h1 style={{ fontSize: '24px', fontWeight: '600', color: '#e0e0e0', marginBottom: '24px' }}>
          {title}
        </h1>
      )}
      {children}
    </div>
  );
}
