import { z } from 'zod';

export const ProtocolSchema = z.enum([
  'vmess',
  'vless',
  'mvless',
  'trojan',
  'shadowsocks',
  'wireguard',
  'hysteria',
  'http',
  'mixed',
  'tunnel',
  'tun',
  'mtproto',
  'amneziawg',
  'tuic',
]);
export type Protocol = z.infer<typeof ProtocolSchema>;

// Const map matching the legacy models/inbound.ts `Protocols` export so
// call sites can swap the import without touching `Protocols.VLESS`-style
// references throughout the codebase. Frozen so downstream code can't
// mutate the dispatch table. TUN is kept here for parity even though the
// Go backend's validator no longer accepts it — existing panel deployments
// may still have TUN inbounds saved that we want to render.
export const Protocols = Object.freeze({
  VMESS: 'vmess',
  VLESS: 'vless',
  MVLESS: 'mvless',
  TROJAN: 'trojan',
  SHADOWSOCKS: 'shadowsocks',
  WIREGUARD: 'wireguard',
  HYSTERIA: 'hysteria',
  HTTP: 'http',
  MIXED: 'mixed',
  TUNNEL: 'tunnel',
  TUN: 'tun',
  MTPROTO: 'mtproto',
  AMNEZIAWG: 'amneziawg',
  TUIC: 'tuic',
});

// Mahsa-Core's MVLESS takes VLESS's settings and differs only on the wire, so every
// VLESS-specific rule (flow, encryption, fallbacks, reverse) applies to it too.
export function isVlessFamily(protocol: string | undefined | null): boolean {
  return protocol === Protocols.VLESS || protocol === Protocols.MVLESS;
}
