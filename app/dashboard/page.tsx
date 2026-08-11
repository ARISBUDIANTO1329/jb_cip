'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { AppShell, PageContainer } from '@/components/layout';
import { ChannelCard } from '@/components/dashboard/ChannelCard';
import { getApiClient } from '@/lib/api/client';
import { Channel, SyncStatusResponse, ApiResponse } from '@/types';

export default function DashboardPage() {
  const router = useRouter();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/login');
      return;
    }

    const fetchData = async () => {
      try {
        const api = getApiClient();
                const channelsRes: ApiResponse<Channel[]> = await api.get<Channel[]>(
          '/integrations/youtube/channels');

        if (channelsRes.success && channelsRes.data) {
          setChannels(channelsRes.data);

          if (channelsRes.data.length > 0) {
            const channelUuid = channelsRes.data[0].id;
            if (channelUuid) {
              const statusRes: ApiResponse<SyncStatusResponse> = await api.get<SyncStatusResponse>(
                `/youtube/sync/status?channel_id=${channelUuid}`);
              if (statusRes.success && statusRes.data) {
                setSyncStatus(statusRes.data);
              }
            }
          }
        } else {
          setError(channelsRes.error?.message || 'Failed to load data');
        }
      } catch (err) {
        if (err instanceof Error && err.name === 'UnauthorizedError') {
          router.push('/login');
          return;
        }
        setError(err instanceof Error ? err.message : 'Failed to load data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [router]);

  const totalSubscribers = channels.reduce((sum, ch) => sum + ch.subscriber_count, 0);
  const totalViews = channels.reduce((sum, ch) => sum + ch.view_count, 0);
  const totalVideos = channels.reduce((sum, ch) => sum + ch.video_count, 0);

  return (
    <AppShell>
      <PageContainer title="Overview">
        {loading && (
          <div style={{ textAlign: 'center', padding: '48px', color: '#888' }}>
            Loading dashboard...
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
              marginBottom: '24px',
            }}
          >
            {error}
          </div>
        )}

        {!loading && !error && (
          <>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', marginBottom: '32px' }}>
              <StatCard title="Channels" value={channels.length.toString()} icon="📺" />
              <StatCard title="Total Videos" value={totalVideos.toLocaleString()} icon="🎬" />
              <StatCard title="Subscribers" value={totalSubscribers.toLocaleString()} icon="👥" />
              <StatCard title="Total Views" value={totalViews.toLocaleString()} icon="👁️" />
              <StatCard
                title="Last Sync"
                value={syncStatus?.last_sync_at ? new Date(syncStatus.last_sync_at).toLocaleDateString() : 'Never'}
                icon="🔄"
              />
              <StatCard
                title="Sync Status"
                value={syncStatus?.last_sync_status || 'N/A'}
                icon={syncStatus?.last_sync_status === 'completed' ? '✅' : '⏳'}
              />
            </div>

            {channels.length > 0 && (
              <div>
                <h2 style={{ fontSize: '18px', fontWeight: '600', color: '#e0e0e0', marginBottom: '16px' }}>
                  Channels
                </h2>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))', gap: '20px' }}>
                  {channels.map((channel) => (
                    <ChannelCard key={channel.external_id} channel={channel} />
                  ))}
                </div>
              </div>
            )}

            {channels.length === 0 && (
              <div style={{ textAlign: 'center', padding: '48px', color: '#888' }}>
                <p style={{ fontSize: '16px', marginBottom: '8px' }}>No channels connected</p>
                <p style={{ fontSize: '14px' }}>Connect a YouTube channel to get started</p>
              </div>
            )}
          </>
        )}
      </PageContainer>
    </AppShell>
  );
}

function StatCard({ title, value, icon }: { title: string; value: string; icon: string }) {
  return (
    <div
      style={{
        backgroundColor: '#16213e',
        borderRadius: '8px',
        border: '1px solid #2a2a4a',
        padding: '16px',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <div style={{ fontSize: '11px', color: '#888', marginBottom: '6px', textTransform: 'uppercase', letterSpacing: '0.5px' }}>{title}</div>
          <div style={{ fontSize: '20px', fontWeight: '600', color: '#e0e0e0' }}>{value}</div>
        </div>
        <div style={{ fontSize: '20px', opacity: 0.7 }}>{icon}</div>
      </div>
    </div>
  );
}
