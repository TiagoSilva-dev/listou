/**
 * Guests have no account, so the token that lets them cancel or confirm a
 * reservation lives in their own browser (localStorage), never on the server
 * in clear text.
 */
export interface StoredReservation {
  id: string;
  token: string;
  itemId: string;
  itemTitle: string;
  kind: "RESERVATION" | "PURCHASE";
  quantity: number;
  createdAt: string;
}

import { useSyncExternalStore } from "react";

const key = (slug: string) => `listou:reservations:${slug}`;
const EMPTY: StoredReservation[] = [];
const listeners = new Set<() => void>();
const snapshots = new Map<string, { raw: string | null; parsed: StoredReservation[] }>();

function emit() {
  listeners.forEach((l) => l());
}

export function loadReservations(slug: string): StoredReservation[] {
  try {
    const raw = window.localStorage.getItem(key(slug));
    return raw ? (JSON.parse(raw) as StoredReservation[]) : [];
  } catch {
    return [];
  }
}

function save(slug: string, list: StoredReservation[]) {
  try {
    window.localStorage.setItem(key(slug), JSON.stringify(list));
  } catch {
    // storage unavailable (private mode): the reservation still exists server-side
  }
  emit();
}

function subscribe(cb: () => void) {
  listeners.add(cb);
  window.addEventListener("storage", cb);
  return () => {
    listeners.delete(cb);
    window.removeEventListener("storage", cb);
  };
}

/** Reactive view of this browser's reservations for a list (stable snapshot per raw value). */
export function useStoredReservations(slug: string): StoredReservation[] {
  return useSyncExternalStore(
    subscribe,
    () => {
      let raw: string | null = null;
      try {
        raw = window.localStorage.getItem(key(slug));
      } catch {
        return EMPTY;
      }
      const cached = snapshots.get(slug);
      if (cached && cached.raw === raw) return cached.parsed;
      let parsed = EMPTY;
      try {
        parsed = raw ? (JSON.parse(raw) as StoredReservation[]) : EMPTY;
      } catch {
        parsed = EMPTY;
      }
      snapshots.set(slug, { raw, parsed });
      return parsed;
    },
    () => EMPTY,
  );
}

export function addReservation(slug: string, r: StoredReservation) {
  save(slug, [...loadReservations(slug).filter((x) => x.id !== r.id), r]);
}

export function removeReservation(slug: string, id: string) {
  save(
    slug,
    loadReservations(slug).filter((x) => x.id !== id),
  );
}

export function updateReservation(slug: string, id: string, patch: Partial<StoredReservation>) {
  save(
    slug,
    loadReservations(slug).map((x) => (x.id === id ? { ...x, ...patch } : x)),
  );
}
