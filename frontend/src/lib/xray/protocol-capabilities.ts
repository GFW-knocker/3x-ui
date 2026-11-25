// Pure-function ports of the legacy Inbound class capability predicates
// (canEnableTls, canEnableReality, canEnableTlsFlow, canEnableStream,
// canEnableVisionSeed, isSS2022, isSSMultiUser). Each accepts the minimal
// slice of an InboundFormValues it needs, so the same predicate can be
// called against a partial-row, a full form value, or a hand-built test
// fixture without the caller projecting a whole object.

const TLS_ELIGIBLE_PROTOCOLS = ['vmess', 'vless', 'mvless', 'trojan', 'shadowsocks'];
const TLS_NETWORKS = [
  'tcp',
  'ws',
  'http',
  'grpc',
  'httpupgrade',
  'xhttp',
  'quic',
  'quictcp',
  'hytcp',
];
// Outside the hysteria protocol, hytcp authenticates with its own transport password.
const HYTCP_PROTOCOLS = ['vless', 'mvless', 'vmess', 'trojan'];
// The core's ALPN when none is set (tls/config.go), offered again when h3 stops fitting.
const CORE_DEFAULT_ALPN = ['h2', 'http/1.1'];
const REALITY_ELIGIBLE_PROTOCOLS = ['vless', 'mvless', 'trojan'];
const REALITY_NETWORKS = ['tcp', 'http', 'grpc', 'xhttp'];
const STREAM_PROTOCOLS = [
  'vmess',
  'vless',
  'mvless',
  'trojan',
  'shadowsocks',
  'hysteria',
  'wireguard',
  'tunnel',
];
const VISION_FLOW = 'xtls-rprx-vision';
const SS_2022_PREFIX = '2022';
const SS_BLAKE3_CHACHA20 = '2022-blake3-chacha20-poly1305';

export interface CapabilityProtocolSlice {
  protocol: string;
  settings?: { encryption?: string; decryption?: string };
  streamSettings?: { network?: string; security?: string };
}

export interface CapabilityVlessSlice extends CapabilityProtocolSlice {
  settings?: { encryption?: string; decryption?: string; clients?: { flow?: string }[] };
}

export interface CapabilityShadowsocksSlice extends CapabilityProtocolSlice {
  settings?: { encryption?: string; method?: string };
}

export function canEnableTls(values: CapabilityProtocolSlice): boolean {
  if (values.protocol === 'hysteria') return true;
  if (!TLS_ELIGIBLE_PROTOCOLS.includes(values.protocol)) return false;
  return TLS_NETWORKS.includes(values.streamSettings?.network ?? '');
}

export function canUseHyTcp(protocol: string): boolean {
  return HYTCP_PROTOCOLS.includes(protocol);
}

// hytcp's hysteria listener serves HTTP/3, so the stream moves onto TLS offering h3; leaving
// it for a non-hysteria transport, an h3-only ALPN would turn away h2/http/1.1 clients.
export function retargetHyTcpTls<T extends Record<string, unknown>>(
  stream: T,
  prevNetwork: string | undefined,
  newTls: () => Record<string, unknown>,
): T {
  const tls = stream.tlsSettings as Record<string, unknown> | undefined;
  if (stream.network === 'hytcp') {
    const base = stream.security === 'tls' && tls ? tls : newTls();
    const moved: Record<string, unknown> = {
      ...stream,
      security: 'tls',
      tlsSettings: { ...base, alpn: ['h3'] },
    };
    delete moved.realitySettings;
    return moved as T;
  }
  const alpn = tls?.alpn;
  const h3Only = Array.isArray(alpn) && alpn.length === 1 && alpn[0] === 'h3';
  const leavesForNonHysteria = prevNetwork === 'hytcp' && stream.network !== 'hysteria';
  if (!leavesForNonHysteria || !tls || !h3Only) return stream;
  return { ...stream, tlsSettings: { ...tls, alpn: [...CORE_DEFAULT_ALPN] } };
}

