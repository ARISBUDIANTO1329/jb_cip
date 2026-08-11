'use client';

interface Props {
  startDate: string;
  endDate: string;
  onStartChange: (date: string) => void;
  onEndChange: (date: string) => void;
  onApply: () => void;
}

export function AnalyticsDateRange({ startDate, endDate, onStartChange, onEndChange, onApply }: Props) {
  const inputStyle: React.CSSProperties = {
    padding: '6px 10px',
    backgroundColor: '#0f0f23',
    border: '1px solid #2a2a4a',
    borderRadius: '6px',
    color: '#e0e0e0',
    fontSize: '13px',
  };

  return (
    <div style={{ display: 'flex', gap: '12px', alignItems: 'center', flexWrap: 'wrap' }}>
      <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
        <span style={{ fontSize: '12px', color: '#888' }}>From</span>
        <input type="date" value={startDate} onChange={(e) => onStartChange(e.target.value)} style={inputStyle} />
      </div>
      <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
        <span style={{ fontSize: '12px', color: '#888' }}>To</span>
        <input type="date" value={endDate} onChange={(e) => onEndChange(e.target.value)} style={inputStyle} />
      </div>
      <button
        onClick={onApply}
        style={{
          padding: '6px 16px',
          backgroundColor: '#ff6b6b',
          color: 'white',
          border: 'none',
          borderRadius: '6px',
          cursor: 'pointer',
          fontSize: '13px',
          fontWeight: '500',
        }}
      >
        Apply
      </button>
    </div>
  );
}
