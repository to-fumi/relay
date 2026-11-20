import { Minus, Plus, QrCode } from "lucide-react";
import type { FormEvent } from "react";
import { cn } from "../../shared/lib";
import { Alert, Button, Card, Field, inputClass } from "../../shared/ui";

type Props = {
  item: string;
  qty: number;
  onItemChange: (v: string) => void;
  onQtyChange: (v: number) => void;
  onSubmit: () => void;
  pending: boolean;
  failed: boolean;
};

export function IssueForm({ item, qty, onItemChange, onQtyChange, onSubmit, pending, failed }: Props) {
  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    onSubmit();
  };
  const setQty = (n: number) => onQtyChange(Math.max(1, Number.isFinite(n) ? Math.floor(n) : 1));

  return (
    <Card>
      <form onSubmit={handleSubmit} className="space-y-5">
        <Field
          id="item"
          label="品目"
          value={item}
          onChange={(e) => onItemChange(e.target.value)}
          placeholder="例：LANケーブル"
          maxLength={200}
          required
        />

        <div>
          <span className="mb-1.5 block font-medium text-sm text-zinc-700 dark:text-zinc-300">数量</span>
          <div className="flex items-center gap-2">
            <StepButton label="1つ減らす" onClick={() => setQty(qty - 1)} disabled={qty <= 1}>
              <Minus className="size-5" />
            </StepButton>
            <input
              aria-label="数量"
              type="number"
              inputMode="numeric"
              min={1}
              value={qty}
              onChange={(e) => setQty(Number(e.target.value))}
              className={cn(inputClass, "text-center font-semibold text-lg tabular-nums")}
              required
            />
            <StepButton label="1つ増やす" onClick={() => setQty(qty + 1)}>
              <Plus className="size-5" />
            </StepButton>
          </div>
        </div>

        {failed && <Alert>QRを発行できませんでした。入力内容を確認してください。</Alert>}

        <Button type="submit" block disabled={pending || item.trim() === ""}>
          <QrCode className="size-5" aria-hidden />
          QRを発行
        </Button>
      </form>
    </Card>
  );
}

function StepButton({
  label,
  children,
  ...props
}: {
  label: string;
  children: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      className="grid size-12 shrink-0 place-items-center rounded-xl bg-zinc-100 text-zinc-700 hover:bg-zinc-200 disabled:opacity-40 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700"
      {...props}
    >
      {children}
    </button>
  );
}
