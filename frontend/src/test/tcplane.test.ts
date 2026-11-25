/// <reference types="vite/client" />
import { describe, expect, it } from 'vitest';

import { genHysteriaLink, genVlessLink, genVmessLink } from '@/lib/xray/inbound-link';
import { parseOutboundLink } from '@/lib/xray/outbound-link-parser';
import { retargetHyTcpTls } from '@/lib/xray/protocol-capabilities';
import { InboundSchema } from '@/schemas/api/inbound';
import { TcpLanePortsSchema } from '@/schemas/protocols/stream/tcplane';

// quictcp/hytcp links carry what the Go generator does: the quic or hysteria fields plus
// lanePassword/lanePorts, the lane options a client must match; the rest stay one-sided.
const clientId = '11111111-2222-4333-8444-555555555555';
const quictcpSettings = {
  security: 'aes-128-gcm',
  key: 'k1',
  header: { type: 'srtp' },
  password: 'lanepw',
  ports: '3000-3002',
};

function quicTcpInbound(protocol: 'vless' | 'vmess') {
  return InboundSchema.parse({
    port: 443,
    protocol,
    settings:
      protocol === 'vless'
        ? { clients: [{ id: clientId, email: 'q' }], decryption: 'none', encryption: 'none' }
        : { clients: [{ id: clientId, email: 'q', security: 'auto' }] },
    streamSettings: {
      network: 'quictcp',
      security: 'none',
      quictcpSettings: { ...quictcpSettings, maxLifeMs: 30000 },
    },
    tag: `in-quictcp-${protocol}`,
  });
}

describe('quictcp transport', () => {
  it('round-trips a vless quictcp link without one-sided lane options', () => {
    const link = genVlessLink({
      inbound: quicTcpInbound('vless'),
      address: 'q.example.test',
      clientId,
    });
    const query = new URL(link).searchParams;
    expect(query.get('type')).toBe('quictcp');
    expect(query.get('lanePassword')).toBe('lanepw');
    expect(query.get('lanePorts')).toBe('3000-3002');
    expect(query.has('maxLifeMs')).toBe(false);
    const parsed = parseOutboundLink(link) as { streamSettings: Record<string, unknown> };
    expect(parsed.streamSettings.network).toBe('quictcp');
    expect(parsed.streamSettings.quictcpSettings).toEqual(quictcpSettings);
  });

  it('round-trips a vmess quictcp link', () => {
    const link = genVmessLink({
      inbound: quicTcpInbound('vmess'),
      address: 'q.example.test',
      clientId,
    });
    const parsed = parseOutboundLink(link) as { streamSettings: Record<string, unknown> };
    expect(parsed.streamSettings.network).toBe('quictcp');
    expect(parsed.streamSettings.quictcpSettings).toEqual(quictcpSettings);
  });
});

describe('hytcp transport', () => {
  it('round-trips a hysteria2 link that names the hytcp lanes', () => {
    const inbound = InboundSchema.parse({
      port: 443,
      protocol: 'hysteria',
      settings: { version: 2, clients: [{ auth: 'secret', email: 'h' }] },
      streamSettings: {
        network: 'hytcp',
        security: 'tls',
        tlsSettings: { serverName: 'h.example.test' },
        hytcpSettings: { version: 2, udpIdleTimeout: 60, password: 'lanepw', ports: 4000 },
      },
      tag: 'in-hytcp',
    });
    const link = genHysteriaLink({
      inbound,
      address: 'h.example.test',
      remark: 'h',
      clientAuth: 'secret',
    });
    const query = new URL(link).searchParams;
    expect(query.get('type')).toBe('hytcp');
    expect(query.get('lanePassword')).toBe('lanepw');
    expect(query.get('lanePorts')).toBe('4000');
    const parsed = parseOutboundLink(link) as { streamSettings: Record<string, unknown> };
    expect(parsed.streamSettings.network).toBe('hytcp');
    expect(parsed.streamSettings.hysteriaSettings).toBeUndefined();
    expect(parsed.streamSettings.hytcpSettings).toEqual({
      version: 2,
      auth: 'secret',
      udpIdleTimeout: 60,
      password: 'lanepw',
      ports: '4000',
    });
  });
});