export function canEnableReality(values: CapabilityProtocolSlice): boolean {
  if (!REALITY_ELIGIBLE_PROTOCOLS.includes(values.protocol)) return false;
  return REALITY_NETWORKS.includes(values.streamSettings?.network ?? '');
}

// A protocol or transport switch can strand a security the new pair cannot carry; the
// core refuses the whole config over e.g. REALITY on QUIC, so it falls back to none.
export function dropUnsupportedSecurity<T extends Record<string, unknown>>(
  protocol: string,
  stream: T,
): T {
  const network = typeof stream.network === 'string' ? stream.network : undefined;
  const security = typeof stream.security === 'string' ? stream.security : undefined;
  const slice = { protocol, streamSettings: { network, security } };
  const stranded =
    (security === 'tls' && !canEnableTls(slice)) ||
    (security === 'reality' && !canEnableReality(slice));
  if (!stranded) return stream;
  const cleaned: Record<string, unknown> = { ...stream, security: 'none' };
  delete cleaned.tlsSettings;
  delete cleaned.realitySettings;
  return cleaned as T;
}

// VLESS encryption (vlessenc / ML-KEM) is on when encryption or decryption holds
// a generated value (e.g. "mlkem768x25519plus.native.0rtt.<key>") rather than
// the "none"/"" sentinel. The value is never the literal "vlessenc" (that is the
// `xray vlessenc` subcommand). decryption is the server-side value; encryption is
// stored for link generation — either being set means it is on.
function hasVlessEncryption(settings: CapabilityProtocolSlice['settings']): boolean {
  const isSet = (v?: string) => v != null && v !== '' && v !== 'none';
  return isSet(settings?.encryption) || isSet(settings?.decryption);
}

export function canEnableTlsFlow(values: CapabilityProtocolSlice): boolean {
  if (values.protocol !== 'vless' && values.protocol !== 'mvless') return false;
  const network = values.streamSettings?.network;
  const security = values.streamSettings?.security;

  // Classic XTLS Vision: raw TCP carried over TLS or REALITY.
  if (network === 'tcp' && (security === 'tls' || security === 'reality')) return true;

  // vlessenc carries Vision over XHTTP without transport TLS.
  if (network === 'xhttp' && hasVlessEncryption(values.settings)) return true;

  return false;
}

export function canEnableStream(values: { protocol: string }): boolean {
  return STREAM_PROTOCOLS.includes(values.protocol);
}

// mtproto and amneziawg are served by an external process/interface, not
// Xray, so the Xray sniffing block does not apply to either. Every other
// inbound supports sniffing.
export function canEnableSniffing(values: { protocol: string }): boolean {
  return (
    values.protocol !== 'mtproto' && values.protocol !== 'amneziawg' && values.protocol !== 'tuic'
  );
}

// Vision seed applies only when XTLS Vision (TCP/TLS) flow is selected
// AND at least one VLESS client uses the vision flow. Excludes UDP variant.
export function canEnableVisionSeed(values: CapabilityVlessSlice): boolean {
  if (!canEnableTlsFlow(values)) return false;
  const clients = values.settings?.clients;
  if (!Array.isArray(clients)) return false;
  return clients.some((c) => c?.flow === VISION_FLOW);
}

// Why: legacy returns true on non-SS protocols too (the method getter
// resolves to "" and "" !== blake3-chacha20-poly1305). Preserved for
// parity with the legacy class; in practice the callers all narrow on
// protocol === shadowsocks before checking.
export function isSSMultiUser(values: CapabilityShadowsocksSlice): boolean {
  const method = values.protocol === 'shadowsocks' ? (values.settings?.method ?? '') : '';
  return method !== SS_BLAKE3_CHACHA20;
}

export function isSS2022(values: CapabilityShadowsocksSlice): boolean {
  const method = values.protocol === 'shadowsocks' ? (values.settings?.method ?? '') : '';
  return method.substring(0, 4) === SS_2022_PREFIX;
}
