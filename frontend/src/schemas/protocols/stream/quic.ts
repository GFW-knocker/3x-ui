import { z } from 'zod';

// QUIC transport, removed from xray-core but kept by Mahsa-Core. The core reads an
// unknown cipher as none and accepts the mKCP header types except dns, which needs a domain.
export const QUIC_SECURITIES = ['none', 'aes-128-gcm', 'chacha20-poly1305'] as const;
export const QUIC_HEADER_TYPES = [
  'none',
  'srtp',
  'utp',
  'wechat-video',
  'dtls',
  'wireguard',
] as const;

export const QuicStreamSettingsSchema = z.object({
  security: z.enum(QUIC_SECURITIES).default('none'),
  key: z.string().default(''),
  header: z
    .object({
      type: z.enum(QUIC_HEADER_TYPES).default('none'),
    })
    .default({ type: 'none' }),
});
export type QuicStreamSettings = z.infer<typeof QuicStreamSettingsSchema>;
