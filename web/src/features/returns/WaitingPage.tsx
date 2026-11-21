import { PackageCheck, ScanLine } from "lucide-react";
import { useState } from "react";
import { Alert, Card, EmptyState, Fab } from "../../shared/ui";
import { LoanRow } from "./LoanRow";
import { ReturnScan } from "./ReturnScan";
import { useOutstanding } from "./useOutstanding";

/** 待ち：自分が送った返却期限付きのもの。返ってきたら「読取」で相手の返却QRを読む */
export function WaitingPage() {
  const { query, waiting } = useOutstanding();
  const [scanning, setScanning] = useState(false);

  if (scanning) return <ReturnScan onBack={() => setScanning(false)} />;
  if (query.isError) return <Alert>取得できませんでした</Alert>;

  return (
    <div className="space-y-6">
      {query.isSuccess && waiting.length === 0 ? (
        <Card>
          <EmptyState icon={PackageCheck} title="なし" />
        </Card>
      ) : (
        <Card className="divide-y divide-zinc-100 p-0 dark:divide-zinc-800">
          {waiting.map((h) => (
            <LoanRow key={h.id} handoff={h} partner={h.to.displayName} />
          ))}
        </Card>
      )}
      {waiting.length > 0 && (
        <Fab label="返却QRを読み取る" onClick={() => setScanning(true)}>
          <ScanLine aria-hidden />
        </Fab>
      )}
    </div>
  );
}
