import "server-only";
import { cookies } from "next/headers";
import { ApiErrorBody, User } from "@listou/types";
import type { ZodType } from "zod";
import { z } from "zod";
import { ApiError } from "./api";
import { apiUrl } from "./env";

interface ServerFetchOptions {
  /** Forward the visitor's cookies (authenticated, uncached). */
  auth?: boolean;
  /** Seconds to cache anonymous responses; ignored when `auth` is set. */
  revalidate?: number;
}

/** Server-side API call (Server Components / route handlers). */
export async function serverApi<T>(
  path: string,
  schema: ZodType<T>,
  opts: ServerFetchOptions = {},
): Promise<T> {
  const headers: Record<string, string> = {};
  if (opts.auth) {
    const jar = await cookies();
    const cookie = jar.toString();
    if (cookie) headers.cookie = cookie;
  }
  const res = await fetch(`${apiUrl()}/api/v1${path}`, {
    headers,
    ...(opts.auth || opts.revalidate === undefined
      ? { cache: "no-store" as const }
      : { next: { revalidate: opts.revalidate } }),
  });
  if (!res.ok) {
    let code = "UNKNOWN";
    let message = "Algo deu errado.";
    try {
      const parsed = ApiErrorBody.safeParse(await res.json());
      if (parsed.success) ({ code, message } = parsed.data.error);
    } catch {
      // keep defaults
    }
    throw new ApiError(res.status, code, message);
  }
  return schema.parse(await res.json());
}

/** The signed-in user, or null for anonymous visitors. */
export async function getCurrentUser(): Promise<User | null> {
  try {
    const { user } = await serverApi("/me", z.object({ user: User }), { auth: true });
    return user;
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) return null;
    throw err;
  }
}
