import type {
  AdminLoginResponse,
  AnonymousSessionResponse,
  CatalogRecommendationResponse,
  ChallengeVerifyResponse,
  CollectorScheduleResponse,
  CollectorScheduleUpsertRequest,
  GenerateBuildRequest,
  KeywordSeed,
  KeywordSeedImportResponse,
  KeywordSeedListResponse,
  KeywordSeedUpsertRequest,
} from './types';

const fingerprintStorageKey = 'givezj8-device-fingerprint';

type APIErrorShape = {
  error?: string | { message?: string; cooldown_seconds?: number; challenge_required?: boolean; risk_level?: string };
};

export class APIError extends Error {
  status: number;
  cooldownSeconds: number;
  challengeRequired: boolean;
  riskLevel: string;

  constructor(message: string, status: number, cooldownSeconds = 0, challengeRequired = false, riskLevel = '') {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.cooldownSeconds = cooldownSeconds;
    this.challengeRequired = challengeRequired;
    this.riskLevel = riskLevel;
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
    let challengeRequired = false;
    let riskLevel = '';
    try {
      const payload = (await response.json()) as APIErrorShape;
      if (typeof payload.error === 'string' && payload.error) {
        message = payload.error;
      } else if (payload.error?.message) {
        message = payload.error.message;
        cooldownSeconds = payload.error.cooldown_seconds ?? 0;
        challengeRequired = payload.error.challenge_required ?? false;
        riskLevel = payload.error.risk_level ?? '';
      }
    } catch {
      // Ignore parse error, keep fallback message.
    }
    throw new APIError(message, response.status, cooldownSeconds, challengeRequired, riskLevel);
  }
  return parseJSON<T>(response);
}

export function getAnonymousSession() {
  return requestJSON<AnonymousSessionResponse>('/api/v1/session/anonymous', {
    headers: { 'X-Device-Fingerprint': getDeviceFingerprint() },
  });
}

export function generateRecommendation(payload: GenerateBuildRequest, anonymousID: string) {
  return requestJSON<CatalogRecommendationResponse>('/catalog/recommend', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Anonymous-Id': anonymousID,
      'X-Device-Fingerprint': getDeviceFingerprint(),
    },
    body: JSON.stringify(payload),
  });
}

export function verifyChallenge(anonymousID: string, challengeToken: string) {
  return requestJSON<ChallengeVerifyResponse>('/api/v1/challenge/verify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      anonymous_id: anonymousID,
      device_fingerprint: getDeviceFingerprint(),
      challenge_token: challengeToken,
    }),
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

export function getJDScheduleConfig() {
  return requestJSON<CollectorScheduleResponse>('/admin/api/v1/jd/schedule');
}

export function updateJDScheduleConfig(payload: CollectorScheduleUpsertRequest) {
  return requestJSON<CollectorScheduleResponse>('/admin/api/v1/jd/schedule', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}

function getDeviceFingerprint() {
  const existing = window.localStorage.getItem(fingerprintStorageKey);
  if (existing) {
    return existing;
  }
  const payload = [
    navigator.userAgent,
    navigator.language,
    String(window.screen.width),
    String(window.screen.height),
    String(window.devicePixelRatio),
    typeof Intl !== 'undefined' ? Intl.DateTimeFormat().resolvedOptions().timeZone ?? '' : '',
  ].join('|');
  const generated = `fp_${hashString(payload)}_${Math.random().toString(36).slice(2, 10)}`;
  window.localStorage.setItem(fingerprintStorageKey, generated);
  return generated;
}

function hashString(input: string) {
  let hash = 2166136261;
  for (let i = 0; i < input.length; i += 1) {
    hash ^= input.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return Math.abs(hash).toString(36);
}
