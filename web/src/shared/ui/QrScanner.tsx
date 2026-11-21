import { type IScannerError, prepareZXingModule, Scanner } from "@yudiel/react-qr-scanner";
import { useState } from "react";
// QR を解読する wasm を、外部の CDN からではなくアプリと一緒に配信する（社内のプロキシで止まらないように）
import zxingReaderWasmUrl from "zxing-wasm/reader/zxing_reader.wasm?url";
import { Alert } from "./Alert";

// ブラウザに QR を読む機能がないとき（iOS Safari など）に使う wasm の場所を、読み取りの前に差し替える。
// バージョンは pnpm-workspace.yaml の overrides で、ライブラリが使う zxing-wasm とそろえている
prepareZXingModule({
  overrides: {
    locateFile: (path: string, prefix: string) => (path.endsWith(".wasm") ? zxingReaderWasmUrl : prefix + path),
  },
});

// ライブラリの既定値は高さの最小が 640px で、640x480 のカメラが起動しないため、希望値だけを指定する
const cameraConstraints: MediaTrackConstraints = {
  facingMode: "environment",
  width: { ideal: 1280 },
  height: { ideal: 720 },
};

type Props = {
  onScan: (value: string) => void;
  paused?: boolean;
};

/** 背面カメラで QR を読み取る。合わせる枠の中だけ明るく、周りは半透明の黒で暗くする。 */
export function QrScanner({ onScan, paused }: Props) {
  const [cameraError, setCameraError] = useState<IScannerError | null>(null);
  return (
    <>
      <div className="relative overflow-hidden rounded-3xl bg-black">
        <Scanner
          constraints={cameraConstraints}
          onScan={(codes) => codes[0] && onScan(codes[0].rawValue)}
          onError={(e) => {
            console.error("scanner error", e.kind, e.message, e.cause);
            setCameraError(e);
          }}
          paused={paused}
          components={{ finder: false }}
          styles={{ container: { aspectRatio: "1 / 1" }, video: { objectFit: "cover" } }}
        />
        <ScanFrame />
      </div>
      {cameraError && (
        <div className="mt-4">
          <Alert>カメラを起動できませんでした（{cameraError.kind}）</Alert>
        </div>
      )}
    </>
  );
}

/** 中央の四角だけを透かし、周りを暗くする。四隅に白いカギ形を置く */
function ScanFrame() {
  const corner = "absolute size-8 border-white";
  return (
    <div className="pointer-events-none absolute inset-0 grid place-items-center" aria-hidden>
      {/* 大きな影で、四角の外側だけを暗くする */}
      <div className="relative size-[68%] rounded-2xl shadow-[0_0_0_100vmax_rgba(0,0,0,0.55)]">
        <span className={`${corner} -top-px -left-px rounded-tl-2xl border-t-[3px] border-l-[3px]`} />
        <span className={`${corner} -top-px -right-px rounded-tr-2xl border-t-[3px] border-r-[3px]`} />
        <span className={`${corner} -bottom-px -left-px rounded-bl-2xl border-b-[3px] border-l-[3px]`} />
        <span className={`${corner} -right-px -bottom-px rounded-br-2xl border-r-[3px] border-b-[3px]`} />
      </div>
    </div>
  );
}
