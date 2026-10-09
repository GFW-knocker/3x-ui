/// <reference types="vite/client" />
import { describe, expect, it } from 'vitest';

import { genVlessLink } from '@/lib/xray/inbound-link';
import { parseOutboundLink } from '@/lib/xray/outbound-link-parser';
import { InboundSchema } from '@/schemas/api/inbound';

// Mahsa-Core's MVLESS shares VLESS's settings and differs only on the wire, so its share
// link is the VLESS link under the mvless:// scheme, and parsing it keeps the protocol.
const vlessFixtures = import.meta.glob<unknown>('./golden/fixtures/inbound-full/vless-*.json', {
  eager: true,
  import: 'default',
});

describe('mvless share links', () => {
  const entries = Object.entries(vlessFixtures);

  it('has vless fixtures to compare against', () => {
    expect(entries.length).toBeGreaterThan(0);
  });

  for (const [path, raw] of entries) {
    const name = path.split('/').pop() ?? path;
    const row = raw as { settings: { clients: Array<{ id: string }> } };
    const clientId = row.settings.clients[0].id;
    const vless = InboundSchema.parse(raw);
    const mvless = InboundSchema.parse({ ...(raw as Record<string, unknown>), protocol: 'mvless' });
    const vlessLink = genVlessLink({ inbound: vless, address: 'example.test', clientId });
    const mvlessLink = genVlessLink({ inbound: mvless, address: 'example.test', clientId });

    it(`${name}: is the vless link under the mvless scheme`, () => {
      expect(vlessLink.startsWith('vless://')).toBe(true);
      expect(mvlessLink).toBe(vlessLink.replace(/^vless:\/\//, 'mvless://'));
    });

    it(`${name}: parses back into an mvless outbound`, () => {
      const fromVless = parseOutboundLink(vlessLink);
      const fromMvless = parseOutboundLink(mvlessLink);
      expect(fromVless?.protocol).toBe('vless');
      expect(fromMvless).toEqual({ ...fromVless, protocol: 'mvless' });
    });
  }
});
