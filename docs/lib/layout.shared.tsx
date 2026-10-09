import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { Logo } from '@/components/logo';
import { TelegramIcon } from '@/components/icons';
import { DocsThemeSwitch } from '@/components/theme-switch';
import {
  appName,
  productRepoUrl,
  telegramChannel,
  telegramChannelUrl,
  siteUrl,
} from './shared';
import { getSiteMessages } from './site-i18n';

// Build locale-aware shared layout options. With `hideLocale: 'default-locale'`,
// English URLs have no prefix while other locales are prefixed (`/fa`, `/ru`, `/zh`).
export function baseOptions(lang: string): BaseLayoutProps {
  const prefix = lang === 'en' ? '' : `/${lang}`;
  const m = getSiteMessages(lang);

  return {
    slots: {
      themeSwitch: DocsThemeSwitch,
    },
    nav: {
      title: (
        <span className="inline-flex items-center gap-2 font-semibold">
          <Logo className="h-6" />
          {appName}
        </span>
      ),
      url: `${prefix}/`,
    },
    githubUrl: productRepoUrl,
    links: [
      {
        text: m.documentation,
        url: `${prefix}/docs`,
        active: 'nested-url',
      },
      // Live component workbench built from frontend/ and published alongside the docs.
      {
        text: 'Storybook',
        url: `${siteUrl}/storybook/`,
        external: true,
      },
      {
        type: 'icon',
        label: `Telegram channel (@${telegramChannel})`,
        icon: <TelegramIcon />,
        text: 'Telegram',
        url: telegramChannelUrl,
        external: true,
      },
    ],
  };
}
