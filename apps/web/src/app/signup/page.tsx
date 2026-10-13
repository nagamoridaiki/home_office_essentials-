"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { ApiRequestError } from "@/lib/api-client";
import { getMe, signUp } from "@/lib/customer-api";

type FieldErrors = { email?: string; password?: string };

export default function SignUpPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);

  // ログイン済みならアカウント作成は不要なので、自分の情報へ移す
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
    setFieldErrors({});
    setFormError(null);

    try {
      await signUp(email, password);
      router.push("/mypage");
    } catch (err: unknown) {
      showError(err);
      setSubmitting(false);
    }
  };

  const showError = (err: unknown) => {
    if (err instanceof ApiRequestError && err.status === 422) {
      const errors: FieldErrors = {};
      for (const d of err.error.details) {
        if (d.field === "email" || d.field === "password") {
          errors[d.field] = d.message;
        }
      }
      setFieldErrors(errors);
      return;
    }
    if (err instanceof ApiRequestError && err.status === 409) {
      setFormError("このメールアドレスはすでに使われています");
      return;
    }
    setFormError("アカウントを作成できませんでした。時間をおいてもう一度お試しください");
  };

  return (
    <main className="max-w-md w-full mx-auto mt-12 p-6 bg-white rounded-xl shadow-md border border-gray-100">
      <h1 className="text-2xl font-bold text-gray-800 mb-6">アカウントを作成</h1>

      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        <Field label="メールアドレス" error={fieldErrors.email}>
          <input
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </Field>
        <Field label="パスワード（8 文字以上）" error={fieldErrors.password}>
          <input
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </Field>

        {formError && <p className="text-sm text-red-600">{formError}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-semibold rounded-lg transition"
        >
          {submitting ? "作成中..." : "アカウントを作成"}
        </button>
      </form>
    </main>
  );
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="block mb-1 text-sm text-gray-700">{label}</span>
      {children}
      {error && <span className="block mt-1 text-sm text-red-600">{error}</span>}
    </label>
  );
}
