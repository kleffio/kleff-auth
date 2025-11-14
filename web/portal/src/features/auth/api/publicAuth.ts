const CLIENT_ID = "kleff-dashboard";
const REDIRECT_URI = "http://localhost:5173/dashboard";

export const publicAuth = {
  beginOAuth(tenant: string, provider: "google" | "github") {
    const params = new URLSearchParams({
      tenant,
      client_id: CLIENT_ID,
      redirect_uri: REDIRECT_URI,
    });

    return `/api/v1/auth/oauth/${provider}/start?${params.toString()}`;
  },
};
