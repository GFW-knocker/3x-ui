import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { Status } from '@/models/status';

// jsdom has no canvas; the sparkline under test is not what these cases check.
vi.mock('uplot', () => ({
  default: class {
    static paths = { spline: () => undefined };
    static pxRatio = 1;
    setData() {}
    setSize() {}
    redraw() {}
    destroy() {}
  },
}));
import SystemStrip from '@/pages/index/SystemStrip';
import ThroughputCard from '@/pages/index/ThroughputCard';

// The 1/5/15-minute load averages were sampled and pushed for months but never
// rendered; the strip must show all three in order after uptime.
describe('SystemStrip', () => {
  it('shows the three load averages', () => {
    const status = new Status({ loads: [1.5, 0.75, 0.25], uptime: 60 });
    const { container } = render(<SystemStrip status={status} showIp={false} onToggleIp={() => {}} />);
    const text = container.textContent ?? '';
    const load = text.indexOf('System Load');
    expect(load).toBeGreaterThan(text.indexOf('Uptime'));
    for (const value of ['1.5', '0.75', '0.25']) {
      expect(text.indexOf(value), value).toBeGreaterThan(load);
    }
    expect(text.indexOf('Panel')).toBeGreaterThan(text.indexOf('0.25'));
  });
});

// Averages first, totals pushed to the right edge.
describe('ThroughputCard foot', () => {
  it('orders averages before totals and drops the window label', () => {
    const status = new Status({ netTraffic: { sent: 1024, recv: 2048 }, netIO: { up: 10, down: 20 } });
    const { container } = render(
      <ThroughputCard status={status} up={[10, 10]} down={[20, 20]} labels={['a', 'b']} isMobile={false} />,
    );
    const text = container.textContent ?? '';
    const order = ['Avg Upload', 'Avg Download', 'Total Sent', 'Total Received'].map((label) =>
      text.indexOf(label),
    );
    expect(order.every((i) => i >= 0), text).toBe(true);
    expect([...order].sort((a, b) => a - b)).toEqual(order);
    expect(text).not.toContain('Avg over window');
    expect(container.querySelector('.ov-foot-spacer')).not.toBeNull();
  });
});
