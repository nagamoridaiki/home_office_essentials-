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
