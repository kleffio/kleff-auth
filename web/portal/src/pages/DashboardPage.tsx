import { useEffect, useRef, useState } from "react";
import { useAuth } from "../features/auth/hooks/useAuth";

export const DashboardPage = () => {
  const { user, tenant, logout, tokens, refreshTokens, authLoading } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const [showTokens, setShowTokens] = useState(false);

  const profileRef = useRef<HTMLDivElement | null>(null);

  const displayName =
    user?.name || user?.username || user?.email?.split("@")[0] || "there";

  const avatarInitial =
    user?.name?.[0]?.toUpperCase() ||
    user?.username?.[0]?.toUpperCase() ||
    user?.email?.[0]?.toUpperCase() ||
    "?";

  const tenantLabel = tenant ?? "No active tenant";
  const shortTenantId =
    tenant && tenant.length > 16
      ? `${tenant.slice(0, 8)}…${tenant.slice(-4)}`
      : tenant;

  const isSignedIn = Boolean(user);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handler = (event: MouseEvent) => {
      if (
        profileRef.current &&
        !profileRef.current.contains(event.target as Node)
      ) {
        setMenuOpen(false);
      }
    };

    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <div className="max-w-5xl mx-auto py-10 space-y-8">
      {/* Header */}
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm text-slate-400">Welcome back,</p>
          <h1 className="text-3xl font-semibold tracking-tight">
            {displayName}
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            Here’s an overview of your account and tenant.
          </p>
        </div>

        {/* Profile card + dropdown */}
        <div
          className="relative flex items-center justify-end"   // ⬅ add `relative` here
          ref={profileRef}
        >
          <button
            type="button"
            onClick={() => setMenuOpen((o) => !o)}
            className="flex items-center gap-3 rounded-full border border-slate-800 bg-slate-900/60 px-3 py-2 text-left shadow-sm hover:border-slate-600 hover:bg-slate-900 transition-colors"
          >
            <div className="flex h-9 w-9 items-center justify-center rounded-full bg-slate-800 text-xs font-semibold">
              {avatarInitial}
            </div>
            <div className="flex flex-col">
              <span className="text-sm font-medium text-slate-100">
                {user?.email ?? "Not signed in"}
              </span>
              {user?.id && (
                <span className="mt-0.5 inline-flex items-center gap-1 text-[10px] font-mono text-slate-400">
                  <span className="rounded-md bg-slate-800 px-1 py-0.5 text-[9px] uppercase tracking-wide text-slate-300">
                    ID
                  </span>
                  <span>{`${user.id.slice(0, 8)}…`}</span>
                </span>
              )}
            </div>
            <span className="ml-1 text-xs text-slate-500">▾</span>
          </button>

          {menuOpen && (
            <div className="absolute right-0 top-full mt-2 w-48 rounded-xl border border-slate-800 bg-slate-900/95 py-1 text-sm shadow-lg">
              <div className="px-3 py-2 border-b border-slate-800">
                <p className="text-xs text-slate-500">Signed in as</p>
                <p className="truncate text-slate-100 text-xs">
                  {user?.email ?? "Unknown"}
                </p>
              </div>

              {isSignedIn && (
                <button
                  type="button"
                  onClick={() => {
                    setMenuOpen(false);
                    logout();
                  }}
                  className="flex w-full items-center px-3 py-2 text-left text-slate-200 hover:bg-slate-800/90"
                >
                  Logout
                </button>
              )}
            </div>
          )}
        </div>
      </header>


      {/* Main grid */}
      <section className="grid gap-6 md:grid-cols-3">
        {/* Tenant card */}
        <div className="md:col-span-2 rounded-2xl border border-slate-800 bg-slate-900/40 p-5 shadow-sm">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                Active Tenant
              </p>
              <p className="mt-1 text-xl font-semibold text-slate-50">
                {tenantLabel}
              </p>
              {tenant && (
                <p className="mt-1 text-xs text-slate-500 font-mono">
                  Tenant ID: {shortTenantId}
                </p>
              )}
            </div>

            <div className="text-right">
              <span
                className={`inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium ${
                  tenant
                    ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30"
                    : "bg-amber-500/10 text-amber-400 border border-amber-500/30"
                }`}
              >
                <span
                  className={`mr-1 h-1.5 w-1.5 rounded-full ${
                    tenant ? "bg-emerald-400" : "bg-amber-400"
                  }`}
                />
                {tenant ? "Tenant linked" : "No tenant selected"}
              </span>
            </div>
          </div>

          <div className="mt-4 grid gap-4 sm:grid-cols-2 text-sm text-slate-400">
            <div>
              <p className="text-xs uppercase tracking-wide text-slate-500">
                Environment
              </p>
              <p className="mt-1">Local development</p>
            </div>
            <div>
              <p className="text-xs uppercase tracking-wide text-slate-500">
                Dashboard URL
              </p>
              <p className="mt-1 break-all text-slate-300">
                http://localhost:5173/
              </p>
            </div>
          </div>
        </div>

        {/* Account card */}
        <div className="rounded-2xl border border-slate-800 bg-slate-900/40 p-5 shadow-sm space-y-3">
          <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
            Account
          </p>
          <div className="space-y-2 text-sm">
            <div>
              <p className="text-xs text-slate-500">Name</p>
              <p className="text-slate-100">
                {user?.name ?? "Not provided"}
              </p>
            </div>
            <div>
              <p className="text-xs text-slate-500">Username</p>
              <p className="text-slate-100">
                {user?.username ?? "Not set"}
              </p>
            </div>
            <div>
              <p className="text-xs text-slate-500">Email</p>
              <p className="text-slate-100 break-all">
                {user?.email ?? "Unknown"}
              </p>
            </div>
          </div>
        </div>

        {/* Session card */}
        <div className="md:col-span-3 rounded-2xl border border-slate-800 bg-slate-900/40 p-5 shadow-sm">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
                Session
              </p>
              <p className="mt-1 text-sm text-slate-300">
                {isSignedIn
                  ? "You are currently signed in and can access tenant resources."
                  : "You are not signed in."}
              </p>
            </div>
            <div>
              <span
                className={`inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium ${
                  isSignedIn
                    ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30"
                    : "bg-rose-500/10 text-rose-400 border border-rose-500/30"
                }`}
              >
                <span
                  className={`mr-1 h-1.5 w-1.5 rounded-full ${
                    isSignedIn ? "bg-emerald-400" : "bg-rose-400"
                  }`}
                />
                {isSignedIn ? "Active session" : "Signed out"}
              </span>
            </div>
          </div>
        </div>

        {/* Tokens debug card */}
        <div className="md:col-span-3 rounded-2xl border border-slate-800 bg-slate-900/40 p-5 shadow-sm space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm uppercase tracking-wide text-slate-500">
              Tokens
            </h2>

            <button
              type="button"
              onClick={() => setShowTokens((s) => !s)}
              className="text-xs text-slate-400 hover:text-slate-200 transition"
            >
              {showTokens ? "Hide" : "Show"}
            </button>
          </div>

          {showTokens && (
            <div className="space-y-4 text-xs">
              <div>
                <p className="mb-1 text-slate-500">Access Token</p>
                <pre className="max-h-40 overflow-auto rounded-xl bg-slate-800/60 p-2 font-mono text-[11px] text-slate-200">
{tokens?.access_token ?? "No access token"}
                </pre>
              </div>

              <div>
                <p className="mb-1 text-slate-500">Refresh Token</p>
                <pre className="max-h-40 overflow-auto rounded-xl bg-slate-800/60 p-2 font-mono text-[11px] text-slate-200">
{tokens?.refresh_token ?? "No refresh token"}
                </pre>
              </div>

              <button
                type="button"
                onClick={refreshTokens}
                disabled={authLoading}
                className="rounded-lg border border-slate-700 bg-slate-800/60 px-3 py-2 text-sm text-slate-200 hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60 transition"
              >
                {authLoading ? "Refreshing…" : "Refresh Tokens"}
              </button>
            </div>
          )}
        </div>

      </section>
    </div>
  );
};
