import { lazy } from "react";
import { createBrowserRouter } from "react-router";
import { IssuePage } from "../features/handoff-issue";
import { HandoffLayout } from "./HandoffLayout";
import { Layout } from "./Layout";
import { ReturnLayout } from "./ReturnLayout";

// カメラなど重い feature は、開いたときに読み込む
const ReceivePage = lazy(() => import("../features/handoff-receive").then((m) => ({ default: m.ReceivePage })));
const ReturnPage = lazy(() => import("../features/returns").then((m) => ({ default: m.ReturnPage })));
const WaitingPage = lazy(() => import("../features/returns").then((m) => ({ default: m.WaitingPage })));
const HistoryPage = lazy(() => import("../features/history").then((m) => ({ default: m.HistoryPage })));

export const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      {
        element: <HandoffLayout />,
        children: [
          { index: true, element: <IssuePage /> },
          { path: "receive", element: <ReceivePage /> },
        ],
      },
      {
        element: <ReturnLayout />,
        children: [
          { path: "return", element: <ReturnPage /> },
          { path: "waiting", element: <WaitingPage /> },
        ],
      },
      { path: "history", element: <HistoryPage /> },
    ],
  },
]);
