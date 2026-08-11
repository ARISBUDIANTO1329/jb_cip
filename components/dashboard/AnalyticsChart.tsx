'use client';

import { AnalyticsTimeseriesPoint } from '@/types';

interface Props {
  data: AnalyticsTimeseriesPoint[];
  loading: boolean;
  metric: string;
}

const metricColors: Record<string, string> = {
  views: '#ff6b6b',
  watch_time: '#4ecdc4',
  likes: '#45b7d1',
  comments: '#96ceb4',
  shares: '#ffeaa7',
  subscribers: '#a29bfe',
};

const metricLabels: Record<string, string> = {
  views: 'Views',
  watch_time: 'Watch Time',
  likes: 'Likes',
  comments: 'Comments',
  shares: 'Shares',
  subscribers: 'Subscribers',
};

export function AnalyticsChart({ data, loading, metric }: Props) {
  const color = metricColors[metric] || '#ff6b6b';
  const label = metricLabels[metric] || metric;

  if (loading) {
    return (
      <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', padding: '48px', textAlign: 'center', color: '#888' }}>
        Loading chart...
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', padding: '48px', textAlign: 'center', color: '#888' }}>
        No timeseries data for this period.
      </div>
    );
  }

  const values = data.map((d) => d.value);
  const maxVal = Math.max(...values, 1);
  const minVal = 0;
  const range = maxVal - minVal || 1;

  const chartW = 100;
  const chartH = 40;
  const padX = 0;
  const padY = 2;

  const points = data.map((d, i) => {
    const x = data.length === 1 ? chartW / 2 : padX + (i / (data.length - 1)) * (chartW - 2 * padX);
    const y = padY + ((maxVal - d.value) / range) * (chartH - 2 * padY);
    return `${x},${y}`;
  });

  const areaPath = `M${padX},${chartH - padY} L${points.join(' L')} L${padX + (chartW - 2 * padX)},${chartH - padY} Z`;
  const linePath = `M${points.join(' L')}`;

  const yTicks = 4;
  const yLines = Array.from({ length: yTicks + 1 }, (_, i) => {
    const val = minVal + (range * i) / yTicks;
    const y = padY + ((maxVal - val) / range) * (chartH - 2 * padY);
    return { y, val: Math.round(val) };
  });

  return (
    <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', padding: '20px' }}>
      <div style={{ fontSize: '14px', fontWeight: '600', color: '#e0e0e0', marginBottom: '16px' }}>
        {label} Over Time
      </div>
      <div style={{ position: 'relative' }}>
        <svg viewBox={`0 0 ${chartW} ${chartH}`} preserveAspectRatio="none" style={{ width: '100%', height: '200px', display: 'block' }}>
          {yLines.map((tick, i) => (
            <g key={i}>
              <line x1={padX} y1={tick.y} x2={chartW - padX} y2={tick.y} stroke="#2a2a4a" strokeWidth="0.3" />
            </g>
          ))}
          <path d={areaPath} fill={color} opacity="0.15" />
          <path d={linePath} fill="none" stroke={color} strokeWidth="0.5" strokeLinecap="round" strokeLinejoin="round" />
          {data.map((d, i) => {
            const x = data.length === 1 ? chartW / 2 : padX + (i / (data.length - 1)) * (chartW - 2 * padX);
            const y = padY + ((maxVal - d.value) / range) * (chartH - 2 * padY);
            return <circle key={i} cx={x} cy={y} r="0.6" fill={color} />;
          })}
        </svg>
        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '8px', padding: '0 2px' }}>
          {data.map((d, i) => (
            <span key={i} style={{ fontSize: '10px', color: '#666' }}>
              {new Date(d.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
