'use client';

import Link from 'next/link';
import Image from 'next/image';
import { Video } from '@/types';
import { formatNumber, formatDuration, formatDate } from '@/lib/utils/formatters';

interface VideoTableProps {
  videos: Video[];
  channelId: string;
  loading?: boolean;
}

export function VideoTable({ videos, channelId, loading }: VideoTableProps) {
  if (loading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', color: '#888' }}>
        Loading videos...
      </div>
    );
  }

  if (videos.length === 0) {
    return (
      <div style={{ padding: '32px', textAlign: 'center', color: '#888' }}>
        No videos found.
      </div>
    );
  }

  return (
    <div style={{ overflowX: 'auto' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr style={{ borderBottom: '1px solid #2a2a4a' }}>
            <th style={{ textAlign: 'left', padding: '12px 16px', fontSize: '12px', color: '#888', fontWeight: '500', width: '60px' }}></th>
            <th style={{ textAlign: 'left', padding: '12px 16px', fontSize: '12px', color: '#888', fontWeight: '500' }}>Title</th>
            <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '12px', color: '#888', fontWeight: '500' }}>Views</th>
            <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '12px', color: '#888', fontWeight: '500' }}>Likes</th>
            <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '12px', color: '#888', fontWeight: '500' }}>Comments</th>
            <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '12px', color: '#888', fontWeight: '500' }}>Duration</th>
            <th style={{ textAlign: 'right', padding: '12px 16px', fontSize: '12px', color: '#888', fontWeight: '500' }}>Published</th>
          </tr>
        </thead>
        <tbody>
          {videos.map((video) => (
            <tr key={video.id} style={{ borderBottom: '1px solid #1a1a2e' }}>
              <td style={{ padding: '8px 16px' }}>
                {video.thumbnail_url ? (
                  <Image
                    src={video.thumbnail_url}
                    alt=""
                    width={48}
                    height={27}
                    style={{ objectFit: 'cover', borderRadius: '3px', display: 'block' }}
                    unoptimized
                  />
                ) : (
                  <div style={{ width: '48px', height: '27px', backgroundColor: '#2a2a4a', borderRadius: '3px' }} />
                )}
              </td>
              <td style={{ padding: '12px 16px' }}>
                <Link
                  href={`/dashboard/videos/${video.video_id}?channel_id=${channelId}`}
                  style={{ textDecoration: 'none' }}
                >
                  <div style={{ fontSize: '14px', color: '#e0e0e0', maxWidth: '400px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {video.title}
                  </div>
                  <div style={{ fontSize: '11px', color: '#666', marginTop: '2px' }}>{video.video_id}</div>
                </Link>
              </td>
              <td style={{ textAlign: 'right', padding: '12px 16px', fontSize: '14px', color: '#e0e0e0' }}>
                {formatNumber(video.view_count)}
              </td>
              <td style={{ textAlign: 'right', padding: '12px 16px', fontSize: '14px', color: '#e0e0e0' }}>
                {formatNumber(video.like_count)}
              </td>
              <td style={{ textAlign: 'right', padding: '12px 16px', fontSize: '14px', color: '#e0e0e0' }}>
                {formatNumber(video.comment_count)}
              </td>
              <td style={{ textAlign: 'right', padding: '12px 16px', fontSize: '13px', color: '#888' }}>
                {formatDuration(video.duration)}
              </td>
              <td style={{ textAlign: 'right', padding: '12px 16px', fontSize: '13px', color: '#888' }}>
                {formatDate(video.published_at)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
