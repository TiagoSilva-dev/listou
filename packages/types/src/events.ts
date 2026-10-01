import type { EventType } from "./enums";

export interface EventTypeMeta {
  type: EventType;
  label: string;
  emoji: string;
  /** Short headline used on cards and social previews. */
  headline: string;
  suggestedCategories: { name: string; emoji: string }[];
}

export const EVENT_TYPES: EventTypeMeta[] = [
  {
    type: "BABY_SHOWER",
    label: "Chá de bebê",
    emoji: "👶",
    headline: "Chá de bebê",
    suggestedCategories: [
      { name: "Quarto do bebê", emoji: "🧸" },
      { name: "Roupinhas", emoji: "👕" },
      { name: "Higiene", emoji: "🛁" },
      { name: "Passeio", emoji: "🚼" },
    ],
  },
  {
    type: "WEDDING",
    label: "Casamento",
    emoji: "💍",
    headline: "Lista de casamento",
    suggestedCategories: [
      { name: "Cozinha", emoji: "🍳" },
      { name: "Mesa posta", emoji: "🍽️" },
      { name: "Quarto", emoji: "🛏️" },
      { name: "Lua de mel", emoji: "✈️" },
    ],
  },
  {
    type: "HOUSEWARMING",
    label: "Casa nova",
    emoji: "🏠",
    headline: "Chá de casa nova",
    suggestedCategories: [
      { name: "Cozinha", emoji: "🍳" },
      { name: "Quarto", emoji: "🛏️" },
      { name: "Sala", emoji: "🛋️" },
      { name: "Banheiro", emoji: "🛁" },
      { name: "Limpeza", emoji: "🧺" },
    ],
  },
  {
    type: "BIRTHDAY",
    label: "Aniversário",
    emoji: "🎂",
    headline: "Aniversário",
    suggestedCategories: [{ name: "Desejos", emoji: "✨" }],
  },
  {
    type: "GRADUATION",
    label: "Formatura",
    emoji: "🎓",
    headline: "Formatura",
    suggestedCategories: [{ name: "Desejos", emoji: "✨" }],
  },
  {
    type: "TRAVEL",
    label: "Viagem",
    emoji: "✈️",
    headline: "Lista de viagem",
    suggestedCategories: [
      { name: "Malas", emoji: "🧳" },
      { name: "Experiências", emoji: "🌅" },
    ],
  },
  {
    type: "CHRISTMAS",
    label: "Natal",
    emoji: "🎄",
    headline: "Lista de Natal",
    suggestedCategories: [{ name: "Desejos", emoji: "🎁" }],
  },
  {
    type: "WISHLIST",
    label: "Wishlist",
    emoji: "⭐",
    headline: "Wishlist",
    suggestedCategories: [{ name: "Desejos", emoji: "✨" }],
  },
  {
    type: "CUSTOM",
    label: "Personalizada",
    emoji: "❤️",
    headline: "Lista especial",
    suggestedCategories: [],
  },
];

export function eventTypeMeta(type: EventType): EventTypeMeta {
  return EVENT_TYPES.find((t) => t.type === type) ?? EVENT_TYPES[EVENT_TYPES.length - 1]!;
}
