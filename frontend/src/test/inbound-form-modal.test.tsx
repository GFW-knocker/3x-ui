import { describe, it, expect, onTestFinished, vi } from 'vitest';
import { screen, act, render, cleanup, fireEvent, waitFor } from '@testing-library/react';

import InboundFormModal from '@/pages/inbounds/form/InboundFormModal';
import { DBInbound } from '@/models/dbinbound';
import { ThemeProvider } from '@/hooks/useTheme';
import { HttpUtil } from '@/utils';
import {
  renderWithProviders,
  fieldLabels,
  listSelectOptions,
  chooseSelectOption,
} from './test-utils';

const { messageError } = vi.hoisted(() => ({ messageError: vi.fn() }));

vi.mock('antd', async (importOriginal) => {
  const actual = await importOriginal<typeof import('antd')>();
  return {
    ...actual,
    message: {
      ...actual.message,
      useMessage: () => [{ error: messageError }, null],
    },
  };
});

function renderModal() {
  return renderWithProviders(
    <InboundFormModal
      open
      mode="add"
      dbInbound={null}
      dbInbounds={[]}
      availableNodes={[]}
      onClose={() => {}}
      onSaved={() => {}}
    />,
  );
}

function primaryButton(): HTMLElement {
  const button = document.querySelector('.ant-modal-footer .ant-btn-primary');
  if (!button) throw new Error('Primary modal button not found');
  return button as HTMLElement;
}

function cloneLikeVlessInbound(target: string) {
  return new DBInbound({
    id: 42,
    port: 41234,
    listen: '',
    protocol: 'vless',
    remark: 'source clone',
    enable: false,
    settings: {
      clients: [],
      decryption: 'none',
      encryption: 'none',
      fallbacks: [],
    },
    streamSettings: {
      network: 'tcp',
      security: 'reality',
      tcpSettings: { header: { type: 'none' } },
      realitySettings: {
        target,
        serverNames: ['example.com'],
        privateKey: 'test-private-key',
        shortIds: ['abcd'],
        settings: {
          publicKey: 'test-public-key',
          fingerprint: 'chrome',
          spiderX: '/',
        },
      },
    },
    sniffing: { enabled: false },
    nodeId: null,
    shareAddrStrategy: 'listen',
    shareAddr: '',
  });
}

function renderCloneLikeEdit(dbInbound: DBInbound) {
  renderWithProviders(
    <InboundFormModal
      open
      mode="edit"
      dbInbound={dbInbound}
      dbInbounds={[dbInbound]}
      availableNodes={[]}
      onClose={() => {}}
      onSaved={() => {}}
    />,
  );
}

// The transmission picker is a bare Select inside a labelled Form.Item, so it has no field id.
function selectInFormItem(label: string): HTMLElement {
  const item = Array.from(document.querySelectorAll('.ant-form-item')).find(
    (el) => (el.querySelector('.ant-form-item-label label')?.textContent ?? '').trim() === label,
  );
  const select = item?.querySelector('.ant-select') as HTMLElement | null;
  if (!select) throw new Error(`${label} select not found`);
  return select;
}

const optionTitle = (o: Element) => (o.getAttribute('title') ?? o.textContent ?? '').trim();

// Every Select shares one test id and closed popups linger in jsdom, so the transport list
// is the one open popup that carries RAW; anything else would read a stale list.
function openTransportOptions(): Element[] {
  const select = selectInFormItem('Transmission');
  fireEvent.mouseDown((select.querySelector('.ant-select-selector') ?? select) as HTMLElement);
  const popups = Array.from(
    document.querySelectorAll('.ant-select-dropdown:not(.ant-select-dropdown-hidden)'),
  ).filter((popup) =>
    Array.from(popup.querySelectorAll('.ant-select-item-option')).some(
      (o) => optionTitle(o) === 'RAW',
    ),
  );
  if (popups.length !== 1)
    throw new Error(`expected one open transport list, got ${popups.length}`);
  return Array.from(popups[0].querySelectorAll('.ant-select-item-option'));
}

function transportOptions(): string[] {
  const options = openTransportOptions().map(optionTitle);
  fireEvent.keyDown(selectInFormItem('Transmission'), { key: 'Escape' });
  return options;
}