describe('hytcp under vless and vmess', () => {
  const hytcpSettings = {
    version: 2,
    auth: 'hyauth',
    udpIdleTimeout: 60,
    password: 'lanepw',
    ports: '4000-4001',
  };

  function hyTcpInbound(protocol: 'vless' | 'vmess') {
    return InboundSchema.parse({
      port: 443,
      protocol,
      settings:
        protocol === 'vless'
          ? { clients: [{ id: clientId, email: 'h' }], decryption: 'none', encryption: 'none' }
          : { clients: [{ id: clientId, email: 'h', security: 'auto' }] },
      streamSettings: {
        network: 'hytcp',
        security: 'tls',
        tlsSettings: { serverName: 'h.example.test', alpn: ['h3'] },
        hytcpSettings: { ...hytcpSettings, maxLifeMs: 30000 },
      },
      tag: `in-hytcp-${protocol}`,
    });
  }

  it('round-trips a vless hytcp link with the transport auth', () => {
    const link = genVlessLink({
      inbound: hyTcpInbound('vless'),
      address: 'h.example.test',
      clientId,
    });
    const query = new URL(link).searchParams;
    expect(query.get('type')).toBe('hytcp');
    expect(query.get('auth')).toBe('hyauth');
    expect(query.get('lanePassword')).toBe('lanepw');
    expect(query.has('maxLifeMs')).toBe(false);
    const parsed = parseOutboundLink(link) as { streamSettings: Record<string, unknown> };
    expect(parsed.streamSettings.network).toBe('hytcp');
    expect(parsed.streamSettings.security).toBe('tls');
    expect(parsed.streamSettings.hytcpSettings).toEqual(hytcpSettings);
  });

  it('round-trips a vmess hytcp link with the transport auth', () => {
    const link = genVmessLink({
      inbound: hyTcpInbound('vmess'),
      address: 'h.example.test',
      clientId,
    });
    const parsed = parseOutboundLink(link) as { streamSettings: Record<string, unknown> };
    expect(parsed.streamSettings.network).toBe('hytcp');
    expect(parsed.streamSettings.hytcpSettings).toEqual(hytcpSettings);
  });
});

describe('retargetHyTcpTls', () => {
  const newTls = () => ({ certificates: [{ certificateFile: '/c.crt', keyFile: '/c.key' }] });

  it('puts a plain stream that moves onto hytcp on TLS offering h3', () => {
    const out: Record<string, unknown> = retargetHyTcpTls(
      { network: 'hytcp', security: 'none' },
      'tcp',
      newTls,
    );
    expect(out.security).toBe('tls');
    expect(out.tlsSettings).toEqual({ ...newTls(), alpn: ['h3'] });
  });

  it('keeps the TLS settings it had but swaps an h2 ALPN for h3', () => {
    const tls = { serverName: 'h.example.test', alpn: ['h2', 'http/1.1'] };
    const out: Record<string, unknown> = retargetHyTcpTls(
      { network: 'hytcp', security: 'tls', tlsSettings: tls },
      'tcp',
      newTls,
    );
    expect(out.tlsSettings).toEqual({ serverName: 'h.example.test', alpn: ['h3'] });
  });

  it('drops REALITY, which the hysteria transport cannot carry', () => {
    const out: Record<string, unknown> = retargetHyTcpTls(
      { network: 'hytcp', security: 'reality', realitySettings: { target: 'x:443' } },
      'tcp',
      newTls,
    );
    expect(out.security).toBe('tls');
    expect(out).not.toHaveProperty('realitySettings');
  });

  it('gives back the default ALPN when a stream leaves hytcp with h3 only', () => {
    const tls = { serverName: 'h.example.test', alpn: ['h3'] };
    const out: Record<string, unknown> = retargetHyTcpTls(
      { network: 'ws', security: 'tls', tlsSettings: tls },
      'hytcp',
      newTls,
    );
    expect(out.tlsSettings).toEqual({ serverName: 'h.example.test', alpn: ['h2', 'http/1.1'] });
  });

  it('keeps h3 when a hysteria stream moves from hytcp back to UDP', () => {
    const stream = { network: 'hysteria', security: 'tls', tlsSettings: { alpn: ['h3'] } };
    expect(retargetHyTcpTls(stream, 'hytcp', newTls)).toBe(stream);
  });

  it('leaves an h3 ALPN the user chose on another transport alone', () => {
    const stream = { network: 'xhttp', security: 'tls', tlsSettings: { alpn: ['h3'] } };
    expect(retargetHyTcpTls(stream, 'tcp', newTls)).toBe(stream);
  });
});

// The core parses lane ports as a PortList and opens one listener per port, refusing
// more than 1024; a bad value fails the whole config, not just the one inbound.
describe('TcpLanePortsSchema', () => {
  const issue = (value: unknown) => {
    const result = TcpLanePortsSchema.safeParse(value);
    return result.success ? null : result.error.issues[0]?.message;
  };

  it('accepts ports and ranges up to 1024 listeners', () => {
    for (const value of [
      undefined,
      '',
      4000,
      '3000-3010,4000',
      ' 3000 , 3002-3003 ',
      '20000-21023',
    ])
      expect(issue(value)).toBeNull();
  });

  it('refuses a list the core cannot parse', () => {
    for (const value of ['3000-abc', '70000', 'abc', '3000-3001-3002'])
      expect(issue(value)).toBe('pages.inbounds.form.lanePortsInvalid');
  });

  it('refuses more than 1024 ports, counting the way the core expands them', () => {
    expect(issue('20000-21024')).toBe('pages.inbounds.form.lanePortsTooMany');
    expect(issue('20000-20600,20000-20600')).toBe('pages.inbounds.form.lanePortsTooMany');
  });
});
