import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MemoryRouter } from 'react-router';

import type { ClientRecord } from '@/hooks/useClients';
import ClientQrModal from '@/pages/clients/ClientQrModal';
import { HttpUtil, Msg } from '@/utils';

import { renderWithProviders } from './test-utils';

vi.mock('@/pages/inbounds/qr', () => ({
  QrPanel: ({ value }: { value: string }) => <div data-testid="qr-panel-value">{value}</div>,
}));

const CLIENT: ClientRecord = { id: 42, email: 'alice@example.com', subId: 'alpha' };
const CONFIG_LINK =
  'vless://11111111-2222-4333-8444-555555555555@srv.example.com:443?type=tcp&security=tls#alice';

// The QR modal used to lead with the subscription panel; operators want the
// per-inbound config first and open, with the subscription collapsed below it.
describe('ClientQrModal section order', () => {
  beforeEach(() => {
    vi.mocked(HttpUtil.get).mockImplementation(async (url: string) =>
      url.includes('/panel/api/clients/subLinks/') ? new Msg(true, '', [CONFIG_LINK]) : new Msg(true, '', {}),
    );
  });

  it('puts the config first and open, the subscription last and closed', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/clients']}>
        <ClientQrModal
          open
          client={CLIENT}
          inboundsById={{}}
          subSettings={{ enable: true, subURI: 'https://panel.example/sub/', subJsonURI: '', subJsonEnable: false }}
          onOpenChange={() => {}}
        />
      </MemoryRouter>,
    );
    await screen.findByText(CONFIG_LINK);

    await waitFor(() => {
      const headers = Array.from(document.body.querySelectorAll('.ant-collapse-header'));
      expect(headers.length).toBe(2);
      expect(headers[0].textContent).toMatch(/vless/i);
      expect(headers[0].getAttribute('aria-expanded')).toBe('true');
      expect(headers[1].textContent).toMatch(/Subscription/);
      expect(headers[1].getAttribute('aria-expanded')).toBe('false');
    });
  });
});
