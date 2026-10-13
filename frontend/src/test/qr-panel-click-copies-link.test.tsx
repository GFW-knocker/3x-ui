import { fireEvent, render } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import QrPanel from '@/pages/inbounds/qr/QrPanel';

const LINK = 'vless://11111111-2222-3333-4444-555555555555@example.com:443?type=tcp#client';

// Clicking the QR used to put a PNG on the clipboard, which pastes as nothing
// in a text field; users read that as "copy is empty". The click must copy the link.
describe('QrPanel click', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('copies the link text, the same as the Copy button', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    const write = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal('isSecureContext', true);
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText, write },
      configurable: true,
    });

    const { container } = render(<QrPanel value={LINK} remark="client" />);
    const canvas = container.querySelector('.qr-panel-canvas');
    expect(canvas).not.toBeNull();

    fireEvent.click(canvas as HTMLElement);
    await vi.waitFor(() => expect(writeText).toHaveBeenCalledWith(LINK));
    expect(write).not.toHaveBeenCalled();
  });
});
