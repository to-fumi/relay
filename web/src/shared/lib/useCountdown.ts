import { useEffect, useState } from "react";

/** target までの残り秒数（0 以上）を 1 秒ごとに更新して返す。 */
export function useCountdown(target: Date | string | undefined): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    if (!target) return;
    setNow(Date.now());
    const id = window.setInterval(() => setNow(Date.now()), 250);
    return () => window.clearInterval(id);
  }, [target]);
  if (!target) return 0;
  return Math.max(0, Math.ceil((new Date(target).getTime() - now) / 1000));
}
