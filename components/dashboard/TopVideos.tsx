'use client';

import Link from 'next/link';
import Image from 'next/image';
import { TopVideoAnalytics } from '@/types';
import { formatNumber, formatWatchTime } from '@/lib/utils/formatters';

interface Props {
  videos: TopVideoAnalytics[];
  channelId: string;
  loading: boolean;
}

export function TopVideos({ videos, channelId, loading }: Props) {
  if (loading) {
    return (
      <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', padding: '24px', textAlign: 'center', color: '#888' }}>
        Loading top videos...
      </div>
    );
  }

  if (!videos || videos.length === 0) {
    return (
      <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', padding: '48px', textAlign: 'center', color: '#888' }}>
        <div style={{ fontSize: '32px', marginBottom: '12px' }}>🎬</div>
        <p style={{ fontSize: '14px' }}>No video analytics available.</p>
      </div>
    );
  }

  return (
    <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', overflow: 'hidden' }}>
      <div style={{ padding: '16px 20px', borderBottom: '1px solid #2a2a4a' }}>
        <h3 style={{ fontSize: '16px', fontWeight: '600', color: '#e0e0e0' }}>Top Videos</h3>
      </div>
      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid #2a2a4a' }}>
              <th style={{ textAlign: 'center', padding: '12px 12px', fontSize: '11px', color: '#666', fontWeight: '500', width: '30px' }}>#</th>
              <th style={{ textAlign: 'left', padding: '12px 12px', fontSize: '11px', color: '#666', fontWeight: '500', width: '50px' }}></th>
              <th style={{ textAlign: 'left', padding: '12px 16px', fontSize: '11px', color: '#666', fontWeight: '500' }}>Title</th>
              <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '11px', color: '#666', fontWeight: '500' }}>Views</th>
              <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '11px', color: '#666', fontWeight: '500' }}>Likes</th>
              <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '11px', color: '#666', fontWeight: '500' }}>Comments</th>
              <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '11px', color: '#666', fontWeight: '500' }}>Watch Time</th>
            </tr>
          </thead>
          <tbody>
            {videos.map((video, i) => (
              <tr key={video.internal_video_id} style={{ borderBottom: '1px solid #1a1a2e' }}>
                <td style={{ textAlign: 'center', padding: '10px 12px', fontSize: '13px', color: i < 3 ? '#ff6b6b' : '#666', fontWeight: i < 3 ? '700' : '400' }}>
                  {i + 1}
                </td>
                <td style={{ padding: '8px 12px' }}>
                  {video.thumbnail_url ? (
                    <Image src={video.thumbnail_url} alt="" width={48} height={27} style={{ objectFit: 'cover', borderRadius: '3px', display: 'block' }} unoptimized />
                  ) : (
                    <div style={{ width: '48px', height: '27px', backgroundColor: '#2a2a4a', borderRadius: '3px' }} />
                  )}
                </td>
                <td style={{ padding: '10px 16px' }}>
                  <Link href={`/dashboard/videos/${video.video_id}?channel_id=${channelId}`} style={{ textDecoration: 'none' }}>
                    <div style={{ fontSize: '13px', color: '#e0e0e0', maxWidth: '360px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {video.title}
                    </div>
                  </Link>
                </td>
                <td style={{ textAlign: 'right', padding: '10px 16px', fontSize: '13px', color: '#e0e0e0' }}>{formatNumber(video.views)}</td>
                <td style={{ textAlign: 'right', padding: '10px 16px', fontSize: '13px', color: '#e0e0e0' }}>{formatNumber(video.likes)}</td>
                <td style={{ textAlign: 'right', padding: '10px 16px', fontSize: '13px', color: '#e0e0e0' }}>{formatNumber(video.comments)}</td>
                <td style={{ textAlign: 'right', padding: '10px 16px', fontSize: '13px', color: '#888' }}>{formatWatchTime(video.watch_time)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
