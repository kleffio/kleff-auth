export type User = {
  id: string;
  email: string;
  username: string;
  name?: string | null;
};

export type Tokens = {
  access_token: string;
  refresh_token: string;
};

export type TenantInfo = {
  slug: string;
  name: string;
};
