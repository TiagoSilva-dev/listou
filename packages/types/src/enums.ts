import { z } from "zod";

export const EventType = z.enum([
  "BABY_SHOWER",
  "WEDDING",
  "HOUSEWARMING",
  "BIRTHDAY",
  "GRADUATION",
  "TRAVEL",
  "CHRISTMAS",
  "WISHLIST",
  "CUSTOM",
]);
export type EventType = z.infer<typeof EventType>;

export const Visibility = z.enum(["PUBLIC", "UNLISTED", "PRIVATE"]);
export type Visibility = z.infer<typeof Visibility>;

export const EventStatus = z.enum(["DRAFT", "PUBLISHED", "ARCHIVED"]);
export type EventStatus = z.infer<typeof EventStatus>;

export const ItemStatus = z.enum([
  "AVAILABLE",
  "PARTIALLY_RESERVED",
  "RESERVED",
  "PURCHASED",
  "ARCHIVED",
]);
export type ItemStatus = z.infer<typeof ItemStatus>;

export const Priority = z.enum(["HIGH", "MEDIUM", "LOW"]);
export type Priority = z.infer<typeof Priority>;

export const Availability = z.enum(["IN_STOCK", "OUT_OF_STOCK", "UNKNOWN"]);
export type Availability = z.infer<typeof Availability>;

export const ReservationKind = z.enum(["RESERVATION", "PURCHASE"]);
export type ReservationKind = z.infer<typeof ReservationKind>;
