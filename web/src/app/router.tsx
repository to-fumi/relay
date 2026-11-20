import { createBrowserRouter } from "react-router";
import { AdminPage } from "../features/admin";
import { AnalyticsPage } from "../features/analytics";
import { IssuePage } from "../features/handoff-issue";
import { ReceivePage } from "../features/handoff-receive";
import { HistoryPage } from "../features/history";
import { Layout } from "./Layout";

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
