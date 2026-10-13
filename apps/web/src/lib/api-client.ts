import type { ApiError } from "@/types/customer";

const API_BASE_URL = "http://localhost:8080";

// API が失敗したときに投げるエラー。ステータスとエラー本文を持つ
export class ApiRequestError extends Error {
  constructor(
    readonly status: number,
    readonly error: ApiError,
  ) {
    super(error.message);
    this.name = "ApiRequestError";
  }
}

type RequestOptions = {
  method?: "GET" | "POST" | "DELETE";
  body?: unknown;
};

// バックエンドを呼ぶ共通処理。Cookie（ログイン状態）を送るため credentials: "include" を付ける
export async function apiRequest<T>(path: string, { method = "GET", body }: RequestOptions = {}): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    method,
    credentials: "include",
    headers: body === undefined ? undefined : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  if (!res.ok) {
    throw new ApiRequestError(res.status, await readApiError(res));
  }
  // 204 は本文がない
  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

// エラー本文を読む。読めないときは汎用のエラーにする
async function readApiError(res: Response): Promise<ApiError> {
  try {
    const body = (await res.json()) as { error: ApiError };
    return body.error;
  } catch {
    return { code: "internal_error", message: "時間をおいてもう一度お試しください", details: [] };
  }
}
