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
