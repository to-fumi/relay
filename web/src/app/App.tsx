import { MutationCache, QueryCache, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "react-router";
import { getMeQueryKey } from "../shared/api";
import { AuthGate, isMeQuery } from "../shared/auth";
import { problemStatus } from "../shared/lib";
import { router } from "./router";

// API が 401 を返したら、セッションの期限が切れたとみなして /me を取り直す。/me も 401 になればログインの画面に変わる
const queryClient: QueryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error, query) => {
      if (problemStatus(error) === 401 && !isMeQuery(query)) expireSession();
    },
  }),
  mutationCache: new MutationCache({
    onError: (error) => {
      if (problemStatus(error) === 401) expireSession();
    },
  }),
  defaultOptions: {
    queries: { retry: (count, error) => problemStatus(error) !== 401 && count < 3 },
  },
});

function expireSession() {
  void queryClient.invalidateQueries({ queryKey: getMeQueryKey() });
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthGate>
        <RouterProvider router={router} />
      </AuthGate>
    </QueryClientProvider>
  );
}
