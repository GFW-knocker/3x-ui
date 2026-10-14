import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import UsageRing, { usageRingColor } from '@/pages/index/UsageRing';

// The gauge colour must flip exactly at the 80% and 90% thresholds.
describe('usageRingColor', () => {
  it('is green below 80, orange from 80, red from 90', () => {
    expect(usageRingColor(0)).toBe('#52c41a');
    expect(usageRingColor(79.9)).toBe('#52c41a');
    expect(usageRingColor(80)).toBe('#faad14');
    expect(usageRingColor(89.9)).toBe('#faad14');
    expect(usageRingColor(90)).toBe('#ff4d4f');
    expect(usageRingColor(100)).toBe('#ff4d4f');
  });
});

describe('UsageRing', () => {
  it('draws an arc proportional to the percent and labels it', () => {
    const { container } = render(<UsageRing percent={25} size={28} strokeWidth={3} />);
    const svg = container.querySelector('svg');
    expect(svg?.getAttribute('aria-label')).toBe('25%');
    expect(container.querySelector('text')?.textContent).toBe('25%');
    expect(svg?.getAttribute('data-level')).toBe('ok');
    const arc = container.querySelectorAll('circle')[1];
    const circumference = 2 * Math.PI * ((28 - 3) / 2);
    expect(Number(arc.getAttribute('stroke-dashoffset'))).toBeCloseTo(circumference * 0.75, 5);
  });

  it('clamps out-of-range input', () => {
    const { container } = render(<UsageRing percent={140} />);
    expect(container.querySelector('svg')?.getAttribute('aria-label')).toBe('100%');
    expect(Number(container.querySelectorAll('circle')[1].getAttribute('stroke-dashoffset'))).toBe(0);
  });
});
