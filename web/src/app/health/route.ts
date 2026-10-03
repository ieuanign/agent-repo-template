import { connection } from "next/server";

import { appEnv, version } from "@/core/env";

// Public route: never add the commit or anything beyond these two values.
export async function GET() {
  await connection();
  return Response.json({ version: version(), environment: appEnv() });
}
