import { z } from "zod";
import { EventType, Priority, Visibility } from "./enums";

/** Client-side form schemas. The API re-validates everything. */

export const RegisterInput = z.object({
  name: z.string().trim().min(2, "Conte seu nome").max(80),
  email: z.email("E-mail inválido").trim().toLowerCase(),
  password: z.string().min(8, "Use pelo menos 8 caracteres").max(128),
});
export type RegisterInput = z.infer<typeof RegisterInput>;

export const LoginInput = z.object({
  email: z.email("E-mail inválido").trim().toLowerCase(),
  password: z.string().min(1, "Informe sua senha"),
});
export type LoginInput = z.infer<typeof LoginInput>;

export const SLUG_PATTERN = /^[a-z0-9]+(-[a-z0-9]+)*$/;

export const EventInput = z.object({
  type: EventType,
  title: z.string().trim().min(1, "Dê um nome ao momento").max(120),
  hostNames: z.string().trim().max(120).optional(),
  description: z.string().trim().max(2000).optional(),
  eventDate: z
    .string()
    .regex(/^\d{4}-\d{2}-\d{2}$/)
    .optional()
    .or(z.literal("")),
  location: z.string().trim().max(160).optional(),
  coverImageUrl: z.url().optional().or(z.literal("")),
  slug: z
    .string()
    .trim()
    .min(3)
    .max(60)
    .regex(SLUG_PATTERN, "Use letras minúsculas, números e hífens")
    .optional()
    .or(z.literal("")),
  visibility: Visibility.optional(),
});
export type EventInput = z.infer<typeof EventInput>;

export const ItemInput = z.object({
  title: z.string().trim().min(1, "Dê um nome ao item").max(160),
  description: z.string().trim().max(1000).optional(),
  imageUrl: z.url("Link de imagem inválido").optional().or(z.literal("")),
  externalUrl: z.url("Link inválido").optional().or(z.literal("")),
  priceReferenceCents: z.number().int().min(0).optional(),
  desiredQuantity: z.number().int().min(1).max(999),
  priority: Priority.optional(),
  categoryId: z.string().optional(),
  notes: z.string().trim().max(1000).optional(),
});
export type ItemInput = z.infer<typeof ItemInput>;

export const ReservationInput = z.object({
  guestName: z.string().trim().min(1, "Como devemos te chamar?").max(80),
  guestContact: z.string().trim().max(120).optional(),
  message: z.string().trim().max(500).optional(),
  quantity: z.number().int().min(1).max(99),
  kind: z.enum(["RESERVATION", "PURCHASE"]),
});
export type ReservationInput = z.infer<typeof ReservationInput>;
