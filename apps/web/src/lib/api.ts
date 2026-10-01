import { ApiErrorBody } from "@listou/types";
import type { ZodType } from "zod";

/** Error thrown for non-2xx API responses, carrying the stable error code. */
export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
    public readonly fields?: Record<string, string>,
    public readonly requestId?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function toError(res: Response): Promise<ApiError> {
  try {
    const parsed = ApiErrorBody.safeParse(await res.json());
    if (parsed.success) {
      const { code, message, fields, requestId } = parsed.data.error;
      return new ApiError(res.status, code, message, fields, requestId);
    }
  } catch {
    // fall through
  }
  return new ApiError(res.status, "UNKNOWN", "Algo deu errado. Tente novamente.");
}

export interface RequestOptions {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  headers?: Record<string, string>;
}

/**
 * Browser-side call to the API through the same-origin rewrite, so the
 * session cookie stays first-party.
 */
export async function api<T>(
  path: string,
  schema: ZodType<T>,
  opts: RequestOptions = {},
): Promise<T> {
  const res = await fetch(`/api/v1${path}`, {
    method: opts.method ?? "GET",
    headers: { ...(opts.body ? { "Content-Type": "application/json" } : {}), ...opts.headers },
    body: opts.body ? JSON.stringify(opts.body) : undefined,
    credentials: "same-origin",
    cache: "no-store",
  });
  if (!res.ok) throw await toError(res);
  return schema.parse(await res.json());
}

/** Same as `api` for endpoints that answer 204 No Content. */
export async function apiVoid(path: string, opts: RequestOptions = {}): Promise<void> {
  const res = await fetch(`/api/v1${path}`, {
    method: opts.method ?? "POST",
    headers: { ...(opts.body ? { "Content-Type": "application/json" } : {}), ...opts.headers },
    body: opts.body ? JSON.stringify(opts.body) : undefined,
    credentials: "same-origin",
    cache: "no-store",
  });
  if (!res.ok && res.status !== 202 && res.status !== 204) throw await toError(res);
}

/** Maps API field errors onto react-hook-form's setError. */
export function applyFieldErrors(
  err: unknown,
  setError: (name: never, error: { message: string }) => void,
): string | null {
  if (!(err instanceof ApiError)) return "Algo deu errado. Tente novamente.";
  if (err.fields) {
    for (const [field, message] of Object.entries(err.fields)) {
      setError(field as never, { message });
    }
    return null;
  }
  return err.message;
}
