import type {
  AdminLoginResponse,
  AnonymousSessionResponse,
  CatalogRecommendationResponse,
  GenerateBuildRequest,
  KeywordSeed,
  KeywordSeedImportResponse,
  KeywordSeedListResponse,
  KeywordSeedUpsertRequest,
} from './types';

type APIErrorShape = {
  error?: string | { message?: string; cooldown_seconds?: number };
};

export class APIError extends Error {
  status: number;
  cooldownSeconds: number;

  constructor(message: string, status: number, cooldownSeconds = 0) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.cooldownSeconds = cooldownSeconds;
  }
}

async function parseJSON<T>(response: Response): Promise<T> {
  return response.json() as Promise<T>;
}

async function requestJSON<T>(input: RequestInfo | URL, init?: RequestInit): Promise<T> {
  const response = await fetch(input, {
    credentials: 'same-origin',
    ...init,
  });
  if (!response.ok) {
    let message = `请求失败 (${response.status})`;
    let cooldownSeconds = 0;
    try {
      const payload = (await response.json()) as APIErrorShape;
      if (typeof payload.error === 'string' && payload.error) {
        message = payload.error;
      } else if (payload.error?.message) {
        message = payload.error.message;
        cooldownSeconds = payload.error.cooldown_seconds ?? 0;
      }
    } catch {
      // Ignore parse error, keep fallback message.
    }
    throw new APIError(message, response.status, cooldownSeconds);
  }
  return parseJSON<T>(response);
}

export function getAnonymousSession() {
  return requestJSON<AnonymousSessionResponse>('/api/v1/session/anonymous');
}

export function generateRecommendation(payload: GenerateBuildRequest, anonymousID: string) {
  return requestJSON<CatalogRecommendationResponse>('/catalog/recommend', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Anonymous-Id': anonymousID,
    },
    body: JSON.stringify(payload),
  });
}

export function loginAdmin(username: string, password: string) {
  return requestJSON<AdminLoginResponse>('/admin/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
}

export function logoutAdmin() {
  return requestJSON<{ ok: boolean }>('/admin/logout', { method: 'POST' });
}

export function listKeywordSeeds(query: URLSearchParams) {
  return requestJSON<KeywordSeedListResponse>(`/admin/api/v1/keyword-seeds?${query.toString()}`);
}

export function getKeywordSeed(id: string) {
  return requestJSON<KeywordSeed>(`/admin/api/v1/keyword-seeds/${id}`);
}

export function createKeywordSeed(payload: KeywordSeedUpsertRequest) {
  return requestJSON<KeywordSeed>('/admin/api/v1/keyword-seeds', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}

export function updateKeywordSeed(id: string, payload: KeywordSeedUpsertRequest) {
  return requestJSON<KeywordSeed>(`/admin/api/v1/keyword-seeds/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}

export function setKeywordSeedEnabled(id: string, enabled: boolean) {
  return requestJSON<KeywordSeed>(`/admin/api/v1/keyword-seeds/${id}/${enabled ? 'enable' : 'disable'}`, {
    method: 'POST',
  });
}

export function importKeywordSeeds(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return requestJSON<KeywordSeedImportResponse>('/admin/api/v1/keyword-seeds/import', {
    method: 'POST',
    body: formData,
  });
}
