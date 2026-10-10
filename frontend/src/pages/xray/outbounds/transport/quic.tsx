import { useTranslation } from 'react-i18next';
import { Input, Select } from 'antd';

import { FormField } from '@/components/form/rhf';
import { QUIC_HEADER_TYPES, QUIC_SECURITIES } from '@/schemas/protocols/stream/quic';

const SECURITY_OPTIONS = QUIC_SECURITIES.map((value) => ({ value, label: value }));
const HEADER_OPTIONS = QUIC_HEADER_TYPES.map((value) => ({ value, label: value }));

export default function QuicForm() {
  const { t } = useTranslation();
  return (
    <>
      <FormField name={['streamSettings', 'quicSettings', 'security']} label={t('encryption')}>
        <Select options={SECURITY_OPTIONS} />
      </FormField>
      <FormField name={['streamSettings', 'quicSettings', 'key']} label={t('password')}>
        <Input />
      </FormField>
      <FormField
        name={['streamSettings', 'quicSettings', 'header', 'type']}
        label={t('camouflage')}
      >
        <Select options={HEADER_OPTIONS} />
      </FormField>
    </>
  );
}
