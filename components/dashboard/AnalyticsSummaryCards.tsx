'use client';

import { AnalyticsSummary } from '@/types';
import { formatNumber, formatWatchTime } from '@/lib/utils/formatters';

interface Props {
  summary: AnalyticsSummary | null;
  loading: boolean;
}

const metricKeys: Record<string, keyof AnalyticsSummary> = {
  views: 'views',
  watch_time: 'watch_time',
  likes: 'likes',
  comments: 'comments',
  shares: 'shares',
  subscribers: 'subscribers',
};

const cards = [
  { key: 'views', label: 'Views', icon: '👁️' },
  { key: 'watch_time', label: 'Watch Time', icon: '⏱️' },
  { key: 'likes', label: 'Likes', icon: '👍' },
  { key: 'comments', label: 'Comments', icon: '💬' },
  { key: 'shares', label: 'Shares', icon: '🔗' },
  { key: 'subscribers', label: 'Subscribers', icon: '👥' },
];

export function AnalyticsSummaryCards({ summary, loading }: Props) {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))', gap: '16px', marginBottom: '24px' }}>
      {cards.map((card) => {
        const raw = summary ? (summary[metricKeys[card.key]] as number) : 0;
        const value = loading ? '...' : card.key === 'watch_time' ? formatWatchTime(raw) : formatNumber(raw);
        return (
          <div
            key={card.key}
            style={{
              backgroundColor: '#16213e',
              borderRadius: '8px',
              border: '1px solid #2a2a4a',
              padding: '16px',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div>
                <div style={{ fontSize: '11px', color: '#888', marginBottom: '6px', textTransform: 'uppercase', letterSpacing: '0.5px' }}>{card.label}</div>
                <div style={{ fontSize: '20px', fontWeight: '600', color: '#e0e0e0' }}>{value}</div>
              </div>
              <div style={{ fontSize: '20px', opacity: 0.7 }}>{card.icon}</div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
