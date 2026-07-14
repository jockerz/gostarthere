import { PUBLIC_API_URL } from "$env/static/public";

export type SuccessBody = {
  success: boolean;
  message: string;
  data?: unknown;
};

export type UserData = {
  id: number;
  email: string;
  username: string;
  name: string;
  avatar: string;
  active: boolean;
};

export type AuthData = {
  access_token: string;
  refresh_token: string;
  user: {
    id: number;
    email: string;
    username: string;
    name: string;
    avatar: string;
    active: boolean;
  };
};

export type ErrorDetail = {
  location: string;
  message: string;
  value?: unknown;
};

export type ErrorModel = {
  status: number;
  title: string;
  detail: string;
  instance?: string;
  type?: string;
  errors?: ErrorDetail[] | null;
};

export function getToken(): string | null {
  return localStorage.getItem("token");
}

export function setToken(token: string): void {
  localStorage.setItem("token", token);
}

export function clearToken(): void {
  localStorage.removeItem("token");
}

export function getRefreshToken(): string | null {
  return localStorage.getItem("refresh_token");
}

export function setRefreshToken(token: string): void {
  localStorage.setItem("refresh_token", token);
}

export function clearTokens(): void {
  localStorage.removeItem("token");
  localStorage.removeItem("refresh_token");
}

type ApiOptions = {
  method?: string;
  json?: Record<string, unknown>;
  formData?: FormData;
};

export async function apiCall<T = SuccessBody>(
  endpoint: string,
  { method = "GET", json, formData }: ApiOptions = {},
): Promise<T> {
  const headers: Record<string, string> = {};
  if (json) {
    headers["Content-Type"] = "application/json";
  }
  const token = getToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  console.log(`API call to ${PUBLIC_API_URL}${endpoint}`, json ?? "formData");
  const response = await fetch(`${PUBLIC_API_URL}${endpoint}`, {
    method,
    headers,
    body: json ? JSON.stringify(json) : formData ?? undefined,
  });

  const data = await response.json();
  console.log("api response");
  console.log(data);

  if (!response.ok) {
    if (response.status == 401) {
      clearToken();
    }
    throw data as ErrorModel;
  }

  return data as T;
}

export const authApi = {
  login: (email: string, password: string, remember: boolean) =>
    apiCall<SuccessBody>("/v1/auth/login", {
      method: "POST",
      json: { email, password, remember },
    }),

  register: (
    name: string,
    username: string,
    email: string,
    password: string,
    password_validate: string,
  ) =>
    apiCall<SuccessBody>("/v1/auth/register", {
      method: "POST",
      json: { name, username, email, password, password_validate },
    }),

  activate: (token: string) =>
    apiCall<SuccessBody>(`/v1/auth/activate/${token}`, {
      method: "POST",
    }),

  forgotPassword: (email: string) =>
    apiCall<SuccessBody>("/v1/auth/forgot-password", {
      method: "POST",
      json: { email },
    }),

  resetPassword: (token: string, password: string) =>
    apiCall<SuccessBody>("/v1/auth/reset-password", {
      method: "POST",
      json: { token, password },
    }),

  resendActivation: (email: string) =>
    apiCall<SuccessBody>("/v1/auth/resend-activation", {
      method: "POST",
      json: { email },
    }),

  logout: (token: string) =>
    apiCall<SuccessBody>("/v1/auth/logout", {
      method: "POST",
      json: { token },
    }),

  refresh: (refreshToken: string) =>
    apiCall<SuccessBody>("/v1/auth/refresh", {
      method: "POST",
      json: { refresh_token: refreshToken },
    }),

  changePassword: (currentPassword: string, newPassword: string) =>
    apiCall<SuccessBody>("/v1/auth/change-password", {
      method: "PUT",
      json: { current_password: currentPassword, new_password: newPassword },
    }),

  changeEmail: (currentPassword: string, newEmail: string) =>
    apiCall<SuccessBody>("/v1/auth/update-email", {
      method: "POST",
      json: { password: currentPassword, email: newEmail },
    }),

  confirmEmail: (token: string) =>
    apiCall<SuccessBody>(`/v1/auth/update-email/${token}/confirm`, {
      method: "POST",
      // json: { token },
    }),

  oauthAuthorize: (provider: string, codeChallenge?: string) =>
    apiCall<{ url: string; state: string }>(
      `/v1/auth/oauth/${provider}/authorize${codeChallenge ? `?code_challenge=${codeChallenge}` : ""}`,
    ),

  oauthCallback: (
    provider: string,
    code: string,
    state: string,
    codeVerifier?: string,
  ) =>
    apiCall<{
      success: boolean;
      message: string;
      data: {
        access_token: string;
        refresh_token: string;
        user: UserData;
        is_new_user: boolean;
      };
    }>(`/v1/auth/oauth/${provider}/callback`, {
      method: "POST",
      json: { code, state, code_verifier: codeVerifier },
    }),

  setPassword: (password: string) =>
    apiCall<SuccessBody>("/v1/auth/set-password", {
      method: "POST",
      json: { password },
    }),

  getOAuthProviderData: () =>
    apiCall<SuccessBody>("/v1/auth/auth_data"),
};

export const profileApi = {
  get: () => apiCall<SuccessBody>("/v1/profile"),

  update: (data: { name?: string; username?: string; avatar?: string }) =>
    apiCall<SuccessBody>("/v1/profile", {
      method: "PUT",
      json: data,
    }),

  uploadAvatar: (file: File) => {
    const formData = new FormData();
    formData.append("avatar", file);
    return apiCall<SuccessBody>("/v1/profile/avatar", {
      method: "POST",
      formData,
    });
  },
};

export const checkApi = {
  health: () => apiCall<SuccessBody>("/v1/check"),
};
