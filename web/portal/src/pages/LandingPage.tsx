import { type FormEvent, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/features/auth/hooks/useAuth";

export const LandingPage = () => {
  const {
    user,
    loginWithProvider,
    loginWithCredentials,
    authLoading,
    authError,
  } = useAuth();
  const navigate = useNavigate();

  const [identity, setIdentity] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    if (user) {
      navigate("/dashboard");
    }
  }, [user, navigate]);

  const handleOAuthLogin = (provider: "google" | "github") => {
    loginWithProvider("kleff", provider);
  };

  const handleCredentialsLogin = async (e: FormEvent) => {
    e.preventDefault();
    await loginWithCredentials(identity.trim(), password);
  };

  return (
    <div className="max-w-md mx-auto mt-20 space-y-6">
      <div>
        <h1 className="text-3xl font-semibold mb-2">Kleff Identity</h1>
        <p className="text-sm opacity-80">
          Multi-tenant identity provider for all Kleff services.
        </p>
      </div>

      <form onSubmit={handleCredentialsLogin} className="space-y-3">
        <div className="space-y-1">
          <label className="block text-sm font-medium opacity-80">
            Username or Email
          </label>
          <input
            type="text"
            value={identity}
            onChange={(e) => setIdentity(e.target.value)}
            className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
            autoComplete="username"
          />
        </div>

        <div className="space-y-1">
          <label className="block text-sm font-medium opacity-80">
            Password
          </label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
            autoComplete="current-password"
          />
        </div>

        {authError && (
          <p className="text-xs text-red-400 mt-1">{authError}</p>
        )}

        <button
          type="submit"
          disabled={authLoading}
          className="w-full px-4 py-2 rounded-md bg-slate-100 text-slate-900 text-sm font-medium hover:bg-white disabled:opacity-60"
        >
          {authLoading ? "Signing in..." : "Sign in"}
        </button>
      </form>

      <div className="space-y-3 pt-4 border-t border-slate-800">
        <p className="text-xs uppercase tracking-wide opacity-60">
          Or continue with
        </p>

        <button
          onClick={() => handleOAuthLogin("google")}
          className="w-full px-4 py-2 rounded-md border border-slate-700 hover:bg-slate-800 text-left"
        >
          Continue with Google
        </button>

        <button
          onClick={() => handleOAuthLogin("github")}
          className="w-full px-4 py-2 rounded-md border border-slate-700 hover:bg-slate-800 text-left"
        >
          Continue with GitHub
        </button>
      </div>
    </div>
  );
};
