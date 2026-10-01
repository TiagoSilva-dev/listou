"use client";

import { useEffect, useState, useSyncExternalStore } from "react";
import { Check, Copy, Mail, MessageCircle, Share2, Instagram, Download } from "lucide-react";
import QRCode from "qrcode";
import { Button, Sheet, useToast } from "@listou/ui";
import { apiVoid } from "@/lib/api";

export interface ShareDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  url: string;
  title: string;
  slug: string;
}

export function ShareDialog({ open, onOpenChange, url, title, slug }: ShareDialogProps) {
  const toast = useToast();
  const [qr, setQr] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const canNativeShare = useSyncExternalStore(
    () => () => undefined,
    () => typeof navigator.share === "function",
    () => false,
  );

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    QRCode.toDataURL(url, { margin: 1, width: 320, color: { dark: "#241c2c", light: "#ffffff" } })
      .then((data) => !cancelled && setQr(data))
      .catch(() => !cancelled && setQr(null));
    return () => {
      cancelled = true;
    };
  }, [open, url]);

  const message = `Olha a nossa lista: ${title} 💜`;

  function track(channel: string) {
    void apiVoid("/analytics/track", {
      method: "POST",
      body: { name: "SHARE_CREATED", slug, properties: { channel } },
    }).catch(() => undefined);
  }

  async function copy(channel = "copy") {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      toast("Link copiado!", "success");
      window.setTimeout(() => setCopied(false), 2000);
      track(channel);
    } catch {
      toast("Não foi possível copiar. Selecione o link manualmente.", "error");
    }
  }

  async function nativeShare() {
    try {
      await navigator.share({ title, text: message, url });
      track("native");
    } catch {
      // user dismissed the sheet
    }
  }

  const options = [
    {
      key: "whatsapp",
      label: "WhatsApp",
      icon: MessageCircle,
      href: `https://wa.me/?text=${encodeURIComponent(`${message}\n${url}`)}`,
    },
    {
      key: "email",
      label: "E-mail",
      icon: Mail,
      href: `mailto:?subject=${encodeURIComponent(title)}&body=${encodeURIComponent(`${message}\n\n${url}`)}`,
    },
  ];

  return (
    <Sheet
      open={open}
      onOpenChange={onOpenChange}
      title="Compartilhar lista"
      description="Quem receber o link não precisa criar conta."
    >
      <div className="flex flex-col gap-6">
        <div className="rounded-control bg-canvas-deep flex items-center gap-2 p-1.5 pl-4">
          <input
            readOnly
            value={url}
            aria-label="Link da lista"
            onFocus={(e) => e.currentTarget.select()}
            className="text-ink-soft min-w-0 flex-1 bg-transparent text-sm outline-none"
          />
          <Button size="sm" onClick={() => copy()}>
            {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
            {copied ? "Copiado" : "Copiar"}
          </Button>
        </div>

        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          {options.map((o) => (
            <a
              key={o.key}
              href={o.href}
              target={o.key === "email" ? undefined : "_blank"}
              rel="noopener noreferrer"
              onClick={() => track(o.key)}
              className="rounded-card bg-canvas-deep hover:bg-primary-soft hover:text-primary flex flex-col items-center gap-2 p-4 text-sm font-semibold transition-colors"
            >
              <o.icon className="size-5" />
              {o.label}
            </a>
          ))}
          <button
            type="button"
            onClick={async () => {
              await copy("instagram");
              toast("Link copiado — cole no link da bio ou nos stories.", "success");
            }}
            className="rounded-card bg-canvas-deep hover:bg-primary-soft hover:text-primary flex flex-col items-center gap-2 p-4 text-sm font-semibold transition-colors"
          >
            <Instagram className="size-5" />
            Instagram
          </button>
          {canNativeShare ? (
            <button
              type="button"
              onClick={nativeShare}
              className="rounded-card bg-canvas-deep hover:bg-primary-soft hover:text-primary flex flex-col items-center gap-2 p-4 text-sm font-semibold transition-colors"
            >
              <Share2 className="size-5" />
              Mais opções
            </button>
          ) : null}
        </div>

        <div className="rounded-card border-line flex items-center gap-5 border p-4">
          {qr ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={qr} alt={`QR Code para ${title}`} className="rounded-chip size-28" />
          ) : (
            <div className="animate-shimmer rounded-chip bg-canvas-deep size-28" />
          )}
          <div className="flex flex-1 flex-col gap-2">
            <p className="font-semibold">QR Code</p>
            <p className="text-ink-muted text-sm">
              Perfeito para convites impressos e lembrancinhas.
            </p>
            {qr ? (
              <a
                href={qr}
                download={`qrcode-${slug}.png`}
                onClick={() => track("qrcode")}
                className="text-primary inline-flex w-fit items-center gap-1.5 text-sm font-semibold hover:underline"
              >
                <Download className="size-4" /> Baixar imagem
              </a>
            ) : null}
          </div>
        </div>
      </div>
    </Sheet>
  );
}
