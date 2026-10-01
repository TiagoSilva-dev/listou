import { z } from "zod";
import {
  Availability,
  EventStatus,
  EventType,
  ItemStatus,
  Priority,
  ReservationKind,
  Visibility,
} from "./enums";

export const ApiErrorBody = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
    requestId: z.string(),
    fields: z.record(z.string(), z.string()).optional(),
  }),
});
export type ApiErrorBody = z.infer<typeof ApiErrorBody>;

export const User = z.object({
  id: z.string(),
  email: z.string(),
  name: z.string(),
  avatarUrl: z.string().nullable(),
  role: z.enum(["USER", "ADMIN"]),
});
export type User = z.infer<typeof User>;

export const Merchant = z.object({ code: z.string(), name: z.string() });
export type Merchant = z.infer<typeof Merchant>;

export const Offer = z.object({
  id: z.string(),
  merchant: Merchant,
  title: z.string(),
  priceCents: z.number().int().nullable(),
  originalPriceCents: z.number().int().nullable(),
  currency: z.string(),
  availability: Availability,
  imageUrl: z.string().nullable(),
  /** Relative outbound URL (/go/{offerId}). Raw merchant URLs are never sent. */
  goUrl: z.string(),
  lastSyncedAt: z.string().nullable(),
  demo: z.boolean(),
});
export type Offer = z.infer<typeof Offer>;

export const ProductSummary = z.object({
  id: z.string(),
  canonicalTitle: z.string(),
  brand: z.string().nullable(),
  imageUrl: z.string().nullable(),
  emoji: z.string().nullable(),
  /** True for fictitious mock-provider data, which the UI must label. */
  demo: z.boolean(),
});
export type ProductSummary = z.infer<typeof ProductSummary>;

export const Category = z.object({
  id: z.string(),
  name: z.string(),
  emoji: z.string().nullable(),
  position: z.number().int(),
});
export type Category = z.infer<typeof Category>;

export const ListItem = z.object({
  id: z.string(),
  listId: z.string(),
  categoryId: z.string().nullable(),
  title: z.string(),
  description: z.string().nullable(),
  notes: z.string().nullable(),
  imageUrl: z.string().nullable(),
  emoji: z.string().nullable(),
  externalUrl: z.string().nullable(),
  priceReferenceCents: z.number().int().nullable(),
  currency: z.string(),
  priority: Priority,
  desiredQuantity: z.number().int(),
  purchasedQuantity: z.number().int(),
  reservedQuantity: z.number().int(),
  availableQuantity: z.number().int(),
  status: ItemStatus,
  position: z.number().int(),
  product: ProductSummary.nullable(),
  offers: z.array(Offer),
  createdAt: z.string(),
});
export type ListItem = z.infer<typeof ListItem>;

export const Event = z.object({
  id: z.string(),
  type: EventType,
  title: z.string(),
  slug: z.string(),
  description: z.string().nullable(),
  hostNames: z.string().nullable(),
  eventDate: z.string().nullable(),
  location: z.string().nullable(),
  coverImageUrl: z.string().nullable(),
  avatarUrl: z.string().nullable(),
  theme: z.string(),
  visibility: Visibility,
  status: EventStatus,
  surpriseMode: z.boolean(),
  showReserverNames: z.boolean(),
  publishedAt: z.string().nullable(),
  listId: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});
export type Event = z.infer<typeof Event>;

export const ListProgress = z.object({
  totalUnits: z.number().int(),
  reservedUnits: z.number().int(),
  purchasedUnits: z.number().int(),
  availableUnits: z.number().int(),
});
export type ListProgress = z.infer<typeof ListProgress>;

export const PublicEvent = Event.pick({
  type: true,
  title: true,
  slug: true,
  description: true,
  hostNames: true,
  eventDate: true,
  location: true,
  coverImageUrl: true,
  avatarUrl: true,
  theme: true,
  visibility: true,
});
export type PublicEvent = z.infer<typeof PublicEvent>;

export const PublicItem = ListItem.omit({
  notes: true,
  listId: true,
  createdAt: true,
  position: true,
});
export type PublicItem = z.infer<typeof PublicItem>;

export const PublicList = z.object({
  event: PublicEvent,
  list: z.object({ id: z.string(), title: z.string(), allowReservations: z.boolean() }),
  categories: z.array(Category),
  items: z.array(PublicItem),
  progress: ListProgress,
  preview: z.boolean(),
});
export type PublicList = z.infer<typeof PublicList>;

