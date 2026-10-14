"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { ApiRequestError } from "@/lib/api-client";
import { getMe, login } from "@/lib/customer-api";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  // ログイン済みならログインは不要なので、自分の情報へ移す
  useEffect(() => {
    getMe()
      .then(() => router.replace("/mypage"))
      .catch(() => {});
  }, [router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    // 送信中は二重に送らない
    if (submitting) return;
    setSubmitting(true);
    setFormError(null);

    try {
      await login(email, password);
      router.push("/mypage");
    } catch (err: unknown) {
      // 401 は API の文言をそのまま出す（未登録かパスワード違いかは区別しない）
      if (err instanceof ApiRequestError && err.status === 401) {
        setFormError(err.error.message);
      } else {
        setFormError("ログインできませんでした。時間をおいてもう一度お試しください");
      }
      setSubmitting(false);
    }
  };

  return (
    <main className="max-w-md w-full mx-auto mt-12 p-6 bg-white rounded-xl shadow-md border border-gray-100">
      <h1 className="text-2xl font-bold text-gray-800 mb-6">ログイン</h1>

      <form onSubmit={handleSubmit} className="space-y-4">
        <label className="block">
          <span className="block mb-1 text-sm text-gray-700">メールアドレス</span>
          <input
            type="email"
            autoComplete="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </label>
        <label className="block">
          <span className="block mb-1 text-sm text-gray-700">パスワード</span>
          <input
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </label>

        {formError && <p className="text-sm text-red-600">{formError}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-semibold rounded-lg transition"
        >
          {submitting ? "ログイン中..." : "ログイン"}
        </button>
      </form>

      <p className="mt-6 text-sm text-center text-gray-600">
        アカウントをお持ちでない方は
        <Link href="/signup" className="text-blue-600 hover:underline">
          アカウントを作成
        </Link>
      </p>
    </main>
  );
}
