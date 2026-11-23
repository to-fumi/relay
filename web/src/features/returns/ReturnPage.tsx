import { useMutation, useQueryClient } from "@tanstack/react-query";
import { PackageCheck, QrCode, Send } from "lucide-react";
import { useState } from "react";
import {
  type Handoff,
  listMyHandoffsQueryKey,
  listMyOutstandingHandoffsQueryKey,
  requestReturnsMutation,
} from "../../shared/api";
import { problemStatus } from "../../shared/lib";
import { Alert, Card, EmptyState, Fab, IconButton } from "../../shared/ui";
import { LoanRow } from "./LoanRow";
import { ReturnQr } from "./ReturnQr";
import { useOutstanding } from "./useOutstanding";
import { useSelection } from "./useSelection";

const errorMessages: Record<number, string> = {
  404: "見つかりません",
  409: "返却済み、または送信済みのものがあります",
};

/**
 * 返す：自分が受け取った返却期限付きのもの。
 * 対面なら右の QR で返却QRを出す。対面でないなら選んでまとめて送り、貸した人の承認を待つ。
 */
export function ReturnPage() {
  const queryClient = useQueryClient();
  const { query, toReturn } = useOutstanding();
  const [showing, setShowing] = useState<Handoff | null>(null);
  const selection = useSelection(toReturn.filter((h) => !h.returnRequestedAt).map((h) => h.id));
  const send = useMutation({
    ...requestReturnsMutation(),
    onSuccess: () => {
      selection.clear();
      return Promise.all([
        queryClient.invalidateQueries({ queryKey: listMyOutstandingHandoffsQueryKey() }),
        queryClient.invalidateQueries({ queryKey: listMyHandoffsQueryKey() }),
      ]);
    },
  });

  if (showing) return <ReturnQr handoff={showing} onBack={() => setShowing(null)} />;

  if (query.isError) return <Alert>取得できませんでした</Alert>;
  if (query.isSuccess && toReturn.length === 0) {
    return (
      <Card>
        <EmptyState icon={PackageCheck} title="なし" />
      </Card>
    );
  }
  const status = problemStatus(send.error);
  return (
    <div className="space-y-6">
      <Card className="divide-y divide-zinc-100 p-0 dark:divide-zinc-800">
        {toReturn.map((h) => (
          <LoanRow
            key={h.id}
            handoff={h}
            side="borrower"
            partner={h.from.displayName}
            selectable
            select={
              h.returnRequestedAt
                ? undefined
                : { checked: selection.isSelected(h.id), onChange: (on) => selection.toggle(h.id, on) }
            }
            action={
              <IconButton label={`${h.item}の返却QRを表示`} variant="ghost" onClick={() => setShowing(h)}>
                <QrCode className="size-5" aria-hidden />
              </IconButton>
            }
          />
        ))}
      </Card>
      {send.isError && <Alert>{(status && errorMessages[status]) ?? "送信できませんでした"}</Alert>}
      {selection.selected.length > 0 && (
        <Fab
          label="選んだものの返却を送信"
          disabled={send.isPending}
          onClick={() => send.mutate({ body: { handoffIds: selection.selected } })}
        >
          <Send aria-hidden />
        </Fab>
      )}
    </div>
  );
}
