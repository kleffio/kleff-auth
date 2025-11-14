import api from "@/lib/axiosInstance";
import type { User } from "../types/user";

export type MeResponse = {
  user: {
    id: string;
    email: string;
    username: string;
    tenant_id?: string;
  };
};

export type SignInResponse = {
  session: {
    access_token: string;
    refresh_token: string;
    expires_in: number;
    token_type: string;
  };
  user: User;
};

const DEFAULT_TENANT = "kleff";

export const privateAuth = {
  async me(): Promise<MeResponse> {
    const res = await api.get<MeResponse>("/auth/me");
    return res.data;
  },

  async refresh() {
    const res = await api.post("/auth/refresh");
    return res.data;
  },

  async signin(identity: string, password: string, tenant = DEFAULT_TENANT) {
    const res = await api.post<SignInResponse>("/auth/signin", {
      tenant,
      identifier: identity,
      password,
    });

    return res.data;
  },
};