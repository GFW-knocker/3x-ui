import { describe, expect, it } from 'vitest';

import { truncateMiddle } from '@/utils';

// The clients table shows a UUID as its first and last four characters; a
// credential short enough to show whole must not gain an ellipsis.
describe('truncateMiddle', () => {
  it('keeps four characters at each end of a UUID', () => {
    expect(truncateMiddle('11111111-2222-4333-8444-555555555555')).toBe('1111…5555');
  });

  it('returns a short value untouched', () => {
    expect(truncateMiddle('abcdefgh')).toBe('abcdefgh');
    expect(truncateMiddle('abcdefghi')).toBe('abcdefghi');
    expect(truncateMiddle('abcdefghij')).toBe('abcd…ghij');
  });
});
