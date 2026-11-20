import { lazy } from "react";
import { createBrowserRouter } from "react-router";
import { IssuePage } from "../features/handoff-issue";
import { Layout } from "./Layout";

// カメラや可視化など重い feature は、開いたときに読み込む
const ReceivePage = lazy(() => import("../features/handoff-receive").then((m) => ({ default: m.ReceivePage })));
const HistoryPage = lazy(() => import("../features/history").then((m) => ({ default: m.HistoryPage })));
const AnalyticsPage = lazy(() => import("../features/analytics").then((m) => ({ default: m.AnalyticsPage })));
const AdminPage = lazy(() => import("../features/admin").then((m) => ({ default: m.AdminPage })));

export const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { index: true, element: <IssuePage /> },
      { path: "receive", element: <ReceivePage /> },
      { path: "history", element: <HistoryPage /> },
      { path: "analytics", element: <AnalyticsPage /> },
      { path: "admin", element: <AdminPage /> },
    ],
  },
]);
