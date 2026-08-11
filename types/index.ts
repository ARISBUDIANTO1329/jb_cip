export interface User {
  id: string;
  email: string;
  name: string;
  status: string;
  avatar_url?: string;
  email_verified_at?: string;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}


export interface Workspace {
  id: string;
  owner_id: string;
  name: string;
  slug: string;
  description?: string;
  status: string;
}

export interface Channel {
  id: string;
  workspace_id: string;
  external_id: string;
  name: string;
  description?: string;
  subscriber_count: number;
  view_count: number;
  video_count: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Video {
  id: string;
  channel_id: string;
  video_id: string;
  title: string;
  description?: string;
  published_at?: string;
  duration: number;
  thumbnail_url?: string;
  privacy_status: string;
  view_count: number;
  like_count: number;
  comment_count: number;
  created_at: string;
  updated_at: string;
}

export interface SyncJob {
  id: string;
  channel_id: string;
  user_id: string;
  workspace_id: string;
  sync_type: string;
  status: string;
  total_videos: number;
  total_success: number;
  total_failed: number;
  duration_seconds: number;
  error_message?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

export interface DailyMetric {
  id: string;
  channel_id: string;
  video_id?: string;
  date: string;
  metric_type: string;
  views: number;
  watch_time: number;
  average_view_duration: number;
  average_percentage_viewed: number;
  impressions: number;
  impression_ctr: number;
  likes: number;
  comments: number;
  shares: number;
  subscribers: number;
  returning_viewers: number;
  new_viewers: number;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  error?: {
    code: string;
    message: string;
    trace_id: string;
  };
  meta: {
    timestamp: string;
    request_id: string;
    api_version: string;
  };
  pagination?: {
    limit: number;
    offset: number;
    total: number;
  };
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface SyncStatusResponse {
  channel_id: string;
  last_sync_at?: string;
  last_sync_status?: string;
  total_videos: number;
  total_synced: number;
  is_syncing: boolean;
}

export interface SyncResponse {
  job_id: string;
  status: string;
  total_videos: number;
  total_success: number;
  total_failed: number;
  duration_seconds: number;
  sync_type: string;
  channel_id: string;
}

export interface AnalyticsSummary {
  channel_id: string;
  views: number;
  watch_time: number;
  likes: number;
  comments: number;
  shares: number;
  subscribers: number;
  start_date: string;
  end_date: string;
}

export interface AnalyticsTimeseriesPoint {
  date: string;
  value: number;
}

export interface TopVideoAnalytics {
  video_id: string;
  internal_video_id: string;
  title: string;
  thumbnail_url: string;
  views: number;
  likes: number;
  comments: number;
  watch_time: number;
}

export interface VideoListData {
  data: Video[];
  pagination: {
    limit: number;
    offset: number;
    total: number;
  };
}
