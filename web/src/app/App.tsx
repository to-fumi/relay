import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "react-router";
import { AuthGate } from "../shared/auth";
import { router } from "./router";

const queryClient = new QueryClient();

export function App() {
  return (
    <AuthGate>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AuthGate>
  );
}
