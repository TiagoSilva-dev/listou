/** Public origin of the web app, used for canonical URLs, OG images and share links. */
export function siteUrl(): string {
  return (process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000").replace(/\/$/, "");
}

/** Internal URL the Next.js server uses to reach the Go API. */
export function apiUrl(): string {
  return (process.env.API_URL ?? "http://localhost:8080").replace(/\/$/, "");
}
