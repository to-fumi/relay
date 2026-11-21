import { useQuery } from "@tanstack/react-query";
import { Hourglass, Undo2 } from "lucide-react";
import { getMeOptions, listMyOutstandingHandoffsOptions } from "../shared/api";
import { daysUntil } from "../shared/lib";
import { SegmentLayout } from "./SegmentLayout";

/** 返却タブ：返す（自分が受け取ったもの）と待ち（自分が送ったもの） */
export function ReturnLayout() {
  const me = useQuery(getMeOptions());
  const outstanding = useQuery(listMyOutstandingHandoffsOptions());
  const all = outstanding.data ?? [];
  const toReturn = all.filter((h) => h.to.id === me.data?.id);
  const waiting = all.filter((h) => h.from.id === me.data?.id);
  const overdue = (hs: typeof all) => hs.some((h) => h.dueDate !== undefined && daysUntil(h.dueDate) < 0);

  return (
    <SegmentLayout
      label="返却"
      segments={[
        { to: "/return", label: "返す", icon: Undo2, count: toReturn.length, alert: overdue(toReturn) },
        { to: "/waiting", label: "待ち", icon: Hourglass, count: waiting.length, alert: overdue(waiting) },
      ]}
    />
  );
}
