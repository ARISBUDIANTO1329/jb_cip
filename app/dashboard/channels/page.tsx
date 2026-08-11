'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { AppShell, PageContainer } from '@/components/layout';
import { ChannelCard } from '@/components/dashboard/ChannelCard';
import { getApiClient } from '@/lib/api/client';
import { Channel, ApiResponse } from '@/types';

export default function ChannelsPage() {
  const router = useRouter();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/login');
      return;
    }

    const fetchChannels = async () => {
      try {
        const api = getApiClient();
                const res: ApiResponse<Channel[]> = await api.get<Channel[]>(
          '/integrations/youtube/channels');

        if (res.success && res.data) {
          setChannels(res.data);
        } else {
          setError(res.error?.message || 'Failed to load channels');
        }
      } catch (err) {
        if (err instanceof Error && err.name === 'UnauthorizedError') {
          router.push('/login');
          return;
        }
        setError(err instanceof Error ? err.message : 'Failed to load channels');
      } finally {
        setLoading(false);
      }
    };

    fetchChannels();
  }, [router]);

  return (
    <AppShell>
      <PageContainer title="Channels">
        {loading && (
          <div style={{ textAlign: 'center', padding: '48px', color: '#888' }}>
            Loading channels...
          </div>
        )}

        {error && (
          <div
            style={{
              padding: '16px',
              backgroundColor: '#3d1f1f',
              border: '1px solid #5a2a2a',
              borderRadius: '8px',
              color: '#ff6b6b',
              fontSize: '14px',
            }}
          >
            {error}
          </div>
        )}

        {!loading && !error && channels.length === 0 && (
          <div style={{ textAlign: 'center', padding: '48px', color: '#888' }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>📺</div>
            <p style={{ fontSize: '16px', marginBottom: '8px' }}>No channels connected</p>
            <p style={{ fontSize: '14px' }}>Connect a YouTube channel via Google OAuth to get started.</p>
          </div>
        )}

        {!loading && !error && channels.length > 0 && (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))', gap: '20px' }}>
            {channels.map((channel) => (
              <ChannelCard key={channel.external_id} channel={channel} />
            ))}
          </div>
        )}
      </PageContainer>
    </AppShell>
  );
}
