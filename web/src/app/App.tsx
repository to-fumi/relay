import { MutationCache, QueryCache, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "react-router";
import { AuthGate, expireSession } from "../shared/auth";
import { problemStatus } from "../shared/lib";
import { router } from "./router";

// API が 401 を返したら、ログインの期限が切れたとみなしてログインし直してもらう
const onError = (error: unknown) => {
  if (problemStatus(error) === 401) expireSession();
};

const queryClient = new QueryClient({
  queryCache: new QueryCache({ onError }),
  mutationCache: new MutationCache({ onError }),
  defaultOptions: {
    queries: { retry: (count, error) => problemStatus(error) !== 401 && count < 3 },
  },
});

export function App() {
  return (
    <AuthGate>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AuthGate>
  );
}
