'use client';

import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { User, LoginRequest, LoginResponse, ApiResponse } from '@/types';
import { getApiClient } from '@/lib/api/client';

interface Workspace {
  id: string;
  name: string;
  slug: string;
  description?: string;
  status: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (req: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  workspaceId: string | null;
  workspaces: Workspace[];
  activeWorkspace: Workspace | null;
  setActiveWorkspace: (ws: Workspace) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

function getStoredAuth(): { user: User | null; workspaceId: string | null } {
  if (typeof window === 'undefined') {
    return { user: null, workspaceId: null };
  }
  const token = localStorage.getItem('access_token');
  const storedUser = localStorage.getItem('user');
  const storedWorkspace = localStorage.getItem('workspace_id');
  if (token && storedUser) {
    try {
      return {
        user: JSON.parse(storedUser) as User,
        workspaceId: storedWorkspace,
      };
    } catch {
      localStorage.clear();
    }
  }
  return { user: null, workspaceId: null };
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [workspaceId, setWorkspaceId] = useState<string | null>(null);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [activeWorkspace, setActiveWorkspaceState] = useState<Workspace | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const initialized = useRef(false);

  const isAuthenticated = !!user;

  const setActiveWorkspace = useCallback((ws: Workspace) => {
    localStorage.setItem('workspace_id', ws.id);
    setWorkspaceId(ws.id);
    setActiveWorkspaceState(ws);
  }, []);

  useEffect(() => {
    if (initialized.current) return;
    initialized.current = true;
    const stored = getStoredAuth();
    setUser(stored.user);
    setWorkspaceId(stored.workspaceId);
    setIsLoading(false);

    // Hydrate workspaces whenever a user is authenticated, even if
    // workspace_id is not yet in localStorage (e.g. login page bypassed context).
    if (stored.user) {
      const controller = new AbortController();
      const api = getApiClient();

      async function hydrateWorkspaces() {
        try {
          const res = await api.get<Workspace[]>('/workspaces');
          if (res.success && res.data && res.data.length > 0) {
            const storedId = localStorage.getItem('workspace_id');
            const active = res.data.find((w) => w.id === storedId) || res.data[0];
            setWorkspaces(res.data);
            if (active) {
              setActiveWorkspaceState(active);
              setWorkspaceId(active.id);
              localStorage.setItem('workspace_id', active.id);
            }
          }
        } catch {
          // ignore
        }
      }

      void hydrateWorkspaces();
      return () => controller.abort();
    }
  }, []);

  const login = useCallback(async (req: LoginRequest) => {
    const api = getApiClient();
    const res: ApiResponse<LoginResponse> = await api.post<LoginResponse>('/auth/login', req);

    if (!res.success || !res.data) {
      throw new Error(res.error?.message || 'Login failed');
    }

    localStorage.setItem('access_token', res.data.access_token);
    localStorage.setItem('refresh_token', res.data.refresh_token);
    localStorage.setItem('user', JSON.stringify(res.data.user));

    setUser(res.data.user);

    // Fetch workspaces and set active after login
    const wsRes = await api.get<Workspace[]>('/workspaces');
    if (wsRes.success && wsRes.data && wsRes.data.length > 0) {
      const first = wsRes.data[0];
      setWorkspaces(wsRes.data);
      setActiveWorkspaceState(first);
      setWorkspaceId(first.id);
      localStorage.setItem('workspace_id', first.id);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      const api = getApiClient();
      await api.post('/auth/logout');
    } catch {
      // Ignore logout errors
    } finally {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
      localStorage.removeItem('workspace_id');
      setUser(null);
      setWorkspaceId(null);
      setActiveWorkspaceState(null);
      setWorkspaces([]);
      router.push('/login');
    }
  }, [router]);

  return (
    <AuthContext.Provider value={{ user, isAuthenticated, isLoading, login, logout, workspaceId, workspaces, activeWorkspace, setActiveWorkspace }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
