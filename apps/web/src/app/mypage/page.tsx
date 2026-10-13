"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { getMe, isUnauthenticated } from "@/lib/customer-api";
import type { Customer } from "@/types/customer";

export default function MyPage() {
  const router = useRouter();
  const [customer, setCustomer] = useState<Customer | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    getMe()
      .then(setCustomer)
      .catch((e: unknown) => {
        if (isUnauthenticated(e)) {
          router.replace("/login");
          return;
        }
        setFailed(true);
      });
  }, [router]);

  if (failed) {
    return <Main>読み込みに失敗しました。時間をおいてもう一度お試しください。</Main>;
  }
  if (!customer) {
    return <Main>読み込み中...</Main>;
  }

  return (
    <Main>
      <h1 className="text-2xl font-bold text-gray-800 mb-6">自分の情報</h1>
      <dl className="space-y-3">
        <Row label="メールアドレス" value={customer.email} />
        <Row label="氏名" value={joinName(customer.family_name, customer.given_name)} />
        <Row label="フリガナ" value={joinName(customer.family_name_kana, customer.given_name_kana)} />
        <Row label="電話番号" value={customer.phone_number} />
      </dl>
    </Main>
  );
}

// 姓と名のどちらかでも未登録なら null（未登録として表示する）
function joinName(family: string | null, given: string | null): string | null {
  if (family === null || given === null) {
    return null;
  }
  return `${family} ${given}`;
}

function Row({ label, value }: { label: string; value: string | null }) {
  return (
    <div className="flex gap-4">
      <dt className="w-32 text-gray-500">{label}</dt>
      <dd className={value === null ? "text-gray-400" : "text-gray-800"}>{value ?? "未登録"}</dd>
    </div>
  );
}

function Main({ children }: { children: React.ReactNode }) {
  return (
    <main className="max-w-md w-full mx-auto mt-12 p-6 bg-white rounded-xl shadow-md border border-gray-100">
      {children}
    </main>
  );
}
