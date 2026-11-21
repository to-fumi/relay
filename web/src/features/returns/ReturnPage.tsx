import { PackageCheck, QrCode } from "lucide-react";
import { useState } from "react";
import type { Handoff } from "../../shared/api";
import { Alert, Card, EmptyState } from "../../shared/ui";
import { LoanRow } from "./LoanRow";
import { ReturnQr } from "./ReturnQr";
import { useOutstanding } from "./useOutstanding";

/** 返す：自分が受け取った返却期限付きのもの。押すと返却QRを出す */
export function ReturnPage() {
  const { query, toReturn } = useOutstanding();
  const [selected, setSelected] = useState<Handoff | null>(null);

  if (selected) return <ReturnQr handoff={selected} onBack={() => setSelected(null)} />;

  if (query.isError) return <Alert>取得できませんでした</Alert>;
  if (query.isSuccess && toReturn.length === 0) {
    return (
      <Card>
        <EmptyState icon={PackageCheck} title="なし" />
      </Card>
    );
  }
  return (
    <Card className="divide-y divide-zinc-100 p-0 dark:divide-zinc-800">
      {toReturn.map((h) => (
        <button
          key={h.id}
          type="button"
          onClick={() => setSelected(h)}
          className="block w-full text-left hover:bg-zinc-50 dark:hover:bg-zinc-800/50"
          aria-label={`${h.item}の返却QRを表示`}
        >
          <LoanRow handoff={h} partner={h.from.displayName} trailing={<QrCode className="size-5" />} />
        </button>
      ))}
    </Card>
  );
}
