import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CalendarClock, CircleCheck, CornerDownLeft, Package, ScanLine, User } from "lucide-react";
import {
  type Handoff,
  listMyHandoffsQueryKey,
  listMyOutstandingHandoffsQueryKey,
  receiveHandoffMutation,
  returnHandoffMutation,
} from "../../shared/api";
import { formatTime, problemStatus, problemType } from "../../shared/lib";
import { Alert, Card, DueBadge, Fab, IconRow, IconRows } from "../../shared/ui";
import { QrScanner } from "../../shared/ui/QrScanner";

/** 受け渡し用の API に返却QRを渡したときの Problem の type（そのときは返却として記録し直す） */
const returnTokenProblem = "urn:relay:problem:return-token";

const receiveErrors: Record<number, string> = {
  400: "無効、または期限切れのQRです",
  409: "読み取り済みのQRです",
  422: "自分のQRは読み取れません",
};

const returnErrors: Record<number, string> = {
  400: "無効、または期限切れのQRです",
  403: "あなたが送ったものではありません",
  404: "見つかりません",
  409: "返却済みです",
};

/** 受け取る：相手の受け渡しQRを読み取る。返却QRを読んだときは、返却として記録する */
export function ReceivePage() {
  const queryClient = useQueryClient();
  const invalidate = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: listMyHandoffsQueryKey() }),
      queryClient.invalidateQueries({ queryKey: listMyOutstandingHandoffsQueryKey() }),
    ]);
  const ret = useMutation({ ...returnHandoffMutation(), onSuccess: invalidate });
  const receive = useMutation({
    ...receiveHandoffMutation(),
    onSuccess: invalidate,
    // 返却QRだったときは、同じトークンで返却を記録する
    onError: (error, variables) => {
      if (problemType(error) === returnTokenProblem) ret.mutate({ body: variables.body });
    },
  });

  const pending = receive.isPending || ret.isPending;
  const reset = () => {
    receive.reset();
    ret.reset();
  };
  const onScan = (token: string) => {
    if (pending || receive.isSuccess || ret.isSuccess) return;
    reset();
    receive.mutate({ body: { token } });
  };

  if (receive.data) return <Done handoff={receive.data} kind="received" onNext={reset} />;
  if (ret.data) return <Done handoff={ret.data} kind="returned" onNext={reset} />;

  const isReturn = problemType(receive.error) === returnTokenProblem;
  const errorMessage = isReturn
    ? ret.isError && ((returnErrors[problemStatus(ret.error) ?? 0] ?? "記録できませんでした") as string)
    : receive.isError && (receiveErrors[problemStatus(receive.error) ?? 0] ?? "記録できませんでした");

  return (
    <section>
      <QrScanner onScan={onScan} paused={pending} />
      {pending && <p className="mt-6 text-center text-sm text-zinc-500">…</p>}
      {errorMessage && (
        <div className="mt-6">
          <Alert>{errorMessage}</Alert>
          <Fab label="もう一度読み取る" onClick={reset}>
            <ScanLine aria-hidden />
          </Fab>
        </div>
      )}
    </section>
  );
}

/** 記録できたとき。受け取り：左上に ✓、返却：左上に ↩ */
function Done({ handoff: h, kind, onNext }: { handoff: Handoff; kind: "received" | "returned"; onNext: () => void }) {
  const returned = kind === "returned";
  const Icon = returned ? CornerDownLeft : CircleCheck;
  const at = returned ? h.returnedAt : h.createdAt;
  return (
    <section>
      <Card className="space-y-6">
        <div className="flex items-center justify-between">
          <div
            className={
              returned
                ? "flex items-center gap-2 text-amber-600 dark:text-amber-400"
                : "flex items-center gap-2 text-emerald-600 dark:text-emerald-400"
            }
          >
            <Icon className="size-8" aria-hidden />
            <span className="font-bold text-lg">{returned ? "返却" : "完了"}</span>
          </div>
          {at && <span className="text-sm text-zinc-500 tabular-nums dark:text-zinc-400">{formatTime(at)}</span>}
        </div>
        <IconRows>
          <IconRow icon={User} label={returned ? "返した人" : "送った人"}>
            {returned ? h.to.displayName : h.from.displayName}
          </IconRow>
          <IconRow icon={Package} label="品目と数量">
            {h.item} <span className="tabular-nums">×{h.qty}</span>
          </IconRow>
          {!returned && h.dueDate && (
            <IconRow icon={CalendarClock} label="返却期限">
              <DueBadge dueDate={h.dueDate} className="text-sm" />
            </IconRow>
          )}
        </IconRows>
      </Card>
      <Fab label="続けて読み取る" onClick={onNext}>
        <ScanLine aria-hidden />
      </Fab>
    </section>
  );
}
