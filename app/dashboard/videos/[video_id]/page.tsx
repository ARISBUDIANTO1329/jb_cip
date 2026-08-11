'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams, useSearchParams } from 'next/navigation';
import Image from 'next/image';
import { AppShell, PageContainer } from '@/components/layout';
import { getApiClient } from '@/lib/api/client';
import { Video } from '@/types';
import { formatNumber, formatDuration, formatDate } from '@/lib/utils/formatters';

export default function VideoDetailPage() {
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();
  const videoId = params.video_id as string;
  const channelId = searchParams.get('channel_id');

  const [video, setVideo] = useState<Video | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) { router.push('/login'); return; }
    if (!channelId) { return; }

    const api = getApiClient();
        const controller = new AbortController();

    async function load() {
      try {
        const res = await api.get<Video>(`/youtube/channels/${channelId}/videos/${videoId}`);
        if (res.success && res.data) {
          setVideo(res.data);
        } else {
          setError(res.error?.message || 'Video not found');
        }
      } catch (err) {
        if (err instanceof Error && err.name === 'UnauthorizedError') {
          router.push('/login');
          return;
        }
        setError(err instanceof Error ? err.message : 'Video not found');
      } finally {
        setLoading(false);
      }
    }

    load();
    return () => controller.abort();
  }, [videoId, channelId, router]);

  return (
    <AppShell>
      <PageContainer title="Video Detail">
        {loading && (
          <div style={{ textAlign: 'center', padding: '48px', color: '#888' }}>Loading video...</div>
        )}

        {!loading && (error || !video) && (
          <div style={{ textAlign: 'center', padding: '48px' }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>🎬</div>
            <p style={{ fontSize: '16px', color: '#ff6b6b', marginBottom: '8px' }}>{error || 'Video not found.'}</p>
            <button onClick={() => router.back()} style={{ padding: '8px 16px', backgroundColor: '#2a2a4a', color: '#e0e0e0', border: '1px solid #3a3a5a', borderRadius: '4px', cursor: 'pointer', marginTop: '16px' }}>
              Go Back
            </button>
          </div>
        )}

        {!loading && video && (
          <>
            <div style={{ display: 'flex', gap: '24px', marginBottom: '24px', flexWrap: 'wrap' }}>
              {video.thumbnail_url && (
                <Image
                  src={video.thumbnail_url}
                  alt={video.title}
                  width={320}
                  height={180}
                  style={{ objectFit: 'cover', borderRadius: '8px', maxWidth: '100%', height: 'auto' }}
                  unoptimized
                />
              )}
              <div style={{ flex: 1, minWidth: '280px' }}>
                <h2 style={{ fontSize: '20px', fontWeight: '600', color: '#e0e0e0', marginBottom: '8px' }}>{video.title}</h2>
                <p style={{ fontSize: '13px', color: '#666', marginBottom: '16px' }}>Video ID: {video.video_id}</p>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))', gap: '16px' }}>
                  <StatBox label="Views" value={formatNumber(video.view_count)} />
                  <StatBox label="Likes" value={formatNumber(video.like_count)} />
                  <StatBox label="Comments" value={formatNumber(video.comment_count)} />
                  <StatBox label="Duration" value={formatDuration(video.duration)} />
                  <StatBox label="Published" value={formatDate(video.published_at)} />
                  <StatBox label="Status" value={video.privacy_status} />
                </div>
              </div>
            </div>
            {video.description && (
              <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', padding: '20px' }}>
                <h3 style={{ fontSize: '14px', fontWeight: '600', color: '#888', marginBottom: '12px' }}>Description</h3>
                <p style={{ fontSize: '14px', color: '#ccc', lineHeight: '1.6', whiteSpace: 'pre-wrap' }}>{video.description}</p>
              </div>
            )}
          </>
        )}
      </PageContainer>
    </AppShell>
  );
}

function StatBox({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ backgroundColor: '#16213e', borderRadius: '6px', border: '1px solid #2a2a4a', padding: '12px' }}>
      <div style={{ fontSize: '11px', color: '#666', marginBottom: '4px', textTransform: 'uppercase' }}>{label}</div>
      <div style={{ fontSize: '16px', fontWeight: '600', color: '#e0e0e0' }}>{value}</div>
    </div>
  );
}
