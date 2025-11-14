import { useAuth } from "../features/auth/hooks/useAuth";

export const DashboardPage = () => {
  const { user, tenant } = useAuth();

  return (
    <div className="max-w-2xl mx-auto mt-10 space-y-4">
      <h2 className="text-2xl font-semibold">Dashboard</h2>

      <div className="rounded-xl border border-slate-800 p-4 space-y-2 bg-slate-900/40">
        <div>
          <div className="text-xs opacity-60 uppercase tracking-wide">
            Active Tenant
          </div>
          <div className="text-lg">{tenant ?? "unknown"}</div>
        </div>

        <div>
          <div className="text-xs opacity-60 uppercase tracking-wide">
            User
          </div>
          <div className="text-lg">
            {user?.name ?? user?.email ?? "Unknown user"}
          </div>
          <div className="text-xs opacity-70">{user?.email}</div>
        </div>
      </div>
    </div>
  );
};
