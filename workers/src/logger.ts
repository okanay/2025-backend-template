import type { Context } from "hono";

export default class DetailLogger {
  private startTime: number;
  private context: Context;
  private requestMeta: any;

  constructor(c: Context) {
    this.startTime = Date.now();
    this.context = c;

    const url = new URL(c.req.url);
    this.requestMeta = {
      path: url.pathname,
      client_ip: c.req.header("cf-connecting-ip") || "unknown",
      user_agent: (c.req.header("user-agent") || "unknown").substring(0, 50),
      cf_country: c.req.header("cf-ipcountry") || "unknown",
    };
  }

  // Sadece özel durumlar için - bot attacks
  logBotAttack(attackType: string) {
    console.log(
      JSON.stringify({
        timestamp: new Date().toISOString(),
        event: "bot_attack_detected",
        execution_time_ms: Date.now() - this.startTime,
        request: this.requestMeta,
        security: {
          attack_type: attackType,
          threat_level: "medium",
          action: "blocked",
        },
      }),
    );
  }

  // Sadece R2 error'ları için
  logR2Error(key: string, error: Error) {
    console.log(
      JSON.stringify({
        timestamp: new Date().toISOString(),
        event: "r2_operation_failed",
        execution_time_ms: Date.now() - this.startTime,
        request: this.requestMeta,
        error: {
          message: error.message,
          type: error.constructor.name,
        },
        asset: { key },
      }),
    );
  }

  // Sadece yavaş request'ler için
  logSlowRequest(duration: number, details: any = {}) {
    if (duration > 100) {
      // >100ms
      console.log(
        JSON.stringify({
          timestamp: new Date().toISOString(),
          event: "slow_request_detected",
          execution_time_ms: duration,
          request: this.requestMeta,
          performance: {
            threshold_exceeded: "100ms",
            actual_duration: `${duration}ms`,
            ...details,
          },
        }),
      );
    }
  }

  // Sadece büyük dosyalar için
  logLargeAsset(key: string, size: number, r2Duration: number) {
    if (size > 5000000) {
      // >5MB
      console.log(
        JSON.stringify({
          timestamp: new Date().toISOString(),
          event: "large_asset_served",
          execution_time_ms: Date.now() - this.startTime,
          request: this.requestMeta,
          asset: {
            key,
            size_bytes: size,
            size_mb: Math.round((size / 1024 / 1024) * 100) / 100,
            r2_fetch_time_ms: r2Duration,
          },
        }),
      );
    }
  }
}
