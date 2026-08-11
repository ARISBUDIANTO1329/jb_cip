'use client';

const metrics = [
  { value: 'views', label: 'Views' },
  { value: 'watch_time', label: 'Watch Time' },
  { value: 'likes', label: 'Likes' },
  { value: 'comments', label: 'Comments' },
  { value: 'shares', label: 'Shares' },
  { value: 'subscribers', label: 'Subscribers' },
];

interface Props {
  value: string;
  onChange: (metric: string) => void;
}

export function AnalyticsMetricSelector({ value, onChange }: Props) {
  return (
    <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
      {metrics.map((m) => (
        <button
          key={m.value}
          onClick={() => onChange(m.value)}
          style={{
            padding: '6px 14px',
            backgroundColor: value === m.value ? '#ff6b6b' : '#2a2a4a',
            color: value === m.value ? 'white' : '#888',
            border: value === m.value ? '1px solid #ff6b6b' : '1px solid #3a3a5a',
            borderRadius: '20px',
            cursor: 'pointer',
            fontSize: '12px',
            fontWeight: '500',
            transition: 'all 0.15s ease',
          }}
        >
          {m.label}
        </button>
      ))}
    </div>
  );
}
