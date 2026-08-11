'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { AppShell, PageContainer } from '@/components/layout';
import { getApiClient } from '@/lib/api/client';
import {
  Channel,
  SyncStatusResponse,
  SyncResponse,
  AnalyticsSyncResponse,
  SyncJob,
  SyncHistoryData,
} from '@/types';
import { formatNumber } from '@/lib/utils/formatters';

type SyncType = 'manual' | 'incremental';

export default function SyncPage() {
  const router = useRouter();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedChannel, setSelectedChannel] = useState('');
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null);
  const [history, setHistory] = useState<SyncJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [videoBusy, setVideoBusy] = useState(false);
  const [analyticsBusy, setAnalyticsBusy] = useState(false);
  const [feedback, setFeedback] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const loadStatus = useCallback(async (channelId: string) => {
    const api = getApiClient();
    try {
      const res = await api.get<SyncStatusResponse>(`/youtube/sync/status?channel_id=${channelId}`);
      if (res.success && res.data) {
        setSyncStatus(res.data);
      }
    } catch {
      // ignore transient status errors
    }
  }, []);

  const loadHistory = useCallback(async (channelId: string) => {
    const api = getApiClient();
    try {
      const res = await api.get<SyncHistoryData>(`/youtube/sync/history?channel_id=${channelId}&limit=20`);
      if (res.success && res.data) {
        setHistory(res.data.jobs || []);
      }
    } catch {
      // ignore transient history errors
    }
  }, []);

  const refreshAll = useCallback(async (channelId: string) => {
    await Promise.all([loadStatus(channelId), loadHistory(channelId)]);
  }, [loadStatus, loadHistory]);

  // Stop polling when sync is not running
  useEffect(() => {
    if (!syncStatus?.is_syncing && pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }, [syncStatus?.is_syncing]);

  // Poll while running
  useEffect(() => {
    if (syncStatus?.is_syncing && selectedChannel && !pollRef.current) {
      pollRef.current = setInterval(() => {
        void loadStatus(selectedChannel);
      }, 4000);
    }
  }, [syncStatus?.is_syncing, selectedChannel, loadStatus]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (pollRef.current) {
        clearInterval(pollRef.current);
        pollRef.current = null;
      }
    };
  }, []);

  // Initial load of channels + status + history
  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      router.push('/login');
      return;
    }

    const api = getApiClient();

    async function init() {
      try {
        const res = await api.get<Channel[]>('/integrations/youtube/channels');
        if (res.success && res.data && res.data.length > 0) {
          setChannels(res.data);
          const ch = res.data[0].id;
          setSelectedChannel(ch);
          await refreshAll(ch);
        } else {
          setError('No YouTube channel connected');
        }
      } catch (err) {
        if (err instanceof Error && err.name === 'UnauthorizedError') {
          router.push('/login');
          return;
        }
        setError(err instanceof Error ? err.message : 'Failed to load sync data');
      } finally {
        setLoading(false);
      }
    }

    void init();
    return () => {
      if (pollRef.current) {
        clearInterval(pollRef.current);
        pollRef.current = null;
      }
    };
  }, [router, refreshAll]);

  const handleVideoSync = useCallback(async (type: SyncType) => {
    if (!selectedChannel || videoBusy) return;
    setVideoBusy(true);
    setFeedback(null);
    setError(null);
    const api = getApiClient();
    try {
      const res = await api.post<SyncResponse>('/youtube/sync', {
        channel_id: selectedChannel,
        sync_type: type,
      });
      if (res.success && res.data) {
        setFeedback(`Video sync started (job: ${res.data.job_id})`);
        setSyncStatus((prev) => (prev ? { ...prev, is_syncing: true } : prev));
        await loadStatus(selectedChannel);
      } else {
        setError(res.error?.message || 'Failed to start video sync');
      }
    } catch (err) {
      if (err instanceof Error && err.name === 'UnauthorizedError') {
        router.push('/login');
        return;
      }
      setError(err instanceof Error ? err.message : 'Failed to start video sync');
    } finally {
      setVideoBusy(false);
    }
  }, [selectedChannel, videoBusy, router, loadStatus]);

  const handleAnalyticsSync = useCallback(async () => {
    if (!selectedChannel || analyticsBusy) return;
    setAnalyticsBusy(true);
    setFeedback(null);
    setError(null);
    const api = getApiClient();
    try {
      const res = await api.post<AnalyticsSyncResponse>('/youtube/analytics/sync', {
        channel_id: selectedChannel,
      });
      if (res.success && res.data) {
        setFeedback(`Analytics sync started (job: ${res.data.job_id})`);
        await loadStatus(selectedChannel);
      } else {
        setError(res.error?.message || 'Failed to start analytics sync');
      }
    } catch (err) {
      if (err instanceof Error && err.name === 'UnauthorizedError') {
        router.push('/login');
        return;
      }
      setError(err instanceof Error ? err.message : 'Failed to start analytics sync');
    } finally {
      setAnalyticsBusy(false);
    }
  }, [selectedChannel, analyticsBusy, router, loadStatus]);

  const handleRetry = useCallback(async (jobId: string) => {
    if (!selectedChannel) return;
    setError(null);
    setFeedback(null);
    const api = getApiClient();
    try {
      const res = await api.post<SyncResponse>('/youtube/sync/retry', { job_id: jobId });
      if (res.success && res.data) {
        setFeedback(`Retry started (job: ${res.data.job_id})`);
        await refreshAll(selectedChannel);
      } else {
        setError(res.error?.message || 'Failed to retry sync');
      }
    } catch (err) {
      if (err instanceof Error && err.name === 'UnauthorizedError') {
        router.push('/login');
        return;
      }
      setError(err instanceof Error ? err.message : 'Failed to retry sync');
    }
  }, [selectedChannel, router, refreshAll]);

  const busy = videoBusy || analyticsBusy || syncStatus?.is_syncing;

  return (
    <AppShell>
      <PageContainer title="Sync Management">
        {loading && (
          <div style={{ textAlign: 'center', padding: '48px', color: '#888' }}>
            Loading sync data...
          </div>
        )}

        {!loading && error && (
          <div style={{ padding: '16px', backgroundColor: '#3d1f1f', border: '1px solid #5a2a2a', borderRadius: '8px', color: '#ff6b6b', fontSize: '14px', marginBottom: '20px' }}>
            {error}
          </div>
        )}

        {!loading && !error && channels.length === 0 && (
          <div style={{ textAlign: 'center', padding: '48px', color: '#888' }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>📺</div>
            <p style={{ fontSize: '16px' }}>No YouTube channel connected</p>
          </div>
        )}

        {!loading && channels.length > 0 && (
          <>
            {feedback && (
              <div style={{ padding: '16px', backgroundColor: '#1a3d1a', border: '1px solid #2a5a2a', borderRadius: '8px', color: '#4ade80', fontSize: '14px', marginBottom: '20px' }}>
                {feedback}
              </div>
            )}

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px', flexWrap: 'wrap', gap: '12px' }}>
              {channels.length > 1 ? (
                <select
                  value={selectedChannel}
                  onChange={(e) => {
                    setSelectedChannel(e.target.value);
                    void refreshAll(e.target.value);
                  }}
                  style={{ padding: '8px 12px', backgroundColor: '#16213e', border: '1px solid #2a2a4a', borderRadius: '6px', color: '#e0e0e0', fontSize: '14px' }}
                >
                  {channels.map((ch) => (
                    <option key={ch.id} value={ch.id}>{ch.name}</option>
                  ))}
                </select>
              ) : (
                <div style={{ fontSize: '14px', color: '#888' }}>{channels[0].name}</div>
              )}
              <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
                <button
                  onClick={() => void handleVideoSync('manual')}
                  disabled={busy}
                  style={{
                    padding: '10px 20px',
                    backgroundColor: videoBusy ? '#555' : '#ff6b6b',
                    color: 'white',
                    border: 'none',
                    borderRadius: '6px',
                    fontSize: '14px',
                    fontWeight: '600',
                    cursor: busy ? 'not-allowed' : 'pointer',
                  }}
                >
                  {videoBusy ? 'Starting...' : 'Sync Videos'}
                </button>
                <button
                  onClick={() => void handleAnalyticsSync()}
                  disabled={busy}
                  style={{
                    padding: '10px 20px',
                    backgroundColor: analyticsBusy ? '#555' : '#4ecdc4',
                    color: '#0f0f23',
                    border: 'none',
                    borderRadius: '6px',
                    fontSize: '14px',
                    fontWeight: '600',
                    cursor: busy ? 'not-allowed' : 'pointer',
                  }}
                >
                  {analyticsBusy ? 'Starting...' : 'Sync Analytics'}
                </button>
              </div>
            </div>

            {/* Current status */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '16px', marginBottom: '24px' }}>
              <StatusCard label="Status" value={syncStatus?.last_sync_status || 'N/A'} isSyncing={syncStatus?.is_syncing} />
              <StatusCard label="Total Videos" value={formatNumber(syncStatus?.total_videos || 0)} />
              <StatusCard label="Synced" value={formatNumber(syncStatus?.total_synced || 0)} />
              <StatusCard label="Last Sync" value={syncStatus?.last_sync_at ? new Date(syncStatus.last_sync_at).toLocaleString() : 'Never'} />
              <StatusCard label="Running" value={syncStatus?.is_syncing ? 'Yes' : 'No'} />
            </div>

            {/* History */}
            <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', overflow: 'hidden' }}>
              <div style={{ padding: '16px 20px', borderBottom: '1px solid #2a2a4a' }}>
                <h3 style={{ fontSize: '16px', fontWeight: '600', color: '#e0e0e0' }}>Sync History</h3>
              </div>
              {history.length === 0 ? (
                <div style={{ padding: '32px', textAlign: 'center', color: '#888' }}>
                  No sync history yet.
                </div>
              ) : (
                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead>
                      <tr style={{ borderBottom: '1px solid #2a2a4a' }}>
                        <th style={{ textAlign: 'left', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Type</th>
                        <th style={{ textAlign: 'left', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Status</th>
                        <th style={{ textAlign: 'right', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Videos</th>
                        <th style={{ textAlign: 'right', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Success</th>
                        <th style={{ textAlign: 'right', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Failed</th>
                        <th style={{ textAlign: 'left', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Started</th>
                        <th style={{ textAlign: 'left', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Completed</th>
                        <th style={{ textAlign: 'left', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Error</th>
                        <th style={{ textAlign: 'right', padding: '10px 16px', fontSize: '11px', color: '#666' }}>Action</th>
                      </tr>
                    </thead>
                    <tbody>
                      {history.map((job) => (
                        <tr key={job.id} style={{ borderBottom: '1px solid #1a1a2e' }}>
                          <td style={{ padding: '10px 16px', fontSize: '13px', color: '#e0e0e0' }}>{job.sync_type}</td>
                          <td style={{ padding: '10px 16px', fontSize: '13px' }}>
                            <span style={{ color: job.status === 'completed' ? '#4ade80' : job.status === 'failed' ? '#f87171' : '#facc15' }}>
                              {job.status}
                            </span>
                          </td>
                          <td style={{ textAlign: 'right', padding: '10px 16px', fontSize: '13px', color: '#e0e0e0' }}>{formatNumber(job.total_videos)}</td>
                          <td style={{ textAlign: 'right', padding: '10px 16px', fontSize: '13px', color: '#4ade80' }}>{formatNumber(job.total_success)}</td>
                          <td style={{ textAlign: 'right', padding: '10px 16px', fontSize: '13px', color: job.total_failed > 0 ? '#f87171' : '#888' }}>{formatNumber(job.total_failed)}</td>
                          <td style={{ padding: '10px 16px', fontSize: '12px', color: '#888' }}>{job.started_at ? new Date(job.started_at).toLocaleString() : '-'}</td>
                          <td style={{ padding: '10px 16px', fontSize: '12px', color: '#888' }}>{job.completed_at ? new Date(job.completed_at).toLocaleString() : '-'}</td>
                          <td style={{ padding: '10px 16px', fontSize: '12px', color: job.error_message ? '#f87171' : '#666', maxWidth: '160px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                            {job.error_message || '-'}
                          </td>
                          <td style={{ textAlign: 'right', padding: '10px 16px' }}>
                            {job.status === 'failed' && (
                              <button
                                onClick={() => void handleRetry(job.id)}
                                disabled={busy}
                                style={{
                                  padding: '4px 12px',
                                  backgroundColor: '#2a2a4a',
                                  color: '#e0e0e0',
                                  border: '1px solid #3a3a5a',
                                  borderRadius: '4px',
                                  fontSize: '12px',
                                  cursor: busy ? 'not-allowed' : 'pointer',
                                }}
                              >
                                Retry
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </>
        )}
      </PageContainer>
    </AppShell>
  );
}

function StatusCard({ label, value, isSyncing }: { label: string; value: string; isSyncing?: boolean }) {
  return (
    <div style={{ backgroundColor: '#16213e', borderRadius: '8px', border: '1px solid #2a2a4a', padding: '16px' }}>
      <div style={{ fontSize: '11px', color: '#888', marginBottom: '6px', textTransform: 'uppercase', letterSpacing: '0.5px' }}>{label}</div>
      <div style={{ fontSize: '18px', fontWeight: '600', color: isSyncing ? '#facc15' : '#e0e0e0' }}>
        {isSyncing ? 'Syncing...' : value}
      </div>
    </div>
  );
}
