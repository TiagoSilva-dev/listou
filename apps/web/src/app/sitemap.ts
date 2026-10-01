import type { MetadataRoute } from "next";
import { z } from "zod";
import { siteUrl } from "@/lib/env";
import { serverApi } from "@/lib/server-api";

export const revalidate = 300;

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const base = siteUrl();
  const entries: MetadataRoute.Sitemap = [{ url: base, changeFrequency: "monthly", priority: 1 }];
  try {
    const { lists } = await serverApi(
      "/public/sitemap",
      z.object({ lists: z.array(z.object({ slug: z.string(), updatedAt: z.string() })) }),
      { revalidate: 300 },
    );
    for (const l of lists)
      entries.push({
        url: `${base}/l/${l.slug}`,
        lastModified: new Date(l.updatedAt),
        changeFrequency: "daily",
        priority: 0.6,
      });
  } catch {
    // the API being down must not break the sitemap for the marketing pages
  }
  return entries;
}
