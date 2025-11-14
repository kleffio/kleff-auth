import type { Tokens } from "../types/user";

const KEY = "kleff_auth_tokens";

export const saveTokens = (tokens: Tokens) => {
  localStorage.setItem(KEY, JSON.stringify(tokens));
};

export const loadTokens = (): Tokens | null => {
  const raw = localStorage.getItem(KEY);
  if (!raw) return null;
  return JSON.parse(raw);
};

export const clearTokens = () => {
  localStorage.removeItem(KEY);
};
