import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

import { privateAuth } from "../api/privateAuth";
import { publicAuth } from "../api/publicAuth";

import { saveTokens, loadTokens, clearTokens } from "../utils/token";

import type { User, Tokens } from "../types/user";

type AuthContextType = {
  user: User | null;
  tenant: string | null;
  loading: boolean;
  authLoading: boolean;
  authError: string | null;
  loginWithProvider: (tenant: string, provider: "google" | "github") => void;
  loginWithCredentials: (identity: string, password: string) => Promise<void>;
  registerWithCredentials: (input: {
    name?: string;
    username: string;
    email: string;
    password: string;
  }) => Promise<void>;

  logout: () => void;
  tokens: Tokens | null;
  refreshTokens: () => Promise<void>;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [tenant, setTenant] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const [authLoading, setAuthLoading] = useState(false);
  const [authError, setAuthError] = useState<string | null>(null);

  const [tokens, setTokens] = useState<Tokens | null>(() => loadTokens());

  useEffect(() => {
    const init = async () => {
      try {
        const me = await privateAuth.me();

        if (me?.user) {
          setUser({
            id: me.user.id,
            email: me.user.email,
            username: me.user.username,
          });


          const tenantFromMe = (me.user as any).tenant_id ?? null;
          setTenant(tenantFromMe);
        }
      } catch {
      } finally {
        setLoading(false);
      }
    };

    void init();
  }, []);


  const loginWithProvider = (
    tenantSlug: string,
    provider: "google" | "github",
  ) => {
    const url = publicAuth.beginOAuth(tenantSlug, provider);
    window.location.href = url;
  };

  const loginWithCredentials = async (identity: string, password: string) => {
  setAuthLoading(true);
  setAuthError(null);

  try {
    const res = await privateAuth.signin(identity, password);
      const session = res.session;

      if (session?.access_token && session?.refresh_token) {
        const newTokens: Tokens = {
          access_token: session.access_token,
          refresh_token: session.refresh_token,
        };
        setTokens(newTokens);
        saveTokens(newTokens);
      }

      const me = await privateAuth.me();
      if (me?.user) {
        setUser({
          id: me.user.id,
          email: me.user.email,
          username: me.user.username,
        });

        const tenantFromMe = (me.user as any).tenant_id ?? null;
        setTenant(tenantFromMe);
      }
    } catch (err) {
      const message =
        (err as Error).message || "Failed to sign in. Please try again.";
      setAuthError(message);
    } finally {
      setAuthLoading(false);
    }
  };

  const refreshTokens = async () => {
    setAuthLoading(true);
    setAuthError(null);

    try {
      const res = await privateAuth.refresh();

      const session = (res as any).session ?? res;

      if (session?.access_token && session?.refresh_token) {
        const newTokens: Tokens = {
          access_token: session.access_token,
          refresh_token: session.refresh_token,
        };
        setTokens(newTokens);
        saveTokens(newTokens);
      }
    } catch (err) {
      const message =
        (err as Error).message || "Failed to refresh tokens. Please try again.";
      setAuthError(message);
    } finally {
      setAuthLoading(false);
    }
  };

  const registerWithCredentials = async (input: {
    name?: string;
    username: string;
    email: string;
    password: string;
  }) => {
    setAuthLoading(true);
    setAuthError(null);

    try {
      const res = await publicAuth.register({
        tenant: "kleff",
        email: input.email,
        username: input.username,
        password: input.password,
        attrs: {
          name: input.name,
          marketing_opt_in: true,
        },
      });

      if (res.data?.session) {
        const newTokens: Tokens = {
          access_token: res.data.session.access_token,
          refresh_token: res.data.session.refresh_token,
        };

        setTokens(newTokens);
        saveTokens(newTokens);

        const me = await privateAuth.me();
        if (me?.user) {
          setUser({
            id: me.user.id,
            email: me.user.email,
            username: me.user.username,
          });

          const tenantFromMe = me.user.tenant_id ?? null;
          setTenant(tenantFromMe);
        }
      }

    } catch (err: any) {
      const msg =
        err.response?.data?.error ??
        err.message ??
        "Failed to create an account.";

      setAuthError(msg);
      throw new Error(msg);
    } finally {
      setAuthLoading(false);
    }
  };

  const logout = () => {
    clearTokens();
    setTokens(null);
    setUser(null);
    setTenant(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        tenant,
        loading,
        authLoading,
        authError,
        loginWithProvider,
        loginWithCredentials,
        registerWithCredentials,
        logout,
        tokens,
        refreshTokens,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuthContext = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuthContext must be used within AuthProvider");
  return ctx;
};
