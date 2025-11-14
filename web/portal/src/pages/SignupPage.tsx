import { type FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "@/features/auth/hooks/useAuth";

export const SignupPage = () => {
  const { registerWithCredentials, authLoading, authError } = useAuth();
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [passwordConfirm, setPasswordConfirm] = useState("");

  const [localError, setLocalError] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLocalError(null);

    if (password !== passwordConfirm) {
      setLocalError("Passwords do not match.");
      return;
    }

    try {
      await registerWithCredentials({
        name: name.trim(),
        username: username.trim(),
        email: email.trim(),
        password,
      });

      navigate("/login"); // or navigate("/dashboard") if auto-login
    } catch (err) {
      setLocalError(
        (err as Error)?.message || "Failed to create account. Please try again."
      );
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-4xl grid gap-10 md:grid-cols-[1.1fr,1fr] items-center">
        <div className="space-y-4">
          <h1 className="text-4xl font-semibold tracking-tight">
            Create your Kleff account
          </h1>
          <p className="text-sm text-slate-400 leading-relaxed">
            One account to access all Kleff services. Manage your tenants,
            deployments, and applications from a single identity provider.
          </p>
        </div>

        <div className="rounded-2xl border border-slate-800 bg-slate-950/70 p-6 shadow-xl shadow-black/40 space-y-5">
          <div>
            <h2 className="text-lg font-medium">Sign up</h2>
            <p className="text-xs text-slate-400 mt-1">
              Fill in your details to create a new account.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="space-y-1">
              <label className="block text-xs font-medium text-slate-300">
                Full name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
              />
            </div>

            <div className="space-y-1">
              <label className="block text-xs font-medium text-slate-300">
                Username
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
                required
              />
            </div>

            <div className="space-y-1">
              <label className="block text-xs font-medium text-slate-300">
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
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
                required
              />
            </div>

            <div className="space-y-1">
              <label className="block text-xs font-medium text-slate-300">
                Confirm password
              </label>
              <input
                type="password"
                value={passwordConfirm}
                onChange={(e) => setPasswordConfirm(e.target.value)}
                className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
                required
              />
            </div>

            {(localError || authError) && (
              <p className="text-xs text-red-400 mt-1">
                {localError || authError}
              </p>
            )}

            <button
              type="submit"
              disabled={authLoading}
              className="w-full px-4 py-2 rounded-md bg-slate-100 text-slate-900 text-sm font-medium hover:bg-white disabled:opacity-60 disabled:cursor-not-allowed"
            >
              {authLoading ? "Creating account..." : "Create account"}
            </button>
          </form>

          <div className="pt-2 text-xs text-slate-500">
            Already have an account?{" "}
            <Link
              to="/login"
              className="text-slate-200 hover:text-white underline underline-offset-2"
            >
              Sign in
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};
