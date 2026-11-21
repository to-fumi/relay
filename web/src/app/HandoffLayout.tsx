import { ArrowDownLeft, ArrowUpRight } from "lucide-react";
import { SegmentLayout } from "./SegmentLayout";

/** 受け渡しタブ（⇄）：送る（QR を出す）と受け取る（QR を読む）。受け渡しは 2 人が同時に行うので同じタブに置く */
export function HandoffLayout() {
  return (
    <SegmentLayout
      label="受け渡し"
      segments={[
        { to: "/", label: "送る", icon: ArrowUpRight, end: true },
        { to: "/receive", label: "受け取る", icon: ArrowDownLeft },
      ]}
    />
  );
}
