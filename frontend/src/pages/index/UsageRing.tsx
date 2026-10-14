import {
  USAGE_CRIT_COLOR,
  USAGE_CRIT_PERCENT,
  USAGE_WARN_COLOR,
  USAGE_WARN_PERCENT,
} from '@/models/status';

const OK_COLOR = '#52c41a';

// Green below the warn threshold, orange from 80%, red from 90%.
export function usageRingColor(percent: number): string {
  if (percent >= USAGE_CRIT_PERCENT) return USAGE_CRIT_COLOR;
  if (percent >= USAGE_WARN_PERCENT) return USAGE_WARN_COLOR;
  return OK_COLOR;
}

interface UsageRingProps {
  percent: number;
  size?: number;
  strokeWidth?: number;
}

// A small live gauge: the arc length is the usage, the colour is its severity.
export default function UsageRing({ percent, size = 52, strokeWidth = 4 }: UsageRingProps) {
  const clamped = Math.min(100, Math.max(0, Number.isFinite(percent) ? percent : 0));
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const color = usageRingColor(clamped);
  const label = `${clamped.toFixed(0)}%`;

  return (
    <svg
      className="ov-tile-ring"
      width={size}
      height={size}
      viewBox={`0 0 ${size} ${size}`}
      role="img"
      aria-label={label}
      data-level={color === OK_COLOR ? 'ok' : color === USAGE_WARN_COLOR ? 'warn' : 'crit'}
    >
      <title>{label}</title>
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke={color}
        strokeOpacity={0.2}
        strokeWidth={strokeWidth}
      />
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke={color}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeDasharray={circumference}
        strokeDashoffset={circumference * (1 - clamped / 100)}
        transform={`rotate(-90 ${size / 2} ${size / 2})`}
        style={{ transition: 'stroke-dashoffset 0.4s ease, stroke 0.4s ease' }}
      />
      <text
        x="50%"
        y="50%"
        textAnchor="middle"
        dominantBaseline="central"
        fill="currentColor"
        fontSize={size * 0.3}
        fontWeight={600}
        style={{ fontVariantNumeric: 'tabular-nums' }}
      >
        {label}
      </text>
    </svg>
  );
}
