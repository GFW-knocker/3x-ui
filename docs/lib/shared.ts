export const appName = '3x-ui';
export const appTagline = 'Advanced web panel for managing Xray-core servers';

export const docsRoute = '/docs';
export const docsImageRoute = '/og/docs';
export const docsContentRoute = '/llms.mdx/docs';

// The 3x-ui product repository — used for the navbar GitHub link,
// build-time star/release stats, and install commands.
export const productRepo = {
  user: 'GFW-knocker',
  repo: '3x-ui',
  branch: 'main',
};

// Where these docs live in the 3x-ui monorepo — used for "Edit on GitHub" links.
export const gitConfig = {
  user: 'GFW-knocker',
  repo: '3x-ui',
  branch: 'main',
  docsDir: 'docs/content/docs',
};

export const productRepoUrl = `https://github.com/${productRepo.user}/${productRepo.repo}`;

// Community channel on Telegram (announcements & support).
export const telegramChannel = 'mahsa_net';
export const telegramChannelUrl = `https://t.me/${telegramChannel}`;

// Public site origin, used for metadataBase / canonical URLs / OG images.
// The docs are not hosted anywhere by default, so this falls back to the local
// dev server; set NEXT_PUBLIC_SITE_URL when deploying. Use `||` (not `??`) so
// an empty string also falls back instead of shipping a blank origin.
export const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || 'http://localhost:3000';
