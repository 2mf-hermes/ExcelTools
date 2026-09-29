import type { model } from "../../wailsjs/go/models";
import { t, type Locale } from "../i18n";

type Props = {
  sheets: model.SheetRef[];
  selected: Record<string, boolean>;
  loading?: boolean;
  locale: Locale;
  onToggle: (name: string) => void;
  onSelectAll: (on: boolean) => void;
  onMove?: (name: string, dir: -1 | 1) => void;
  order: string[];
};

export function SheetList({
  sheets,
  selected,
  loading,
  locale,
  onToggle,
  onSelectAll,
  onMove,
  order,
}: Props) {
  if (loading) {
    return <div className="loading-row">{t(locale, "loadingSheets")}</div>;
  }
  if (sheets.length === 0) {
    return (
      <div className="empty-state">
        <strong>{t(locale, "sheetListEmpty")}</strong>
      </div>
    );
  }

  const byName = new Map(sheets.map((s) => [s.name, s]));
  const names = order.length ? order.filter((n) => byName.has(n)) : sheets.map((s) => s.name);

  return (
    <div>
      <div className="summary-bar">
        <button type="button" className="btn btn-quiet" onClick={() => onSelectAll(true)}>
          {t(locale, "selectAll")}
        </button>
        <button type="button" className="btn btn-quiet" onClick={() => onSelectAll(false)}>
          {t(locale, "selectNone")}
        </button>
      </div>
      <div className="panel">
        <table className="list-table">
          <thead>
            <tr>
              <th style={{ width: 36 }} />
              <th>{t(locale, "sheetStepSheets")}</th>
              <th style={{ width: 70 }}>{t(locale, "rows")}</th>
              <th style={{ width: 70 }}>{t(locale, "cols")}</th>
              <th style={{ width: 100 }} />
            </tr>
          </thead>
          <tbody>
            {names.map((name, idx) => {
              const s = byName.get(name)!;
              return (
                <tr key={name}>
                  <td>
                    <input
                      className="sheet-check"
                      type="checkbox"
                      checked={!!selected[name]}
                      onChange={() => onToggle(name)}
                      aria-label={name}
                    />
                  </td>
                  <td className="col-name">
                    {name}
                    {s.empty ? (
                      <span className="chip chip-muted" style={{ marginLeft: 8 }}>
                        {t(locale, "emptySheet")}
                      </span>
                    ) : null}
                    {s.hidden ? (
                      <span className="chip chip-muted" style={{ marginLeft: 8 }}>
                        {t(locale, "hiddenSheet")}
                      </span>
                    ) : null}
                  </td>
                  <td className="col-num">{s.rows}</td>
                  <td className="col-num">{s.cols}</td>
                  <td className="col-num">
                    {onMove && (
                      <span className="btn-row">
                        <button
                          type="button"
                          className="btn btn-quiet"
                          disabled={idx === 0}
                          onClick={() => onMove(name, -1)}
                          aria-label="up"
                        >
                          ↑
                        </button>
                        <button
                          type="button"
                          className="btn btn-quiet"
                          disabled={idx === names.length - 1}
                          onClick={() => onMove(name, 1)}
                          aria-label="down"
                        >
                          ↓
                        </button>
                      </span>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
