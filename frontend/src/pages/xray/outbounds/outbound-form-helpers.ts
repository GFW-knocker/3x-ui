import { rawOutboundToFormValues } from '@/lib/xray/outbound-form-adapter';
import { canEnableReality, canEnableTls, retargetHyTcpTls } from '@/lib/xray/protocol-capabilities';
import type { OutboundFormValues } from '@/schemas/forms/outbound-form';

import { MUX_PROTOCOLS } from './outbound-form-constants';

export function isMuxAllowed(protocol: string, flow: string, network: string): boolean {
  if (!MUX_PROTOCOLS.has(protocol)) return false;
  if ((protocol === 'vless' || protocol === 'mvless') && flow) return false;
  if (network === 'xhttp') return false;
  return true;
}

// Per-network bootstrap. Mirrors the legacy class constructors so the
// initial state for each transport matches what xray-core expects.
export function newStreamSlice(network: string): Record<string, unknown> {
  switch (network) {
    case 'tcp':
      return { network: 'tcp', tcpSettings: { header: { type: 'none' } } };
    case 'kcp':
      return {
        network: 'kcp',
        kcpSettings: {
          mtu: 1350,
          tti: 20,
          uplinkCapacity: 5,
          downlinkCapacity: 20,
          cwndMultiplier: 1,
          maxSendingWindow: 2097152,
        },
      };
    case 'quic':
      return {
        network: 'quic',
        quicSettings: { security: 'none', key: '', header: { type: 'none' } },
      };
    case 'quictcp':
      return {
        network: 'quictcp',
        quictcpSettings: { security: 'none', key: '', header: { type: 'none' } },
      };
    case 'ws':
      return {
        network: 'ws',
        wsSettings: { path: '/', host: '', headers: {}, heartbeatPeriod: 0 },
      };
    case 'grpc':
      return {
        network: 'grpc',
        grpcSettings: { serviceName: '', authority: '', multiMode: false },
      };
    case 'httpupgrade':
      return {
        network: 'httpupgrade',
        httpupgradeSettings: { path: '/', host: '', headers: {} },
      };
    case 'xhttp':
      return {
        network: 'xhttp',
        xhttpSettings: {
          path: '/',
          host: '',
          mode: '',
          headers: [],
          xPaddingBytes: '100-1000',
        },
      };
    case 'hysteria':
      return {
        network: 'hysteria',
        hysteriaSettings: {
          version: 2,
          auth: '',
          udpIdleTimeout: 60,
        },
      };
    case 'hytcp':
      return {
        network: 'hytcp',
        hytcpSettings: { version: 2, auth: '', udpIdleTimeout: 60 },
      };
    default:
      return { network: 'tcp', tcpSettings: { header: { type: 'none' } } };
  }
}

// Hysteria2 always rides its own QUIC transport with TLS — the panel never
// offers another transport or 'none' security for it.
export function hysteriaStreamSlice(
  network: 'hysteria' | 'hytcp' = 'hysteria',
): Record<string, unknown> {
  return {
    ...newStreamSlice(network),
    security: 'tls',
    tlsSettings: {
      serverName: '',
      alpn: ['h3'],
      fingerprint: '',
      echConfigList: '',
      verifyPeerCertByName: '',
      pinnedPeerCertSha256: '',
    },
  };
}

// Network change cascade: swap the per-network sub-key (tcpSettings,
// wsSettings, etc.) so the DU branch matches. Carry over the security mode
// and its settings (tlsSettings/realitySettings, including SNI serverName)
// when the new network still supports it; otherwise fall back to 'none'.
// Dropping tlsSettings here silently wiped the spoofed SNI on save (#4791).
export function applyNetworkChange(
  protocol: string,
  prevStream: Record<string, unknown> | undefined,
  next: string,
): Record<string, unknown> {
  const stream = prevStream ?? {};
  const prevNetwork = stream.network as string | undefined;
  const hysteriaTls = () => hysteriaStreamSlice('hytcp').tlsSettings as Record<string, unknown>;
  if (next === 'hysteria' || next === 'hytcp') {
    return retargetHyTcpTls(switchHysteriaNetwork(stream, next), prevNetwork, hysteriaTls);
  }
  const currentSecurity = (stream.security as string) ?? 'none';
  const stillTls = canEnableTls({
    protocol,
    streamSettings: { network: next, security: currentSecurity },
  });
  const stillReality = canEnableReality({
    protocol,
    streamSettings: { network: next, security: currentSecurity },
  });
  const newSecurity =
    currentSecurity === 'tls' && !stillTls
      ? 'none'
      : currentSecurity === 'reality' && !stillReality
        ? 'none'
        : currentSecurity;
  const newStream: Record<string, unknown> = { ...newStreamSlice(next), security: newSecurity };
  if (newSecurity === 'tls' && stream.tlsSettings) newStream.tlsSettings = stream.tlsSettings;
  else if (newSecurity === 'reality' && stream.realitySettings)
    newStream.realitySettings = stream.realitySettings;
  return retargetHyTcpTls(newStream, prevNetwork, hysteriaTls);
}

// Auth, idle timeout, masquerade and TLS survive a swap between the two hysteria
// transports; the lane options belong to hytcp alone.
function switchHysteriaNetwork(
  stream: Record<string, unknown>,
  next: 'hysteria' | 'hytcp',
): Record<string, unknown> {
  const slice = hysteriaStreamSlice(next);
  const prev = (stream.hytcpSettings ?? stream.hysteriaSettings) as
    | Record<string, unknown>
    | undefined;
  if (prev) {
    const { auth, udpIdleTimeout, masquerade } = prev;
    const key = `${next}Settings`;
    slice[key] = { ...(slice[key] as Record<string, unknown>), auth, udpIdleTimeout, masquerade };
  }
  if (stream.tlsSettings) slice.tlsSettings = stream.tlsSettings;
  return slice;
}

export function buildAddModeValues(): OutboundFormValues {
  return rawOutboundToFormValues({});
}
