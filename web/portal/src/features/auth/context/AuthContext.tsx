import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

import { privateAuth } from "../api/privateAuth";
import { publicAuth } from "../api/publicAuth";
import type { User } from "../types/user";

type AuthContextType = {
  user: User | null;
  tenant: string | null;
  loading: boolean;
  authLoading: boolean;
  authError: string | null;
  loginWithProvider: (tenant: string, provider: "google" | "github") => void;
  loginWithCredentials: (identity: string, password: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [tenant, setTenant] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const [authLoading, setAuthLoading] = useState(false);
  const [authError, setAuthError] = useState<string | null>(null);

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
      await privateAuth.signin(identity, password);

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


  const logout = () => {
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
        logout,
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
