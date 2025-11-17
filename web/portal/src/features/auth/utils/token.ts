import type { Tokens } from "../types/user";

const KEY = "kleff_tokens";

export function saveTokens(tokens: Tokens) {
  try {
    localStorage.setItem(KEY, JSON.stringify(tokens));
  } catch (err) {
    console.error("Failed to save tokens:", err);
  }
}

export function loadTokens(): Tokens | null {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return null;
    return JSON.parse(raw) as Tokens;
  } catch (err) {
    console.error("Failed to load tokens:", err);
    return null;
  }
}

export function clearTokens() {
  try {
    localStorage.removeItem(KEY);
  } catch (err) {
    console.error("Failed to clear tokens:", err);
  }
}
