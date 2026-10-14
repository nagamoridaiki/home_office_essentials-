import { ApiRequestError, apiRequest } from "@/lib/api-client";
import type { Customer } from "@/types/customer";

// ログインしていない（401）かどうか
export function isUnauthenticated(e: unknown): boolean {
  return e instanceof ApiRequestError && e.status === 401;
}

// 自分の情報を返す。ログインしていなければ 401 の ApiRequestError を投げる
export function getMe(): Promise<Customer> {
  return apiRequest<Customer>("/me");
}

// アカウントを作る。成功するとそのままログインした状態になる
export function signUp(email: string, password: string): Promise<Customer> {
  return apiRequest<Customer>("/customers", { method: "POST", body: { email, password } });
}

// ログインする。成功すると Cookie が発行され、自分の情報が返る
export function login(email: string, password: string): Promise<Customer> {
  return apiRequest<Customer>("/sessions", { method: "POST", body: { email, password } });
}

// ログアウトする。すでにログアウト済み（401）でも、目的は果たせているので成功として扱う
export async function logout(): Promise<void> {
  try {
    await apiRequest<void>("/sessions", { method: "DELETE" });
  } catch (e: unknown) {
    if (!isUnauthenticated(e)) throw e;
  }
}
