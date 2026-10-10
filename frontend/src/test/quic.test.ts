/// <reference types="vite/client" />
import { describe, expect, it } from 'vitest';

import { genVlessLink, genVmessLink } from '@/lib/xray/inbound-link';
import { parseOutboundLink } from '@/lib/xray/outbound-link-parser';
import { InboundSchema } from '@/schemas/api/inbound';
import { NetworkSettingsSchema } from '@/schemas/protocols';

// Mahsa-Core keeps xray's QUIC transport; links carry it as the Go generator does:
// quicSecurity/key/headerType params, and type/host/path in the vmess JSON object.
const quicSettings = {
  security: 'chacha20-poly1305',
  key: 's3cret',
  header: { type: 'wechat-video' },
};
const streamSettings = { network: 'quic', security: 'none', quicSettings };
const clientId = '11111111-2222-4333-8444-555555555555';

function quicInbound(protocol: 'vless' | 'vmess') {
  return InboundSchema.parse({
    port: 443,
    protocol,
    settings:
      protocol === 'vless'
        ? { clients: [{ id: clientId, email: 'q' }], decryption: 'none', encryption: 'none' }
        : { clients: [{ id: clientId, email: 'q', security: 'auto' }] },
    streamSettings,
    tag: `in-quic-${protocol}`,
  });
}

describe('quic transport', () => {
  it('fills the core defaults for a bare quic stream', () => {
    const parsed = NetworkSettingsSchema.parse({ network: 'quic', quicSettings: {} });
    expect(parsed).toEqual({
      network: 'quic',
      quicSettings: { security: 'none', key: '', header: { type: 'none' } },
    });
  });

  it('round-trips a vless quic link', () => {
    const inbound = quicInbound('vless');
    const link = genVlessLink({ inbound, address: 'q.example.test', clientId });
    const query = new URL(link).searchParams;
    expect(query.get('type')).toBe('quic');
    expect(query.get('quicSecurity')).toBe('chacha20-poly1305');
    expect(query.get('key')).toBe('s3cret');
    expect(query.get('headerType')).toBe('wechat-video');
    const parsed = parseOutboundLink(link) as { streamSettings: Record<string, unknown> };
    expect(parsed.streamSettings.network).toBe('quic');
    expect(parsed.streamSettings.quicSettings).toEqual(quicSettings);
  });

  it('round-trips a vmess quic link', () => {
    const inbound = quicInbound('vmess');
    const link = genVmessLink({ inbound, address: 'q.example.test', clientId });
    const parsed = parseOutboundLink(link) as { streamSettings: Record<string, unknown> };
    expect(parsed.streamSettings.network).toBe('quic');
    expect(parsed.streamSettings.quicSettings).toEqual(quicSettings);
  });
});
