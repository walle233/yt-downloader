import type { FreeLimitErrorResponse } from "@ytvd/shared-types";

export const API_BASE = "/api/v1";

export class APIError extends Error {
  status: number;
  data?: unknown;

  constructor(message: string, status: number, data?: unknown) {
    super(message);
    this.name = "APIError";
    this.status = status;
    this.data = data;
  }
}

export async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

  const raw = await response.text();
  const parsed = raw ? safeParseJSON(raw) : undefined;

  if (!response.ok) {
    const message =
      (typeof parsed === "object" && parsed !== null && "error" in parsed && typeof parsed.error === "string" ? parsed.error : "") ||
      raw ||
      `Request failed: ${response.status}`;
    throw new APIError(message, response.status, parsed);
  }

  return (parsed ?? {}) as T;
}

export async function authorizedRequestJSON<T>(path: string, getToken: () => Promise<string | null>, init?: RequestInit): Promise<T> {
  const token = await getToken();
  if (!token) {
    throw new APIError("authentication required", 401);
  }

  return requestJSON<T>(path, {
    ...init,
    headers: {
      Authorization: `Bearer ${token}`,
      ...(init?.headers ?? {}),
    },
  });
}

export function isFreeLimitErrorResponse(data: unknown): data is FreeLimitErrorResponse {
  return typeof data === "object" && data !== null && "code" in data && data.code === "free_limit_reached" && "billing" in data;
}

function safeParseJSON(raw: string): unknown {
  try {
    return JSON.parse(raw);
  } catch {
    return raw;
  }
}
