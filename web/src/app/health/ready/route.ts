import { connection } from "next/server";

// Never check backend here: its outage would pull web out of Traefik and users would see a bare 503.
export async function GET() {
  await connection();
  return Response.json({});
}
