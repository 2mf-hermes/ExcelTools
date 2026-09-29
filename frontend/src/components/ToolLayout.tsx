import type { ReactNode } from "react";

type HeroProps = {
  icon: ReactNode;
  iconClass: string;
  title: string;
  lead?: string;
};

export function ToolHero({ icon, iconClass, title, lead }: HeroProps) {
  return (
    <div className="tool-hero">
      <span className={`tool-icon ${iconClass}`} aria-hidden="true">
        {icon}
      </span>
      <div>
        <h2 className="page-title" style={{ marginBottom: 4 }}>
          {title}
        </h2>
        {lead ? <p className="page-lead" style={{ margin: 0 }}>
          {lead}
        </p> : null}
      </div>
    </div>
  );
}

type SectionProps = {
  icon?: ReactNode;
  label: string;
  children: ReactNode;
  className?: string;
};

export function SectionBlock({ icon, label, children, className }: SectionProps) {
  return (
    <section className={`section-block${className ? ` ${className}` : ""}`}>
      <div className="section-block-head">
        {icon ? (
          <span className="section-block-icon" aria-hidden="true">
            {icon}
          </span>
        ) : null}
        <span className="section-block-label">{label}</span>
      </div>
      {children}
    </section>
  );
}
