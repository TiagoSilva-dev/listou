import { ImageResponse } from "next/og";
import { eventTypeMeta, formatEventDate, PublicList } from "@listou/types";
import { THEMES, themeOf, type ThemeName } from "@/components/cover-art";
import { ApiError } from "@/lib/api";
import { serverApi } from "@/lib/server-api";

export const alt = "Lista de presentes no Listou";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

/** Social preview: "Lista de casamento · João & Maria · 12 de dezembro". */
export default async function OpenGraphImage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  let headline = "Lista de presentes";
  let names = "Listou";
  let date: string | null = null;
  let themeName: ThemeName = "blush";
  try {
    const data = await serverApi(`/public/lists/${encodeURIComponent(slug)}`, PublicList, {
      revalidate: 60,
    });
    headline = eventTypeMeta(data.event.type).headline;
    names = data.event.hostNames ?? data.event.title;
    date = formatEventDate(data.event.eventDate);
    themeName = themeOf(data.event.theme);
  } catch (err) {
    if (!(err instanceof ApiError)) throw err;
  }
  const theme = THEMES[themeName];
  const dark = themeName === "night";

  return new ImageResponse(
    <div
      style={{
        width: "100%",
        height: "100%",
        display: "flex",
        flexDirection: "column",
        justifyContent: "space-between",
        padding: 72,
        color: dark ? "#ffffff" : "#241c2c",
        backgroundColor: theme.base,
        backgroundImage: `radial-gradient(circle at 12% 18%, ${theme.a} 0%, transparent 55%), radial-gradient(circle at 90% 25%, ${theme.b} 0%, transparent 55%), radial-gradient(circle at 60% 110%, ${theme.c} 0%, transparent 60%)`,
      }}
    >
      <div
        style={{ display: "flex", alignItems: "center", gap: 14, fontSize: 34, fontWeight: 700 }}
      >
        <div
          style={{
            width: 44,
            height: 44,
            borderRadius: 14,
            background: "linear-gradient(135deg,#6b4fa0,#e8947a)",
          }}
        />
        Listou
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 18 }}>
        <div style={{ fontSize: 38, opacity: 0.75 }}>{headline}</div>
        <div
          style={{
            fontSize: names.length > 22 ? 84 : 112,
            fontWeight: 700,
            lineHeight: 1.02,
            letterSpacing: -2,
            maxWidth: 1000,
          }}
        >
          {names}
        </div>
        {date ? <div style={{ fontSize: 40, opacity: 0.8 }}>{date}</div> : null}
      </div>
      <div style={{ fontSize: 28, opacity: 0.7 }}>Escolha um presente sem precisar criar conta</div>
    </div>,
    size,
  );
}
