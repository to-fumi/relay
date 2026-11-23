import { useMutation, useQueryClient } from "@tanstack/react-query";
import { PackageCheck, ScanLine } from "lucide-react";
import { useState } from "react";
import { confirmReturnsMutation, listMyHandoffsQueryKey, listMyOutstandingHandoffsQueryKey } from "../../shared/api";
import { problemStatus } from "../../shared/lib";
import { Alert, Card, EmptyState, Fab } from "../../shared/ui";
import { LoanRow } from "./LoanRow";
import { ReturnScan } from "./ReturnScan";
import { useOutstanding } from "./useOutstanding";
import { useSelection } from "./useSelection";

const errorMessages: Record<number, string> = {
  404: "見つかりません",
  409: "返却済み、または承認待ちでないものがあります",
};

/**
 * 待ち：自分が送った返却期限付きのもの。
 * 対面なら「読取」で相手の返却QRを読む。対面でない返却が届いたものは、選んでまとめて承認する。
 */
export function WaitingPage() {
  const queryClient = useQueryClient();
  const { query, waiting } = useOutstanding();
  const [scanning, setScanning] = useState(false);
  const requested = waiting.filter((h) => h.returnRequestedAt);
  const selection = useSelection(requested.map((h) => h.id));
  const confirm = useMutation({
    ...confirmReturnsMutation(),
    onSuccess: () => {
      selection.clear();
      return Promise.all([
        queryClient.invalidateQueries({ queryKey: listMyOutstandingHandoffsQueryKey() }),
        queryClient.invalidateQueries({ queryKey: listMyHandoffsQueryKey() }),
      ]);
    },
  });

  if (scanning) return <ReturnScan onBack={() => setScanning(false)} />;
  if (query.isError) return <Alert>取得できませんでした</Alert>;

  // 承認待ちのものを上にまとめる
  const rows = [...requested, ...waiting.filter((h) => !h.returnRequestedAt)];
  const status = problemStatus(confirm.error);
  return (
    <div className="space-y-6">
      {query.isSuccess && waiting.length === 0 ? (
        <Card>
          <EmptyState icon={PackageCheck} title="なし" />
        </Card>
      ) : (
        <Card className="divide-y divide-zinc-100 p-0 dark:divide-zinc-800">
          {rows.map((h) => (
            <LoanRow
              key={h.id}
              handoff={h}
              partner={h.to.displayName}
              selectable={requested.length > 0}
              select={
                h.returnRequestedAt
                  ? { checked: selection.isSelected(h.id), onChange: (on) => selection.toggle(h.id, on) }
                  : undefined
              }
            />
          ))}
        </Card>
      )}
      {confirm.isError && <Alert>{(status && errorMessages[status]) ?? "記録できませんでした"}</Alert>}
      {selection.selected.length > 0 ? (
        <Fab
          label="選んだものの返却を承認"
          disabled={confirm.isPending}
          onClick={() => confirm.mutate({ body: { handoffIds: selection.selected } })}
        >
          <PackageCheck aria-hidden />
        </Fab>
      ) : (
        waiting.length > 0 && (
          <Fab label="返却QRを読み取る" onClick={() => setScanning(true)}>
            <ScanLine aria-hidden />
          </Fab>
        )
      )}
    </div>
  );
}
