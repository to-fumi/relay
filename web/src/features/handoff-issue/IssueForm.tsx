import { CalendarClock, Hash, Minus, Package, Plus, QrCode } from "lucide-react";
import { type FormEvent, type KeyboardEvent, type ReactNode, useRef } from "react";
import { cn } from "../../shared/lib";
import { Alert, Card, Fab, inputClass } from "../../shared/ui";
import { DueDateInput } from "./DueDateInput";

type Props = {
  item: string;
  qty: number;
  dueDate: string | null;
  onItemChange: (v: string) => void;
  onQtyChange: (v: number) => void;
  onDueDateChange: (v: string | null) => void;
  onSubmit: () => void;
  pending: boolean;
  failed: boolean;
};

export function IssueForm({
  item,
  qty,
  dueDate,
  onItemChange,
  onQtyChange,
  onDueDateChange,
  onSubmit,
  pending,
  failed,
}: Props) {
  const qtyRef = useRef<HTMLInputElement>(null);
  // Enter では発行しない（数量や期限も入れてから、右下のボタンで発行する）。
  // 日本語入力の変換を確定する Enter はそのまま通す
  const onEnter = (next: () => void) => (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key !== "Enter" || e.nativeEvent.isComposing) return;
    e.preventDefault();
    next();
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    onSubmit();
  };
  const setQty = (n: number) => onQtyChange(Math.max(1, Number.isFinite(n) ? Math.floor(n) : 1));

  return (
    <form onSubmit={handleSubmit} className="space-y-8">
      <Card className="divide-y divide-zinc-100 py-2 dark:divide-zinc-800">
        <Row icon={<Package />}>
          <input
            id="item"
            aria-label="品目"
            enterKeyHint="next"
            onKeyDown={onEnter(() => qtyRef.current?.select())}
            value={item}
            onChange={(e) => onItemChange(e.target.value)}
            placeholder="LANケーブル"
            maxLength={200}
            required
            className={inputClass}
          />
        </Row>

        <Row icon={<Hash />}>
          {/* 数量は桁が少ないので、右に寄せた小さなステッパーにする */}
          <div className="ml-auto flex w-fit items-center gap-2">
            <StepButton label="1つ減らす" onClick={() => setQty(qty - 1)} disabled={qty <= 1}>
              <Minus className="size-5" />
            </StepButton>
            <input
              ref={qtyRef}
              aria-label="数量"
              enterKeyHint="done"
              onKeyDown={onEnter(() => qtyRef.current?.blur())}
              type="number"
              inputMode="numeric"
              min={1}
              value={qty}
              onChange={(e) => setQty(Number(e.target.value))}
              className={cn(inputClass, "h-11 w-16 px-1 text-center font-semibold text-lg tabular-nums")}
              required
            />
            <StepButton label="1つ増やす" onClick={() => setQty(qty + 1)}>
              <Plus className="size-5" />
            </StepButton>
          </div>
        </Row>

        <Row icon={<CalendarClock />} label="返却期限" alignTop>
          <DueDateInput value={dueDate} onChange={onDueDateChange} />
        </Row>
      </Card>

      {failed && <Alert>発行できませんでした</Alert>}

      <Fab type="submit" label="QRを発行" disabled={pending || item.trim() === ""}>
        <QrCode aria-hidden />
      </Fab>
    </form>
  );
}

/** 左にアイコン、右に入力を置く 1 行 */
function Row({
  icon,
  label,
  alignTop,
  children,
}: {
  icon: ReactNode;
  label?: string;
  /** 2 段の入力のとき、アイコンを上の段にそろえる */
  alignTop?: boolean;
  children: ReactNode;
}) {
  return (
    // biome-ignore lint/a11y/useSemanticElements: 見た目の行。入力にはそれぞれ aria-label を付けている
    <div
      className={cn("flex min-h-18 gap-4 py-3", alignTop ? "items-start" : "items-center")}
      role="group"
      aria-label={label}
    >
      <span
        className={cn("shrink-0 text-zinc-400 dark:text-zinc-500 [&>svg]:size-5", alignTop && "pt-2.5")}
        aria-hidden
      >
        {icon}
      </span>
      <div className="min-w-0 flex-1">{children}</div>
    </div>
  );
}

function StepButton({
  label,
  children,
  ...props
}: {
  label: string;
  children: ReactNode;
  onClick: () => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      className="grid size-11 shrink-0 place-items-center rounded-full bg-zinc-100 text-zinc-700 hover:bg-zinc-200 disabled:opacity-40 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700"
      {...props}
    >
      {children}
    </button>
  );
}
