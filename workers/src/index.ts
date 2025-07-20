import { Hono } from "hono";
import { logger } from "hono/logger";
import DetailLogger from "./logger";
import { ASSET_EXTENSION_REGEX, BOT_PATH_REGEX } from "./configs";

interface Env {
  R2_ASSETS: R2Bucket;
  BACKEND_URL: string;
}

const app = new Hono<{ Bindings: Env }>();
app.use("*", logger());

app.get("/*", async (c) => {
  const detailLogger = new DetailLogger(c);
  const pathname = new URL(c.req.url).pathname;

  // Bot attack - ekstra detay log
  if (BOT_PATH_REGEX.test(pathname)) {
    detailLogger.logBotAttack("malicious_path");

    return new Response("Forbidden", {
      status: 403,
      headers: {
        "cache-control": "public, max-age=604800",
        "content-type": "text/plain",
      },
    });
  }

  // Invalid extension - sadece Hono logger yeterli (detay log yok)
  if (!ASSET_EXTENSION_REGEX.test(pathname)) {
    return new Response("Not Found", {
      status: 404,
      headers: {
        "cache-control": "public, max-age=1800",
        "content-type": "text/plain",
      },
    });
  }

  const key = pathname.slice(1);
  const requestStart = performance.now();

  try {
    const r2StartTime = performance.now();
    const object = await c.env.R2_ASSETS.get(key);
    const r2Duration = performance.now() - r2StartTime;
    const totalDuration = performance.now() - requestStart;

    if (!object) {
      // Normal 404 - sadece Hono logger yeterli
      return new Response("Not Found", {
        status: 404,
        headers: { "cache-control": "public, max-age=300" },
      });
    }

    const headers = new Headers();
    object.writeHttpMetadata(headers);
    headers.set("etag", object.httpEtag);
    headers.set("cache-control", "public, max-age=31536000, immutable");

    detailLogger.logSlowRequest(totalDuration, {
      r2_fetch_time_ms: r2Duration,
    });

    detailLogger.logLargeAsset(key, object.size, r2Duration);

    return new Response(object.body, { headers });
  } catch (error) {
    detailLogger.logR2Error(key, error as Error);

    return new Response("Error", { status: 500 });
  }
});

export default app;
