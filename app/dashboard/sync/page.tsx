'use client';

import { AppShell, PageContainer } from '@/components/layout';

export default function SyncPage() {
  return (
    <AppShell>
      <PageContainer title="Sync">
        <div
          style={{
            backgroundColor: '#16213e',
            borderRadius: '8px',
            border: '1px solid #2a2a4a',
            padding: '48px',
            textAlign: 'center',
          }}
        >
          <div style={{ fontSize: '48px', marginBottom: '16px' }}>🔄</div>
          <h2 style={{ fontSize: '20px', fontWeight: '600', color: '#e0e0e0', marginBottom: '8px' }}>
            Sync Management
          </h2>
          <p style={{ fontSize: '14px', color: '#888' }}>
            Sync management features will be available in V4.
          </p>
        </div>
      </PageContainer>
    </AppShell>
  );
}
