import "server-only";

/**
 * Server-only env. Importing this from client code will fail the build thanks to
 * the `server-only` package.
 */
export const serverEnv = {
  apiUrl: process.env.API_URL ?? "http://localhost:8080",
  nodeEnv: process.env.NODE_ENV,
};
