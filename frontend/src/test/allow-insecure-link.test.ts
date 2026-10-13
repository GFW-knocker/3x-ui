/// <reference types="vite/client" />
import { describe, expect, it } from 'vitest';

import {
  genHysteriaLink,
  genTrojanLink,
  genVlessLink,
  genVmessLink,
} from '@/lib/xray/inbound-link';
import { InboundSchema } from '@/schemas/api/inbound';

import hysteriaRaw from './golden/fixtures/inbound-full/hysteria-tls.json';
import trojanRaw from './golden/fixtures/inbound-full/trojan-ws-tls.json';
import vlessRaw from './golden/fixtures/inbound-full/vless-ws-tls.json';
import vmessRaw from './golden/fixtures/inbound-full/vmess-tcp-tls.json';

// Mahsa-Core kept allowInsecure after upstream Xray dropped it; the inbound's
// TLS switch has to reach every share link or self-signed setups cannot connect.
function withAllowInsecure(raw: Record<string, unknown>, allow: boolean) {
  const copy = structuredClone(raw) as {
    streamSettings: { tlsSettings: { settings: Record<string, unknown> } };
  };
  copy.streamSettings.tlsSettings.settings.allowInsecure = allow;
  return InboundSchema.parse(copy);
}

describe('allowInsecure in share links', () => {
  it('vless and trojan carry allowInsecure=1 only when switched on', () => {
    for (const allow of [true, false]) {
      const vless = genVlessLink({
        inbound: withAllowInsecure(vlessRaw as Record<string, unknown>, allow),
        address: 'srv.example.com',
        clientId: '11111111-2222-4333-8444-555555555555',
        remark: 'r',
      });
      expect(new URL(vless).searchParams.get('allowInsecure'), `vless allow=${allow}`).toBe(
        allow ? '1' : null,
      );
      const trojan = genTrojanLink({
        inbound: withAllowInsecure(trojanRaw as Record<string, unknown>, allow),
        address: 'srv.example.com',
        clientPassword: 'pw',
        remark: 'r',
      });
      expect(new URL(trojan).searchParams.get('allowInsecure'), `trojan allow=${allow}`).toBe(
        allow ? '1' : null,
      );
    }
  });

  it('vmess object carries allowInsecure: true', () => {
    const link = genVmessLink({
      inbound: withAllowInsecure(vmessRaw as Record<string, unknown>, true),
      address: 'srv.example.com',
      clientId: '11111111-2222-4333-8444-555555555555',
      remark: 'r',
    });
    const obj = JSON.parse(atob(link.replace('vmess://', ''))) as Record<string, unknown>;
    expect(obj.allowInsecure).toBe(true);
  });

  it('hysteria uses the insecure=1 spelling', () => {
    const link = genHysteriaLink({
      inbound: withAllowInsecure(hysteriaRaw as Record<string, unknown>, true),
      address: 'srv.example.com',
      clientAuth: 'secret',
      remark: 'r',
    });
    expect(new URL(link).searchParams.get('insecure')).toBe('1');
  });
});
