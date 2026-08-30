import type { NextRequest } from "next/server";

type RouteContext = { params: Promise<{ path: string[] }> };

async function proxy(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  const apiBaseURL = (process.env.API_BASE_URL ?? "http://127.0.0.1:8080").replace(/\/$/, "");
  const target = `${apiBaseURL}/${path.map(encodeURIComponent).join("/")}${request.nextUrl.search}`;
  const headers = new Headers();
  const contentType = request.headers.get("content-type");
  if (contentType) headers.set("content-type", contentType);

  const body = request.method === "GET" || request.method === "HEAD"
    ? undefined
    : new Uint8Array(await request.arrayBuffer());
  const upstream = await fetch(target, {
    method: request.method,
    headers,
    body,
    cache: "no-store",
  });

  return new Response(upstream.body, {
    status: upstream.status,
    headers: { "content-type": upstream.headers.get("content-type") ?? "application/json" },
  });
}

export const dynamic = "force-dynamic";
export const GET = proxy;
export const POST = proxy;
