export function FlowdLogo({ size = 40 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="none"
      aria-hidden="true"
    >
      <defs>
        <linearGradient x1="10" y1="54" x2="54" y2="10">
          <stop stopColor="var(--primary)" />
          <stop offset=".55" stopColor="var(--accent)" />
          <stop offset="1" stopColor="var(--warning)" />
        </linearGradient>
      </defs>
      <path
        d="M17 41c9 0 10-18 22-18 5 0 8 3 8 8 0 12-15 18-25 9"
        stroke="var(--primary)"
        strokeWidth="1"
        strokeLinecap="round"
      />
      <circle cx="17" cy="41" r="5" fill="var(--primary)" />
      <circle cx="47" cy="31" r="5" fill="var(--warning)" />
    </svg>
  );
}
