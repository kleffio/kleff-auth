import api from "@/lib/axiosInstance";

const CLIENT_ID = "kleff-dashboard";
const REDIRECT_URI = "http://localhost:5173/";

export const publicAuth = {
  beginOAuth(tenant: string, provider: "google" | "github") {
    const params = new URLSearchParams({
      tenant,
      client_id: CLIENT_ID,
      redirect_uri: REDIRECT_URI,
    });

    return `/api/v1/auth/oauth/${provider}/start?${params.toString()}`;
  },

  register(input: {
    tenant: string;
    email: string;
    username: string;
    password: string;
    attrs?: Record<string, any>;
  }) {
    return api.post("/auth/signup", {
      tenant: input.tenant,
      email: input.email,
      username: input.username,
      password: input.password,
      attrs: input.attrs ?? {},
    });
  }
};