function pickTransport(optionText: string) {
  const option = openTransportOptions().find((o) => optionTitle(o) === optionText);
  if (!option) throw new Error(`Transport '${optionText}' not offered`);
  fireEvent.click(option);
}

async function flushUi() {
  await act(async () => {
    await new Promise((r) => setTimeout(r, 0));
  });
}

describe('InboundFormModal', () => {
  it('renders add mode without crashing', () => {
    renderModal();
    expect(document.querySelector('.ant-modal')).toBeTruthy();
    expect(fieldLabels().length).toBeGreaterThan(0);
  });

  it('field structure differs per protocol (not a vacuous snapshot loop)', async () => {
    renderModal();
    const protocols = listSelectOptions('protocol');
    expect(protocols.length).toBeGreaterThan(3);

    const labelsByProto: Record<string, string[]> = {};
    for (const proto of protocols) {
      chooseSelectOption('protocol', proto);
      // Flush antd Form.useWatch('protocol') before reading — without it every iteration
      // sees the same pre-update DOM and the loop asserts nothing (the original bug here).
      await act(async () => {
        await new Promise((r) => setTimeout(r, 0));
      });
      labelsByProto[proto] = fieldLabels();
    }

    // The loop must actually exercise protocol-specific rendering: distinct protocols
    // must yield distinct field sets (a vacuous loop makes them all identical).
    const distinctShapes = new Set(Object.values(labelsByProto).map((l) => l.join('|')));
    expect(distinctShapes.size).toBeGreaterThan(1);

    // Spot-check a protocol-distinguishing field that must appear after the switch.
    if (labelsByProto.shadowsocks) {
      expect(labelsByProto.shadowsocks).toContain('Encryption method');
    }
  }, 30000); // iterates every protocol, re-rendering a heavy modal each time — slow on CI runners

  it('preserves custom share address strategy when editing a local inbound', async () => {
    renderWithProviders(
      <InboundFormModal
        open
        mode="edit"
        dbInbound={
          new DBInbound({
            id: 1,
            port: 12345,
            listen: '',
            protocol: 'shadowsocks',
            remark: 'edge',
            enable: true,
            settings: {
              method: '2022-blake3-aes-128-gcm',
              password: 'server-password',
              network: 'tcp,udp',
              clients: [],
            },
            streamSettings: { network: 'tcp', security: 'none', tcpSettings: {} },
            sniffing: { enabled: false },
            nodeId: null,
            shareAddrStrategy: 'custom',
            shareAddr: 'edge.example.test',
          })
        }
        dbInbounds={[]}
        availableNodes={[]}
        onClose={() => {}}
        onSaved={() => {}}
      />,
    );

    const shareAddrInput = await screen.findByDisplayValue('edge.example.test');
    expect((shareAddrInput as HTMLInputElement).value).toBe('edge.example.test');
  });

  it('uses Hosts instead of showing the custom share address fields for MTProto', async () => {
    renderWithProviders(
      <InboundFormModal
        open
        mode="edit"
        dbInbound={
          new DBInbound({
            id: 2,
            port: 4060,
            listen: '',
            protocol: 'mtproto',
            remark: 'proxy',
            enable: true,
            settings: { clients: [] },
            streamSettings: {},
            sniffing: { enabled: false },
            nodeId: null,
            shareAddrStrategy: 'custom',
            shareAddr: 'proxy.example.test',
          })
        }
        dbInbounds={[]}
        availableNodes={[]}
        onClose={() => {}}
        onSaved={() => {}}
      />,
    );

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 0));
    });
    expect(fieldLabels()).not.toContain('Share address strategy');
    expect(screen.queryByDisplayValue('proxy.example.test')).toBeNull();
  });

  it('keeps the persisted node share strategy through the nodes-loading race (#5375)', async () => {
    const node = { id: 1, name: 'arm2', enable: true, status: 'online' } as never;
    const buildInbound = () =>
      new DBInbound({
        id: 1,
        port: 23456,
        listen: '',
        protocol: 'vless',
        remark: 'noded',
        enable: true,
        settings: { clients: [] },
        streamSettings: { network: 'tcp', security: 'none', tcpSettings: {} },
        sniffing: { enabled: false },
        nodeId: 1,
        shareAddrStrategy: 'node',
      });
    const flush = async () => {
      await act(async () => {
        await new Promise((r) => setTimeout(r, 0));
      });
    };
    const strategyItem = (title: string) =>
      document.querySelector(`.ant-select-content[title="${title}"]`);
    const modal = (nodes: never[], fetched: boolean) => (
      <ThemeProvider>
        <InboundFormModal
          open
          mode="edit"
          dbInbound={buildInbound()}
          dbInbounds={[]}
          availableNodes={nodes}
          availableNodesFetched={fetched}
          onClose={() => {}}
          onSaved={() => {}}
        />
      </ThemeProvider>
    );

    // Baseline: nodes already loaded, so the node option is offered and selected.
    render(modal([node], true));
    await flush();
    expect(strategyItem('Node address')).toBeTruthy();
    cleanup();

    // Race: the modal mounts before /nodes/list resolves (empty placeholder),
    // then nodes arrive. The persisted 'node' strategy must survive the gap and
    // stay selected once the option reappears — not silently revert to listen.
    const { rerender } = render(modal([], false));
    await flush();
    rerender(modal([node], true));
    await flush();
    expect(strategyItem('Node address')).toBeTruthy();
    expect(strategyItem('Inbound listen')).toBeFalsy();
  });

  it('surfaces a Reality validation error and switches to its tab', async () => {
    const post = vi.mocked(HttpUtil.post);
    post.mockClear();
    messageError.mockClear();
    renderCloneLikeEdit(cloneLikeVlessInbound('example.com'));

    fireEvent.click(primaryButton());

    await waitFor(() => {
      const securityTab = screen.getByRole('tab', { name: 'Security' });
      expect(securityTab.getAttribute('aria-selected')).toBe('true');
    });
    expect(messageError).toHaveBeenCalledWith(
      expect.stringContaining('REALITY target must include a port'),
    );
    expect(post).not.toHaveBeenCalled();
  });

  it('blocks adding TLS without a certificate and directs the user to Security', async () => {
    const post = vi.mocked(HttpUtil.post);
    post.mockClear();
    messageError.mockClear();
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
    onTestFinished(() => consoleError.mockRestore());
    renderModal();

    fireEvent.click(screen.getByRole('tab', { name: 'Security' }));
    fireEvent.click(screen.getByRole('radio', { name: 'TLS' }));
    fireEvent.click(screen.getByRole('tab', { name: 'Basics' }));
    fireEvent.click(primaryButton());

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: 'Security' }).getAttribute('aria-selected')).toBe(
        'true',
      );
      expect(messageError).toHaveBeenCalledWith(
        expect.stringContaining('TLS certificate 1: Import a TLS certificate'),
      );
    });
    expect(consoleError).toHaveBeenCalledWith('[InboundFormModal] schema validation failed:', [
      'TLS certificate 1: Import a TLS certificate or enter its file path before saving',
      'TLS certificate 1: Import the TLS private key or enter its file path before saving',
    ]);
    expect(post).not.toHaveBeenCalled();
  });

  it('submits a valid clone-like Reality inbound', async () => {
    const post = vi.mocked(HttpUtil.post);
    post.mockClear();
    renderCloneLikeEdit(cloneLikeVlessInbound('example.com:443'));

    fireEvent.click(primaryButton());

    await waitFor(() => {
      expect(post).toHaveBeenCalledWith(
        '/panel/api/inbounds/update/42',
        expect.objectContaining({ enable: false, port: 41234, protocol: 'vless' }),
      );
    });
  });

  // Clients and enable change through their own endpoints; the server keeps the
  // stored ones, so the edit form must neither send nor validate its stale copy.
  it('edit save neither sends nor validates the clients it loaded', async () => {
    const post = vi.mocked(HttpUtil.post);
    post.mockClear();
    const dbInbound = cloneLikeVlessInbound('example.com:443');
    const legacy = new DBInbound({
      ...dbInbound,
      settings: {
        ...(dbInbound.settings as Record<string, unknown>),
        clients: [{ email: 'legacy', id: '' }],
      },
    });
    renderCloneLikeEdit(legacy);

    fireEvent.click(primaryButton());

    await waitFor(() => expect(post).toHaveBeenCalled());
    const payload = post.mock.calls[0][1] as { settings: string };
    expect(JSON.parse(payload.settings)).not.toHaveProperty('clients');
  });

  it('offers hytcp to vless, vmess and trojan but not to shadowsocks', async () => {
    renderModal();
    expect(transportOptions()).toContain('Hysteria over TCP lanes');
    for (const protocol of ['vmess', 'trojan']) {
      chooseSelectOption('protocol', protocol);
      await flushUi();
      expect(transportOptions()).toContain('Hysteria over TCP lanes');
    }
    chooseSelectOption('protocol', 'shadowsocks');
    await flushUi();
    expect(transportOptions()).not.toContain('Hysteria over TCP lanes');
  }, 20000);

  // The hysteria listener serves HTTP/3 behind TLS and, outside the hysteria protocol,
  // checks its own auth; any of the three missing and the core cannot start.
  it('moves a vless inbound onto hytcp with TLS on h3 and a generated transport password', async () => {
    const post = vi.mocked(HttpUtil.post);
    post.mockClear();
    renderModal();
    pickTransport('Hysteria over TCP lanes');
    await flushUi();

    const auth = (screen.getByLabelText('Transport password') as HTMLInputElement).value;
    expect(auth.length).toBeGreaterThanOrEqual(16);
    fireEvent.click(screen.getByRole('tab', { name: 'Security' }));
    await flushUi();
    // jsdom leaves the antd radio input unchecked; the wrapper carries the selection.
    const tlsRadio = screen.getByRole('radio', { name: 'TLS' }).closest('label');
    expect(tlsRadio?.classList.contains('ant-radio-button-wrapper-checked')).toBe(true);
    expect(screen.queryByRole('radio', { name: 'None' })).toBeNull();

    fireEvent.change(screen.getAllByLabelText('Public Key')[0], { target: { value: '/c.crt' } });
    fireEvent.change(screen.getAllByLabelText('Private Key')[0], { target: { value: '/c.key' } });
    fireEvent.click(primaryButton());

    const addCall = () => post.mock.calls.find(([url]) => url === '/panel/api/inbounds/add');
    await waitFor(() => expect(addCall()).toBeDefined());
    const stream = JSON.parse((addCall()![1] as { streamSettings: string }).streamSettings);
    expect(stream).toMatchObject({ network: 'hytcp', security: 'tls', hytcpSettings: { auth } });
    expect(stream.tlsSettings.alpn).toEqual(['h3']);
  }, 20000);

  it('refuses to save a vless hytcp inbound without a transport password', async () => {
    const post = vi.mocked(HttpUtil.post);
    post.mockClear();
    messageError.mockClear();
    renderModal();
    pickTransport('Hysteria over TCP lanes');
    await flushUi();

    fireEvent.change(screen.getByLabelText('Transport password'), { target: { value: '' } });
    fireEvent.click(primaryButton());

    await waitFor(() => {
      expect(messageError).toHaveBeenCalledWith(expect.stringContaining('transport password'));
      expect(screen.getByRole('tab', { name: 'Stream' }).getAttribute('aria-selected')).toBe(
        'true',
      );
    });
    expect(post).not.toHaveBeenCalled();
  }, 20000);

  it('refuses to save more lane ports than the core will listen on', async () => {
    const post = vi.mocked(HttpUtil.post);
    post.mockClear();
    messageError.mockClear();
    renderModal();
    pickTransport('QUIC over TCP lanes');
    await flushUi();

    fireEvent.change(screen.getByLabelText('Lane Ports'), { target: { value: '20000-21024' } });
    fireEvent.click(primaryButton());

    await waitFor(() => {
      expect(messageError).toHaveBeenCalledWith(expect.stringContaining('1024'));
    });
    expect(post).not.toHaveBeenCalledWith('/panel/api/inbounds/add', expect.anything());
  }, 20000);

  it('offers the enable switch when adding an inbound but not when editing one', () => {
    renderModal();
    expect(document.getElementById('inbound-enable')).not.toBeNull();
    cleanup();
    renderCloneLikeEdit(cloneLikeVlessInbound('example.com:443'));
    expect(document.getElementById('inbound-enable')).toBeNull();
  });
});
