import { useTranslation } from 'react-i18next';
import { Divider, Input, InputNumber } from 'antd';

import { FormField } from '@/components/form/rhf';

export type TcpLaneSettingsKey = 'quictcpSettings' | 'hytcpSettings';

// An emptied input must drop the key, not store ""/null: the core keeps its own default
// only for an absent key, and the Zod schema refuses null.
const EMPTY_TO_UNDEFINED = {
  output: (v: unknown) => (v === '' || v === null ? undefined : v),
};

interface LaneField {
  key: string;
  label: string;
  placeholder: string;
  range?: boolean;
  min?: number;
  max?: number;
  tooltip?: string;
}

// Placeholders are Mahsa-Core's defaults (infra/conf/transport_finalmask_tcplane.go).
const CLIENT_FIELDS: LaneField[] = [
  { key: 'packets', label: 'Packets per Lane', placeholder: '8-12', range: true },
  { key: 'lifeMs', label: 'Up Lane Life (ms)', placeholder: '800-1200', range: true },
  {
    key: 'bytes',
    label: 'Down Lane Bytes',
    placeholder: '12288-16384',
    range: true,
    tooltip: '-1 turns the per-lane byte budget off',
  },
  { key: 'lanes', label: 'Down Lanes', placeholder: '24', min: 1, max: 64 },
  { key: 'maxRate', label: 'Max Lanes/s', placeholder: '50', min: 0, tooltip: '0 = no limit' },
  { key: 'graceMs', label: 'Grace (ms)', placeholder: '2000', min: 0 },
  { key: 'idleMs', label: 'Idle (ms)', placeholder: '20000', min: 0 },
  { key: 'downLifeMs', label: 'Down Lane Life (ms)', placeholder: '3000', min: 0 },
  { key: 'flushMs', label: 'Flush (ms)', placeholder: '10', min: 0 },
];

const SERVER_FIELDS: LaneField[] = [
  { key: 'maxLifeMs', label: 'Max Lane Life (ms)', placeholder: '60000', min: 0 },
  { key: 'sndbuf', label: 'Send Buffer (bytes)', placeholder: 'OS default', min: 0 },
  {
    key: 'maxLanesPerIP',
    label: 'Max Lanes per IP',
    placeholder: '1024',
    min: 0,
    tooltip: '0 = no limit',
  },
];

export default function TcpLaneFields({
  settingsKey,
  side,
}: {
  settingsKey: TcpLaneSettingsKey;
  side: 'server' | 'client';
}) {
  const { t } = useTranslation();
  const path = (key: string) => ['streamSettings', settingsKey, key];
  const fields = side === 'server' ? SERVER_FIELDS : CLIENT_FIELDS;
  return (
    <>
      <Divider style={{ margin: 0 }}>TCP Lanes</Divider>
      <FormField
        label={t('password')}
        name={path('password')}
        tooltip="Both ends must use the same lane password"
        transform={EMPTY_TO_UNDEFINED}
      >
        <Input />
      </FormField>
      <FormField
        label="Lane Ports"
        name={path('ports')}
        tooltip={
          side === 'server'
            ? 'Extra TCP ports to listen on; empty = the inbound port'
            : "Server ports to spread lanes over; empty = the outbound's port"
        }
        transform={EMPTY_TO_UNDEFINED}
      >
        <Input placeholder="3000-3010,4000" />
      </FormField>
      {fields.map((field) => (
        <FormField
          key={field.key}
          label={field.label}
          name={path(field.key)}
          tooltip={field.tooltip}
          transform={EMPTY_TO_UNDEFINED}
        >
          {field.range ? (
            <Input placeholder={field.placeholder} />
          ) : (
            <InputNumber
              min={field.min}
              max={field.max}
              placeholder={field.placeholder}
              style={{ width: '100%' }}
            />
          )}
        </FormField>
      ))}
    </>
  );
}
