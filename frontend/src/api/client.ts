import { encryptRequestBody, decryptResponseData } from './e2e';

const API_BASE = '/api/v1';

let accessToken: string | null = null;
let refreshPromise: Promise<string | null> | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

async function refreshToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const res = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
      });
      if (!res.ok) return null;
      const data = await res.json();
      setAccessToken(data.access_token);
      return data.access_token;
    } catch {
      return null;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

export async function initAuth(): Promise<boolean> {
  const token = await refreshToken();
  return !!token;
}

export async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${API_BASE}${endpoint}`;
  const isFormData = options.body instanceof FormData;
  const headers: Record<string, string> = {
    ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
    ...(options.headers as Record<string, string> || {}),
  };

  if (!accessToken && !endpoint.includes('/auth/')) {
    await refreshToken();
  }

  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`;
  }

  // E2E: encrypt sensitive fields in request body
  const method = (options.method || 'GET').toUpperCase();
  if ((method === 'POST' || method === 'PUT') && typeof options.body === 'string') {
    options = { ...options, body: await encryptRequestBody(endpoint, options.body as string) };
  }

  let res = await fetch(url, { ...options, headers, credentials: 'include' });

  if (res.status === 401 && !endpoint.includes('/auth/')) {
    const newToken = await refreshToken();
    if (newToken) {
      headers['Authorization'] = `Bearer ${newToken}`;
      res = await fetch(url, { ...options, headers, credentials: 'include' });
    } else {
      setAccessToken(null);
      window.location.href = '/login';
      throw new Error('Session expired');
    }
  }

  if (res.status === 403) {
    const err = await res.json().catch(() => ({ error: 'Access denied' }));
    throw new Error(err.error || 'Access denied');
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }

  const data = await res.json();

  // E2E: decrypt sensitive fields in response
  return decryptResponseData<T>(endpoint, data);
}

export async function apiDownload(endpoint: string): Promise<void> {
  const url = `${API_BASE}${endpoint}`;
  const headers: Record<string, string> = {};

  if (!accessToken) {
    await refreshToken();
  }
  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`;
  }

  const res = await fetch(url, { headers, credentials: 'include' });

  if (res.status === 401) {
    const newToken = await refreshToken();
    if (newToken) {
      headers['Authorization'] = `Bearer ${newToken}`;
      const retry = await fetch(url, { headers, credentials: 'include' });
      if (!retry.ok) throw new Error(`Export failed: HTTP ${retry.status}`);
      return triggerDownload(retry);
    }
    throw new Error('Session expired');
  }

  if (!res.ok) throw new Error(`Export failed: HTTP ${res.status}`);
  return triggerDownload(res);
}

async function triggerDownload(res: Response) {
  const blob = await res.blob();
  const disposition = res.headers.get('Content-Disposition');
  let filename = 'export';
  if (disposition) {
    const match = disposition.match(/filename="?([^"]+)"?/);
    if (match) filename = match[1];
  }
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = filename;
  a.click();
  URL.revokeObjectURL(a.href);
}
