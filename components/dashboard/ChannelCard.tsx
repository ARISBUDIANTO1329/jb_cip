'use client';

import { Channel } from '@/types';

interface ChannelCardProps {
  channel: Channel;
}

export function ChannelCard({ channel }: ChannelCardProps) {
  return (
    <div
      style={{
        backgroundColor: '#16213e',
        borderRadius: '8px',
        border: '1px solid #2a2a4a',
        padding: '20px',
        transition: 'border-color 0.15s ease',
      }}
      onMouseOver={(e) => (e.currentTarget.style.borderColor = '#3a3a5a')}
      onMouseOut={(e) => (e.currentTarget.style.borderColor = '#2a2a4a')}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '16px' }}>
        <div>
          <h3 style={{ fontSize: '16px', fontWeight: '600', color: '#e0e0e0', marginBottom: '4px' }}>
            {channel.name}
          </h3>
          <p style={{ fontSize: '12px', color: '#666' }}>{channel.external_id}</p>
        </div>
        <span
          style={{
            padding: '4px 8px',
            backgroundColor: channel.status === 'active' ? '#1a3d1a' : '#3d1a1a',
            color: channel.status === 'active' ? '#4ade80' : '#f87171',
            borderRadius: '4px',
            fontSize: '11px',
            fontWeight: '500',
            textTransform: 'uppercase',
          }}
        >
          {channel.status}
        </span>
      </div>

      {channel.description && (
        <p
          style={{
            fontSize: '13px',
            color: '#888',
            marginBottom: '16px',
            lineHeight: '1.5',
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
          }}
        >
          {channel.description}
        </p>
      )}

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px' }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '11px', color: '#666', marginBottom: '4px' }}>Subscribers</div>
          <div style={{ fontSize: '18px', fontWeight: '600', color: '#e0e0e0' }}>
            {channel.subscriber_count.toLocaleString()}
          </div>
        </div>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '11px', color: '#666', marginBottom: '4px' }}>Views</div>
          <div style={{ fontSize: '18px', fontWeight: '600', color: '#e0e0e0' }}>
            {channel.view_count.toLocaleString()}
          </div>
        </div>
        <div style={{ textAlign: 'center' }}>
          <div style={{ fontSize: '11px', color: '#666', marginBottom: '4px' }}>Videos</div>
          <div style={{ fontSize: '18px', fontWeight: '600', color: '#e0e0e0' }}>
            {channel.video_count.toLocaleString()}
          </div>
        </div>
      </div>
    </div>
  );
}
