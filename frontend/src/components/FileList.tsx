import type { model } from "../../wailsjs/go/models";
import { t, type Locale } from "../i18n";

function statusChip(status: string, locale: Locale) {
  switch (status) {
    case "ok":
      return <span className="chip chip-ok">{t(locale, "statusOk")}</span>;
    case "skipped":
      return <span className="chip chip-skipped">{t(locale, "statusSkipped")}</span>;
    case "failed":
      return <span className="chip chip-failed">{t(locale, "statusFailed")}</span>;
    default:
      return <span className="chip chip-muted">{status}</span>;
  }
}

function formatSize(n: number) {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

type Props = {
  files: model.FileRef[];
  locale: Locale;
  onRemove?: (path: string) => void;
  onClear?: () => void;
};

export function FileList({ files, locale, onRemove, onClear }: Props) {
  const ok = files.filter((f) => f.status === "ok").length;
  const skipped = files.filter((f) => f.status === "skipped").length;
  const failed = files.filter((f) => f.status === "failed").length;

  if (files.length === 0) {
    return (
      <div className="empty-state">
        <strong>{t(locale, "fileListEmpty")}</strong>
        <span>{t(locale, "fileListEmptyDesc")}</span>
      </div>
    );
  }

  return (
    <div>
      <div className="summary-bar">
        <span>{t(locale, "fileCount", { n: files.length })}</span>
        <span className="chip chip-ok">{t(locale, "okCount", { n: ok })}</span>
        {skipped > 0 && (
          <span className="chip chip-skipped">{t(locale, "skippedCount", { n: skipped })}</span>
        )}
        {failed > 0 && (
          <span className="chip chip-failed">{t(locale, "failedCount", { n: failed })}</span>
        )}
        {onClear && (
          <button type="button" className="btn btn-quiet" onClick={onClear}>
            {t(locale, "clearAll")}
          </button>
        )}
      </div>
      <div className="panel">
        <table className="list-table">
          <thead>
            <tr>
              <th>{t(locale, "fileListTitle")}</th>
              <th style={{ width: 88 }}>Size</th>
              <th style={{ width: 100 }}>Status</th>
              {onRemove && <th style={{ width: 72 }} />}
            </tr>
          </thead>
          <tbody>
            {files.map((f) => (
              <tr key={f.path}>
                <td className="col-name">
                  {f.name}
                  {f.reason ? <span className="reason">{f.reason}</span> : null}
                </td>
                <td className="col-num">{formatSize(f.size || 0)}</td>
                <td>{statusChip(f.status, locale)}</td>
                {onRemove && (
                  <td>
                    <button
                      type="button"
                      className="btn btn-quiet"
                      aria-label={t(locale, "removeFile")}
                      onClick={() => onRemove(f.path)}
                    >
                      ×
                    </button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
