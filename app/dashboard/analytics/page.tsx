'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { AppShell, PageContainer } from '@/components/layout';
import { AnalyticsSummaryCards } from '@/components/dashboard/AnalyticsSummaryCards';
import { AnalyticsChart } from '@/components/dashboard/AnalyticsChart';
import { AnalyticsMetricSelector } from '@/components/dashboard/AnalyticsMetricSelector';
import { AnalyticsDateRange } from '@/components/dashboard/AnalyticsDateRange';
import { TopVideos } from '@/components/dashboard/TopVideos';
import { getApiClient } from '@/lib/api/client';
import { Channel, AnalyticsSummary, AnalyticsTimeseriesPoint, TopVideoAnalytics } from '@/types';

export default function AnalyticsPage() {
  const router = useRouter();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedChannel, setSelectedChannel] = useState('');
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [timeseries, setTimeseries] = useState<AnalyticsTimeseriesPoint[]>([]);
  const [topVideos, setTopVideos] = useState<TopVideoAnalytics[]>([]);
  const [metric, setMetric] = useState('views');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [loadingSummary, setLoadingSummary] = useState(true);
  const [loadingChart, setLoadingChart] = useState(true);
  const [loadingTop, setLoadingTop] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [noChannel, setNoChannel] = useState(false);
  const initialized = useRef(false);

  useEffect(() => {
    if (initialized.current) return;
    initialized.current = true;

    const token = localStorage.getItem('access_token');
    if (!token) { router.push('/login'); return; }

    const api = getApiClient();
        async function init() {
      try {
        const res = await api.get<Channel[]>('/integrations/youtube/channels');
        if (res.success && res.data && res.data.length > 0) {
          setChannels(res.data);
          const ch = res.data[0].id;
          setSelectedChannel(ch);
          return ch;
        }
        setNoChannel(true);
        setLoadingSummary(false);
        setLoadingChart(false);
        setLoadingTop(false);
        return null;
      } catch (err) {
        if (err instanceof Error && err.name === 'UnauthorizedError') { router.push('/login'); return null; }
        setError('Failed to load channels');
        setLoadingSummary(false);
        setLoadingChart(false);
        setLoadingTop(false);
        return null;
      }
    }

    async function initAndFetch() {
      const ch = await init();
      if (!ch) return;
      const qs = new URLSearchParams({ channel_id: ch });
      const [s, t, v] = await Promise.allSettled([
        api.get<AnalyticsSummary>(`/youtube/analytics/summary?${qs}`),
        api.get<AnalyticsTimeseriesPoint[]>(`/youtube/analytics/timeseries?${qs.toString()}&metric=views`),
        api.get<TopVideoAnalytics[]>(`/youtube/analytics/top-videos?${qs.toString()}&limit=10`),
      ]);
      if (s.status === 'fulfilled' && s.value.success && s.value.data) setSummary(s.value.data);
      if (t.status === 'fulfilled' && t.value.success && t.value.data) setTimeseries(t.value.data as unknown as AnalyticsTimeseriesPoint[]);
      if (v.status === 'fulfilled' && v.value.success && v.value.data) setTopVideos(v.value.data as unknown as TopVideoAnalytics[]);
      setLoadingSummary(false);
      setLoadingChart(false);
      setLoadingTop(false);
    }

    void initAndFetch();
  }, [router]);

  const fetchMetric = useCallback(async (channelId: string, m: string, start: string, end: string) => {
    const api = getApiClient();
        const qs = new URLSearchParams({ channel_id: channelId, metric: m });
    if (start) qs.set('start_date', start);
    if (end) qs.set('end_date', end);

    setLoadingChart(true);
    try {
      const res = await api.get<AnalyticsTimeseriesPoint[]>(`/youtube/analytics/timeseries?${qs}`);
      if (res.success && res.data) setTimeseries(res.data as unknown as AnalyticsTimeseriesPoint[]);
    } catch { /* ignore */ }
    finally { setLoadingChart(false); }
  }, []);

  const handleMetricChange = useCallback((m: string) => {
    setMetric(m);
    if (selectedChannel) void fetchMetric(selectedChannel, m, startDate, endDate);
  }, [selectedChannel, startDate, endDate, fetchMetric]);

  const handleApply = useCallback(() => {
    if (!selectedChannel) return;
    const api = getApiClient();
        const qs = new URLSearchParams({ channel_id: selectedChannel });
    if (startDate) qs.set('start_date', startDate);
    if (endDate) qs.set('end_date', endDate);

    setLoadingSummary(true);
    setLoadingChart(true);
    setLoadingTop(true);

    void Promise.allSettled([
      api.get<AnalyticsSummary>(`/youtube/analytics/summary?${qs}`),
      api.get<AnalyticsTimeseriesPoint[]>(`/youtube/analytics/timeseries?${qs}&metric=${metric}`),
      api.get<TopVideoAnalytics[]>(`/youtube/analytics/top-videos?${qs}&limit=10`),
    ]).then(([s, t, v]) => {
      if (s.status === 'fulfilled' && s.value.success && s.value.data) setSummary(s.value.data);
      if (t.status === 'fulfilled' && t.value.success && t.value.data) setTimeseries(t.value.data as unknown as AnalyticsTimeseriesPoint[]);
      if (v.status === 'fulfilled' && v.value.success && v.value.data) setTopVideos(v.value.data as unknown as TopVideoAnalytics[]);
      setLoadingSummary(false);
      setLoadingChart(false);
      setLoadingTop(false);
    });
  }, [selectedChannel, startDate, endDate, metric]);

  const handleChannelChange = useCallback((ch: string) => {
    setSelectedChannel(ch);
    const api = getApiClient();
        const qs = new URLSearchParams({ channel_id: ch });
    if (startDate) qs.set('start_date', startDate);
    if (endDate) qs.set('end_date', endDate);

    setLoadingSummary(true);
    setLoadingChart(true);
    setLoadingTop(true);

    void Promise.allSettled([
      api.get<AnalyticsSummary>(`/youtube/analytics/summary?${qs}`),
      api.get<AnalyticsTimeseriesPoint[]>(`/youtube/analytics/timeseries?${qs}&metric=${metric}`),
      api.get<TopVideoAnalytics[]>(`/youtube/analytics/top-videos?${qs}&limit=10`),
    ]).then(([s, t, v]) => {
      if (s.status === 'fulfilled' && s.value.success && s.value.data) setSummary(s.value.data);
      if (t.status === 'fulfilled' && t.value.success && t.value.data) setTimeseries(t.value.data as unknown as AnalyticsTimeseriesPoint[]);
      if (v.status === 'fulfilled' && v.value.success && v.value.data) setTopVideos(v.value.data as unknown as TopVideoAnalytics[]);
      setLoadingSummary(false);
      setLoadingChart(false);
      setLoadingTop(false);
    });
  }, [startDate, endDate, metric]);

  return (
    <AppShell>
      <PageContainer title="Analytics">
        {error && (
          <div style={{ padding: '16px', backgroundColor: '#3d1f1f', border: '1px solid #5a2a2a', borderRadius: '8px', color: '#ff6b6b', fontSize: '14px', marginBottom: '20px' }}>
            {error}
          </div>
        )}

        {noChannel && !error && (
          <div style={{ textAlign: 'center', padding: '48px', color: '#888' }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>📺</div>
            <p style={{ fontSize: '16px', marginBottom: '8px' }}>No YouTube channel connected</p>
            <p style={{ fontSize: '14px' }}>Connect a channel via Google OAuth to view analytics.</p>
          </div>
        )}

        {!noChannel && (
          <>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', flexWrap: 'wrap', gap: '12px' }}>
              {channels.length > 1 ? (
                <select
                  value={selectedChannel}
                  onChange={(e) => handleChannelChange(e.target.value)}
                  style={{ padding: '8px 12px', backgroundColor: '#16213e', border: '1px solid #2a2a4a', borderRadius: '6px', color: '#e0e0e0', fontSize: '14px' }}
                >
                  {channels.map((ch) => (<option key={ch.id} value={ch.id}>{ch.name}</option>))}
                </select>
              ) : channels.length === 1 ? (
                <div style={{ fontSize: '14px', color: '#888' }}>{channels[0].name}</div>
              ) : null}
              <AnalyticsDateRange startDate={startDate} endDate={endDate} onStartChange={setStartDate} onEndChange={setEndDate} onApply={handleApply} />
            </div>

            <AnalyticsSummaryCards summary={summary} loading={loadingSummary} />

            <div style={{ marginBottom: '24px' }}>
              <AnalyticsMetricSelector value={metric} onChange={handleMetricChange} />
            </div>

            <div style={{ marginBottom: '24px' }}>
              <AnalyticsChart data={timeseries} loading={loadingChart} metric={metric} />
            </div>

            <TopVideos videos={topVideos} channelId={selectedChannel} loading={loadingTop} />
          </>
        )}
      </PageContainer>
    </AppShell>
  );
}
