import { z } from 'zod';

import { HysteriaStreamSettingsSchema } from './hysteria';
import { QuicStreamSettingsSchema } from './quic';

// Mahsa-Core's tcplane options: the transport's datagrams ride many short TCP "lanes".
// Ranges take a number or "min-max"; an absent key keeps the core's measured default.
const LaneRangeSchema = z.union([z.number().int(), z.string()]).optional();
const LaneCountSchema = z.number().int().optional();

export const TcpLaneSettingsShape = {
  password: z.string().optional(),
  ports: z.union([z.string(), z.number().int()]).optional(),
  packets: LaneRangeSchema,
  lifeMs: LaneRangeSchema,
  bytes: LaneRangeSchema,
  lanes: z.number().int().min(1).max(64).optional(),
  maxRate: LaneCountSchema,
  graceMs: LaneCountSchema,
  idleMs: LaneCountSchema,
  downLifeMs: LaneCountSchema,
  flushMs: LaneCountSchema,
  maxLifeMs: LaneCountSchema,
  sndbuf: LaneCountSchema,
  maxLanesPerIP: LaneCountSchema,
};

// "quictcpSettings": the quic transport's settings and the tcplane options in one object.
export const QuicTcpStreamSettingsSchema = QuicStreamSettingsSchema.extend(TcpLaneSettingsShape);
export type QuicTcpStreamSettings = z.infer<typeof QuicTcpStreamSettingsSchema>;

// "hytcpSettings": the hysteria transport's settings and the tcplane options in one object.
export const HyTcpStreamSettingsSchema = HysteriaStreamSettingsSchema.extend(TcpLaneSettingsShape);
export type HyTcpStreamSettings = z.infer<typeof HyTcpStreamSettingsSchema>;

// vless/vmess/trojan over hytcp: the hysteria listener's own auth, without which the core cannot start.
export const HyTcpTransportAuthSchema = z
  .string()
  .trim()
  .min(1, { error: 'pages.inbounds.form.hyTransportAuthRequired' });

// The core opens one listener per lane port and refuses more (transport_finalmask_tcplane.go).
export const TCP_LANE_MAX_PORTS = 1024;

function parsePortText(text: string): number | null {
  if (!/^\d+$/.test(text)) return null;
  const port = Number(text);
  return port <= 65535 ? port : null;
}

// lanePortCount expands lane ports the way the core's PortList does ("a-b" ranges and
// single ports, comma-separated, duplicates counted); null when the core would refuse it.
export function lanePortCount(value: string | number | undefined): number | null {
  if (value === undefined) return 0;
  if (typeof value === 'number') return value >= 0 && value <= 65535 ? Math.min(value, 1) : null;
  let count = 0;
  for (const raw of value.split(',')) {
    const item = raw.trim();
    if (!item) continue;
    const dash = item.indexOf('-');
    if (dash < 0) {
      if (parsePortText(item) === null) return null;
      count += 1;
      continue;
    }
    const from = parsePortText(item.slice(0, dash));
    const to = parsePortText(item.slice(dash + 1));
    if (from === null || to === null) return null;
    count += Math.max(0, to - from + 1);
  }
  return count;
}

export const TcpLanePortsSchema = z
  .union([z.string(), z.number().int()])
  .optional()
  .superRefine((value, ctx) => {
    const count = lanePortCount(value);
    if (count === null) {
      ctx.addIssue({ code: 'custom', message: 'pages.inbounds.form.lanePortsInvalid' });
    } else if (count > TCP_LANE_MAX_PORTS) {
      ctx.addIssue({ code: 'custom', message: 'pages.inbounds.form.lanePortsTooMany' });
    }
  });
