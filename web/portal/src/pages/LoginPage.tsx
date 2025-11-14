import { type FormEvent, useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "@/features/auth/hooks/useAuth";

export const LoginPage = () => {
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
      navigate("/");
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
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-4xl grid gap-10 md:grid-cols-[1.1fr,1fr] items-center">
        <div className="space-y-4">
          <h1 className="text-4xl font-semibold tracking-tight">
            Kleff Identity
          </h1>
          <p className="text-sm text-slate-400 leading-relaxed">
            Secure multi-tenant identity provider for all Kleff services.
            Sign in to manage your applications, tenants, and account settings
            across the Kleff platform.
          </p>

          <div className="flex items-center gap-3 text-xs text-slate-500">
            <span className="h-px flex-1 bg-slate-800" />
            <span>Environment: local</span>
            <span className="h-px flex-1 bg-slate-800" />
          </div>
        </div>

        <div className="rounded-2xl border border-slate-800 bg-slate-950/70 p-6 shadow-xl shadow-black/40 space-y-5">
          <div>
            <h2 className="text-lg font-medium">Sign in</h2>
            <p className="text-xs text-slate-400 mt-1">
              Use your Kleff credentials or continue with a provider.
            </p>
          </div>

          <form onSubmit={handleCredentialsLogin} className="space-y-3">
            <div className="space-y-1">
              <label className="block text-xs font-medium text-slate-300">
                Username or Email
              </label>
              <input
                type="text"
                value={identity}
                onChange={(e) => setIdentity(e.target.value)}
                className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
                autoComplete="username"
                required
              />
            </div>

            <div className="space-y-1">
              <label className="block text-xs font-medium text-slate-300">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
                autoComplete="current-password"
                required
              />
            </div>

            {authError && (
              <p className="text-xs text-red-400 mt-1">{authError}</p>
            )}

            <button
              type="submit"
              disabled={authLoading}
              className="w-full px-4 py-2 rounded-md bg-slate-100 text-slate-900 text-sm font-medium hover:bg-white disabled:opacity-60 disabled:cursor-not-allowed"
            >
              {authLoading ? "Signing in..." : "Sign in"}
            </button>
          </form>

          <div className="space-y-3 pt-4 border-t border-slate-800">
            <p className="text-[11px] uppercase tracking-wide text-slate-500">
              Or continue with
            </p>

            <div className="flex flex-col gap-2">
              <button
                type="button"
                onClick={() => handleOAuthLogin("google")}
                className="w-full px-4 py-2 rounded-md border border-slate-700 hover:bg-slate-900 text-left text-sm"
              >
                Continue with Google
              </button>

              <button
                type="button"
                onClick={() => handleOAuthLogin("github")}
                className="w-full px-4 py-2 rounded-md border border-slate-700 hover:bg-slate-900 text-left text-sm"
              >
                Continue with GitHub
              </button>
            </div>
          </div>

          <div className="pt-2 text-xs text-slate-500 flex items-center justify-between">
            <span>
              Don&apos;t have an account?{" "}
              <Link
                to="/signup"
                className="text-slate-200 hover:text-white underline underline-offset-2"
              >
                Create one
              </Link>
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};
