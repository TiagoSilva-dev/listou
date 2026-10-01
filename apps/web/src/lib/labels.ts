import type { ItemStatus } from "@listou/types";

export const STATUS_LABEL: Record<ItemStatus, string> = {
  AVAILABLE: "Disponível",
  PARTIALLY_RESERVED: "Parcialmente reservado",
  RESERVED: "Reservado",
  PURCHASED: "Comprado",
  ARCHIVED: "Arquivado",
};

export const STATUS_TONE: Record<ItemStatus, "success" | "warning" | "primary" | "neutral"> = {
  AVAILABLE: "success",
  PARTIALLY_RESERVED: "warning",
  RESERVED: "warning",
  PURCHASED: "primary",
  ARCHIVED: "neutral",
};

export function plural(n: number, one: string, many: string): string {
  return `${n} ${n === 1 ? one : many}`;
}
