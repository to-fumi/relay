import { getAccessToken } from "../auth";
import { client } from "./generated/client.gen";

client.setConfig({ auth: () => getAccessToken() });

export * from "./generated/@tanstack/react-query.gen";
export type * from "./generated/types.gen";
