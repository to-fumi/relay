import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, CircleCheck, RotateCw, User } from "lucide-react";
import { useEffect } from "react";
import {
  type Handoff,
  issueReturnTokenMutation,
  listMyHandoffsQueryKey,
  listMyOutstandingHandoffsOptions,
} from "../../shared/api";
import { problemStatus } from "../../shared/lib";
import { Alert, Card, DueBadge, IconButton, TokenQr } from "../../shared/ui";

const errorMessages: Record<number, string> = {
  404: "見つかりません",
  409: "返却済みです",
};

/** 借りた側が返却QRを表示する。貸した側が読み取ると返却リストから消えるので、それを待って完了を表示する。 */
export function ReturnQr({ handoff: h, onBack }: { handoff: Handoff; onBack: () => void }) {
  const queryClient = useQueryClient();
  const issue = useMutation(issueReturnTokenMutation());
  const { mutate } = issue;

  useEffect(() => {
    mutate({ body: { handoffId: h.id } });
  }, [mutate, h.id]);

  // QR を表示している間は、返却されたかを確かめるために返却リストを取り直す
  const outstanding = useQuery({
    ...listMyOutstandingHandoffsOptions(),
    refetchInterval: issue.isSuccess ? 2000 : false,
  });
  const returned = outstanding.isSuccess && issue.isSuccess && !outstanding.data.some((o) => o.id === h.id);

  useEffect(() => {
    if (returned) queryClient.invalidateQueries({ queryKey: listMyHandoffsQueryKey() });
  }, [returned, queryClient]);

  const status = problemStatus(issue.error);

  return (
    <Card className="flex flex-col items-center gap-6">
      <div className="w-full space-y-1">
        <div className="flex items-baseline justify-between gap-3">
          <p className="min-w-0 truncate font-semibold text-xl">{h.item}</p>
          <span className="shrink-0 font-semibold text-xl tabular-nums">×{h.qty}</span>
        </div>
        <div className="flex items-center justify-between gap-3">
          <p className="flex min-w-0 items-center gap-1.5 text-zinc-500 dark:text-zinc-400">
            <User className="size-4 shrink-0" aria-label="返却先" />
            <span className="truncate">{h.from.displayName}</span>
          </p>
          {h.dueDate && <DueBadge dueDate={h.dueDate} />}
        </div>
      </div>

      {returned ? (
        <div className="flex flex-col items-center gap-2 py-16 text-emerald-600 dark:text-emerald-400">
          <CircleCheck className="size-20" aria-hidden />
          <span className="font-bold text-lg">完了</span>
        </div>
      ) : issue.data ? (
        <TokenQr token={issue.data.token} expiresAt={issue.data.expiresAt} title="返却用のQRコード" />
      ) : issue.isError ? (
        <Alert>{(status && errorMessages[status]) ?? "発行できませんでした"}</Alert>
      ) : (
        <div className="size-72" />
      )}

      <div className="flex w-full justify-between">
        <IconButton label="戻る" onClick={onBack}>
          <ArrowLeft className="size-5" aria-hidden />
        </IconButton>
        {!returned && (
          <IconButton
            label="再発行"
            variant="primary"
            onClick={() => mutate({ body: { handoffId: h.id } })}
            disabled={issue.isPending}
          >
            <RotateCw className="size-5" aria-hidden />
          </IconButton>
        )}
      </div>
    </Card>
  );
}
