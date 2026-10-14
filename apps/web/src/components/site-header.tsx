"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { getMe, logout } from "@/lib/customer-api";

// loading: 確認中 / in: ログイン中 / out: 未ログイン
type LoginState = "loading" | "in" | "out";

export function SiteHeader() {
  const router = useRouter();
  const pathname = usePathname();
  const [state, setState] = useState<LoginState>("loading");
  const [loggingOut, setLoggingOut] = useState(false);

  // ログイン・ログアウトで画面が移るたびに、ログイン状態を確かめ直す
  useEffect(() => {
    let ignore = false;
    getMe()
      .then(() => !ignore && setState("in"))
      .catch(() => !ignore && setState("out"));
    return () => {
      ignore = true;
    };
  }, [pathname]);

  const handleLogout = async () => {
    if (loggingOut) return;
    setLoggingOut(true);
    try {
      await logout();
      setState("out");
      router.push("/login");
    } catch {
      alert("ログアウトできませんでした。時間をおいてもう一度お試しください");
    } finally {
      setLoggingOut(false);
    }
  };

  return (
    <header className="w-full border-b border-gray-200 bg-white">
      <div className="max-w-4xl mx-auto px-4 h-14 flex items-center justify-between">
        <Link href="/" className="font-bold text-gray-800">
          Home Office Essentials
        </Link>
        <nav className="flex items-center gap-4 text-sm">
          {state === "in" && (
            <>
              <Link href="/mypage" className="text-gray-700 hover:underline">
                自分の情報
              </Link>
              <button
                type="button"
                onClick={handleLogout}
                disabled={loggingOut}
                className="px-3 py-1.5 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 disabled:text-gray-400"
              >
                {loggingOut ? "ログアウト中..." : "ログアウト"}
              </button>
            </>
          )}
          {state === "out" && (
            <>
              <Link href="/login" className="text-gray-700 hover:underline">
                ログイン
              </Link>
              <Link href="/signup" className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg">
                アカウント作成
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
