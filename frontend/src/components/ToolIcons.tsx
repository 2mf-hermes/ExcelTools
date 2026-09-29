type Props = { className?: string };

/** Excel-like workbook/sheet grid icon (green Excel accent). */
export function IconSheets({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      {/* back sheet */}
      <rect x="5" y="3.5" width="12" height="15" rx="1.5" fill="currentColor" opacity="0.28" />
      {/* front sheet */}
      <rect x="7" y="5.5" width="13" height="15.5" rx="1.5" fill="currentColor" opacity="0.55" />
      <rect
        x="4"
        y="4"
        width="13"
        height="16"
        rx="1.75"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
      />
      {/* grid */}
      <path d="M4 9h13M4 13.5h13M9.5 4v16M13.5 4v16" stroke="currentColor" strokeWidth="1.2" opacity="0.9" />
      {/* header band */}
      <path d="M4 4h13v5H4z" fill="currentColor" opacity="0.35" />
    </svg>
  );
}

/** Multiple Excel workbooks stacked with a merge arrow. */
export function IconFiles({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      {/* back workbook */}
      <rect x="3" y="3" width="11" height="14" rx="1.5" fill="currentColor" opacity="0.3" />
      <path d="M3 7h11" stroke="currentColor" strokeWidth="1.2" opacity="0.7" />
      {/* front workbook */}
      <rect
        x="6"
        y="6"
        width="12"
        height="15"
        rx="1.75"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
      />
      <path d="M6 10.5h12M12 6v15" stroke="currentColor" strokeWidth="1.2" />
      <path d="M6 6h12v4.5H6z" fill="currentColor" opacity="0.35" />
      {/* merge arrows into front book */}
      <path d="M20.5 8.5v7M18 13.5l2.5 2.5 2.5-2.5" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function IconChecklist({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <rect x="5" y="3" width="14" height="18" rx="2" fill="none" stroke="currentColor" strokeWidth="1.6" />
      <path d="M8 8l1.5 1.5L12 7" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M8 14l1.5 1.5L12 13" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M14 8.5h3M14 14.5h3" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </svg>
  );
}

export function IconSliders({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <path d="M5 7h14M5 12h14M5 17h14" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <circle cx="9" cy="7" r="2" fill="currentColor" />
      <circle cx="15" cy="12" r="2" fill="currentColor" />
      <circle cx="11" cy="17" r="2" fill="currentColor" />
    </svg>
  );
}

export function IconPlay({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="8" fill="none" stroke="currentColor" strokeWidth="1.6" />
      <path d="M10 9.5v5l4.5-2.5z" fill="currentColor" />
    </svg>
  );
}

export function IconResult({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <rect x="5" y="3" width="14" height="18" rx="2" fill="none" stroke="currentColor" strokeWidth="1.6" />
      <path d="M5 8h14" stroke="currentColor" strokeWidth="1.3" />
      <path d="M8.5 13.5l2.5 2.5 4.5-5" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function IconFolder({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <path
        d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function IconInfo({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="8" fill="none" stroke="currentColor" strokeWidth="1.6" />
      <path d="M12 11v5M12 8.5h.01" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  );
}

export function IconGlobe({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="8" fill="none" stroke="currentColor" strokeWidth="1.6" />
      <path d="M4 12h16M12 4c2.5 2.8 2.5 13.2 0 16M12 4c-2.5 2.8-2.5 13.2 0 16" fill="none" stroke="currentColor" strokeWidth="1.3" />
    </svg>
  );
}

export function IconMoon({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <path d="M16 14.5A6.5 6.5 0 0 1 9.5 8 6.5 6.5 0 1 0 16 14.5z" fill="currentColor" />
    </svg>
  );
}

export function IconRefresh({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <path d="M19 8a7 7 0 1 0 1 6" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
      <path d="M19 4v4h-4" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function IconCloudDown({ className }: Props) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <path
        d="M7 17h10a4 4 0 0 0 .4-8 5.5 5.5 0 0 0-10.7 1.5A3.5 3.5 0 0 0 7 17z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
      />
      <path d="M12 11v5M10 14l2 2 2-2" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
