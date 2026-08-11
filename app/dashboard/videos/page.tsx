'use client';

import { useCallback, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { AppShell, PageContainer } from '@/components/layout';
import { VideoTable } from '@/components/dashboard/VideoTable';
import { getApiClient } from '@/lib/api/client';
import { Channel, Video } from '@/types';
import { formatNumber } from '@/lib/utils/formatters';

interface Pagination {
  limit: number;
  offset: number;
  total: number;
}

interface VideoListData {
  data: Video[];
  pagination: Pagination;
}

export default function VideosPage() {
  const router = useRouter();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedChannel, setSelectedChannel] = useState('');
  const [videos, setVideos] = useState<Video[]>([]);
  const [pagination, setPagination] = useState<Pagination>({ limit: 20, offset: 0, total: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchVideos = useCallback(async (channelId: string, offset: number) => {
    const api = getApiClient();
        try {
      const res = await api.get<VideoListData>(
        `/youtube/channels/${channelId}/videos?limit=20&offset=${offset}`);
      if (res.success && res.data) {
        setVideos(res.data.data || []);
        setPagination(res.data.pagination || { limit: 20, offset: 0, total: 0 });
        setError(null);
      } else {
        setError(res.error?.message || 'Failed to load videos');
      }
    } catch (err) {
      if (err instanceof Error && err.name === 'UnauthorizedError') {
        router.push('/login');
        return;
      }
      setError(err instanceof Error ? err.message : 'Failed to load videos');
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/login');
      return;
    }
    const api = getApiClient();
        const controller = new AbortController();

    async function loadChannels() {
      try {
        const res = await api.get<Channel[]>('/integrations/youtube/channels');
        if (res.success && res.data && res.data.length > 0) {
          setChannels(res.data);
          setSelectedChannel(res.data[0].id);
          void fetchVideos(res.data[0].id, 0);
        }
      } catch (err) {
        if (err instanceof Error && err.name === 'UnauthorizedError') {
          router.push('/login');
          return;
        }
        setError('Failed to load channels');
      }
    }

    void loadChannels();
    return () => controller.abort();
  }, [router, fetchVideos]);

  const handleChannelChange = useCallback((channelId: string) => {
    setSelectedChannel(channelId);
    setLoading(true);
    setError(null);
    void fetchVideos(channelId, 0);
  }, [fetchVideos]);

  const goToPage = useCallback((offset: number) => {
    if (!selectedChannel) return;
    setLoading(true);
    setError(null);
    void fetchVideos(selectedChannel, offset);
  }, [selectedChannel, fetchVideos]);

  const pageStart = pagination.offset + 1;
  const pageEnd = Math.min(pagination.offset + pagination.limit, pagination.total);
  const hasPrev = pagination.offset > 0;
  const hasNext = pagination.offset + pagination.limit < pagination.total;

  return (
    <AppShell>
      <PageContainer title="Videos">
        {channels.length > 1 && (
          <div style={{ marginBottom: '20px' }}>
            <select
              value={selectedChannel}
              onChange={(e) => handleChannelChange(e.target.value)}
              style={{
                padding: '8px 12px',
                backgroundColor: '#16213e',
                border: '1px solid #2a2a4a',
                borderRadius: '6px',
                color: '#e0e0e0',
                fontSize: '14px',
              }}
            >
              {channels.map((ch) => (
                <option key={ch.id} value={ch.id}>{ch.name}</option>
              ))}
            </select>
          </div>
        )}

        {error && (
          <div style={{ padding: '16px', backgroundColor: '#3d1f1f', border: '1px solid #5a2a2a', borderRadius: '8px', color: '#ff6b6b', fontSize: '14px', marginBottom: '20px' }}>
            {error}
          </div>
        )}

        <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', overflow: 'hidden' }}>
          <VideoTable videos={videos} channelId={selectedChannel} loading={loading} />

          {pagination.total > 0 && (
            <div style={{ padding: '12px 16px', borderTop: '1px solid #2a2a4a', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: '13px', color: '#888' }}>
                Showing {pageStart}–{pageEnd} of {formatNumber(pagination.total)}
              </span>
              <div style={{ display: 'flex', gap: '8px' }}>
                <button
                  onClick={() => goToPage(pagination.offset - pagination.limit)}
                  disabled={!hasPrev}
                  style={{
                    padding: '6px 16px',
                    backgroundColor: hasPrev ? '#2a2a4a' : '#1a1a2e',
                    color: hasPrev ? '#e0e0e0' : '#555',
                    border: '1px solid #3a3a5a',
                    borderRadius: '4px',
                    cursor: hasPrev ? 'pointer' : 'not-allowed',
                    fontSize: '13px',
                  }}
                >
                  Previous
                </button>
                <button
                  onClick={() => goToPage(pagination.offset + pagination.limit)}
                  disabled={!hasNext}
                  style={{
                    padding: '6px 16px',
                    backgroundColor: hasNext ? '#2a2a4a' : '#1a1a2e',
                    color: hasNext ? '#e0e0e0' : '#555',
                    border: '1px solid #3a3a5a',
                    borderRadius: '4px',
                    cursor: hasNext ? 'pointer' : 'not-allowed',
                    fontSize: '13px',
                  }}
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      </PageContainer>
    </AppShell>
  );
}
