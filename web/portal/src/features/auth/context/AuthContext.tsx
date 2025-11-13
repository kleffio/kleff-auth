import React, { createContext, useContext, useEffect, useState } from "react";
import { getMe, login, logout } from "../api/privateAuth";
import type { User } from "../types/user";

type AuthContextValue = {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export const useAuth = () => { /* ... */ };