export const SearchResult = z.object({
  providerCode: z.string(),
  externalId: z.string(),
  title: z.string(),
  brand: z.string().nullable(),
  imageUrl: z.string().nullable(),
  emoji: z.string().nullable(),
  category: z.string().nullable(),
  /** Only present when the source's terms allow showing ratings. */
  rating: z.number().nullable(),
  demo: z.boolean(),
  offers: z.array(
    z.object({
      merchant: Merchant,
      priceCents: z.number().int().nullable(),
      availability: Availability,
    }),
  ),
  lowestPriceCents: z.number().int().nullable(),
});
export type SearchResult = z.infer<typeof SearchResult>;

/** What could be read from a pasted product link. No price: it is never taken from a page. */
export const LinkPreview = z.object({
  url: z.string(),
  title: z.string(),
  imageUrl: z.string().nullable(),
  storeName: z.string(),
  /** False when the page gave no title; the user must type the item name. */
  readable: z.boolean(),
});
export type LinkPreview = z.infer<typeof LinkPreview>;

export const Reservation = z.object({
  id: z.string(),
  itemId: z.string(),
  kind: ReservationKind,
  quantity: z.number().int(),
  status: z.enum(["ACTIVE", "CANCELLED", "EXPIRED", "CONFIRMED"]),
  expiresAt: z.string().nullable(),
  createdAt: z.string(),
  /** Returned once, at creation. Lets the guest cancel without an account. */
  manageToken: z.string().optional(),
});
export type Reservation = z.infer<typeof Reservation>;

export const RecommendedCategory = z.object({
  name: z.string(),
  emoji: z.string(),
  reason: z.string(),
  desires: z
    .array(z.object({ title: z.string(), emoji: z.string(), quantity: z.number().int() }))
    .nullable(),
});
export type RecommendedCategory = z.infer<typeof RecommendedCategory>;

export const SuggestedDesire = z.object({
  title: z.string(),
  emoji: z.string(),
  quantity: z.number().int(),
  importance: z.enum(["ESSENTIAL", "RECOMMENDED", "OPTIONAL"]),
});
export type SuggestedDesire = z.infer<typeof SuggestedDesire>;

export const SuggestedCategory = z.object({
  name: z.string(),
  emoji: z.string(),
  reason: z.string(),
  desires: z.array(SuggestedDesire),
});
export type SuggestedCategory = z.infer<typeof SuggestedCategory>;

export const Suggestions = z.object({
  categories: z.array(SuggestedCategory),
  source: z.string(),
});
export type Suggestions = z.infer<typeof Suggestions>;

export const ApplySuggestionsResult = z.object({
  addedItems: z.number().int(),
  createdCategories: z.number().int(),
});
export type ApplySuggestionsResult = z.infer<typeof ApplySuggestionsResult>;

export const DashboardActivity = z.object({
  kind: z.enum(["RESERVED", "PURCHASED", "CANCELLED", "CLICKED"]),
  itemTitle: z.string().nullable(),
  guestName: z.string().nullable(),
  at: z.string(),
});
export type DashboardActivity = z.infer<typeof DashboardActivity>;

export const Dashboard = z.object({
  event: Event,
  stats: z.object({
    daysRemaining: z.number().int().nullable(),
    itemsCount: z.number().int(),
    totalUnits: z.number().int(),
    reservedUnits: z.number().int(),
    purchasedUnits: z.number().int(),
    views: z.number().int(),
    outboundClicks: z.number().int(),
  }),
  surpriseMode: z.boolean(),
  recentActivity: z.array(DashboardActivity),
});
export type Dashboard = z.infer<typeof Dashboard>;

export const GiftListMeta = z.object({
  id: z.string(),
  eventId: z.string(),
  title: z.string(),
  description: z.string().nullable(),
  allowReservations: z.boolean(),
  allowGroupContributions: z.boolean(),
});
export type GiftListMeta = z.infer<typeof GiftListMeta>;

/** Owner view of a list (GET /events/{id}/list). */
export const OwnerList = z.object({
  list: GiftListMeta,
  categories: z.array(Category),
  items: z.array(ListItem),
});
export type OwnerList = z.infer<typeof OwnerList>;

/** Admin: a hand-curated product and the affiliate links we generated for it. */
export const CuratedOffer = z.object({ merchant: z.string(), url: z.string() });
export type CuratedOffer = z.infer<typeof CuratedOffer>;

export const CuratedMerchant = z.object({ code: z.string(), name: z.string() });
export type CuratedMerchant = z.infer<typeof CuratedMerchant>;

export const CuratedProduct = z.object({
  id: z.string(),
  externalId: z.string(),
  title: z.string(),
  brand: z.string(),
  category: z.string(),
  keywords: z.string(),
  imageUrl: z.string(),
  active: z.boolean(),
  offers: z.array(CuratedOffer),
});
export type CuratedProduct = z.infer<typeof CuratedProduct>;
