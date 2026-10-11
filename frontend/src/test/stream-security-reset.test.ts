import { describe, expect, it } from 'vitest';

import { dropUnsupportedSecurity } from '@/lib/xray/protocol-capabilities';

const realitySettings = { target: 'www.microsoft.com:443', privateKey: 'k' };
const tlsSettings = { serverName: 'q.example.com' };

// The inbound form kept REALITY after a switch to QUIC/WS/mKCP; the core then refused the
// whole config, so one inbound stopped every inbound on the server.
describe('dropUnsupportedSecurity', () => {
  it('drops REALITY when the transport moves to quic', () => {
    const stream = { network: 'quic', security: 'reality', realitySettings, quicSettings: {} };
    expect(dropUnsupportedSecurity('mvless', stream)).toEqual({
      network: 'quic',
      security: 'none',
      quicSettings: {},
    });
  });

  it.each(['ws', 'kcp', 'httpupgrade'])('drops REALITY over %s', (network) => {
    const out = dropUnsupportedSecurity('vless', { network, security: 'reality', realitySettings });
    expect(out).toEqual({ network, security: 'none' });
  });

  it('drops REALITY when the protocol cannot carry it', () => {
    const stream = { network: 'tcp', security: 'reality', realitySettings };
    expect(dropUnsupportedSecurity('vmess', stream)).toEqual({ network: 'tcp', security: 'none' });
  });

  it.each(['tcp', 'xhttp', 'grpc'])('keeps REALITY over %s', (network) => {
    const stream = { network, security: 'reality', realitySettings };
    expect(dropUnsupportedSecurity('mvless', stream)).toBe(stream);
  });

  it('keeps TLS over quic and drops it over mKCP, as the security tab offers', () => {
    const quic = { network: 'quic', security: 'tls', tlsSettings };
    expect(dropUnsupportedSecurity('vless', quic)).toBe(quic);
    const kcp = { network: 'kcp', security: 'tls', tlsSettings };
    expect(dropUnsupportedSecurity('vless', kcp)).toEqual({ network: 'kcp', security: 'none' });
  });
});
