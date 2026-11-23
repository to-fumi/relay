import { useState } from "react";

/** 一覧で選んだ id。今の一覧にないもの（返却済みなど）は数えない */
export function useSelection(available: string[]) {
  const [picked, setPicked] = useState<ReadonlySet<string>>(new Set());
  const selected = available.filter((id) => picked.has(id));
  return {
    selected,
    isSelected: (id: string) => picked.has(id),
    toggle: (id: string, on: boolean) =>
      setPicked((prev) => {
        const next = new Set(prev);
        if (on) next.add(id);
        else next.delete(id);
        return next;
      }),
    clear: () => setPicked(new Set()),
  };
}
