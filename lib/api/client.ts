import { ApiResponse } from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'https://cip.jaybani.my.id/api/v1';

export class UnauthorizedError extends Error {
  constructor() {
    super('Unauthorized');
    this.name = 'UnauthorizedError';
  }
}

class ApiClient {
  private baseUrl: string;
  private getToken: () => string | null;
  private getWorkspaceId: () => string | null;

  constructor(baseUrl: string, getToken: () => string | null, getWorkspaceId: () => string | null) {
    this.baseUrl = baseUrl;
    this.getToken = getToken;
    this.getWorkspaceId = getWorkspaceId;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
    headers?: Record<string, string>
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${path}`;
    const token = this.getToken();
    const workspaceId = this.getWorkspaceId();

    const defaultHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers,
    };

    if (token) {
      defaultHeaders['Authorization'] = `Bearer ${token}`;
    }

    if (workspaceId) {
      defaultHeaders['X-Workspace-ID'] = workspaceId;
    }

    const response = await fetch(url, {
      method,
      headers: defaultHeaders,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (response.status === 401) {
      throw new UnauthorizedError();
    }

    const data = await response.json();
    return data as ApiResponse<T>;
  }

  async get<T>(path: string, headers?: Record<string, string>): Promise<ApiResponse<T>> {
    return this.request<T>('GET', path, undefined, headers);
  }

  async post<T>(path: string, body?: unknown, headers?: Record<string, string>): Promise<ApiResponse<T>> {
    return this.request<T>('POST', path, body, headers);
  }

  async put<T>(path: string, body?: unknown, headers?: Record<string, string>): Promise<ApiResponse<T>> {
    return this.request<T>('PUT', path, body, headers);
  }

  async delete<T>(path: string, headers?: Record<string, string>): Promise<ApiResponse<T>> {
    return this.request<T>('DELETE', path, undefined, headers);
  }
}

let apiClient: ApiClient | null = null;

export function getApiClient(): ApiClient {
  if (!apiClient) {
    apiClient = new ApiClient(
      API_BASE_URL,
      () => {
        if (typeof window !== 'undefined') {
          return localStorage.getItem('access_token');
        }
        return null;
      },
      () => {
        if (typeof window !== 'undefined') {
          return localStorage.getItem('workspace_id');
        }
        return null;
      }
    );
  }
  return apiClient;
}

export function createApiClient(getToken: () => string | null): ApiClient {
  return new ApiClient(API_BASE_URL, getToken, () => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('workspace_id');
    }
    return null;
  });
}
