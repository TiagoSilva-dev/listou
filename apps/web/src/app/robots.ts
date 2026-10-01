import type { MetadataRoute } from "next";
import { siteUrl } from "@/lib/env";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: ["/", "/l/"],
        disallow: ["/dashboard", "/criar", "/entrar", "/criar-conta", "/api/", "/go/", "/demo/"],
      },
    ],
    sitemap: `${siteUrl()}/sitemap.xml`,
  };
}
