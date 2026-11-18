import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  CancelFileMerge,
  ClassifyPaths,
  MergeFiles,
  PickExcelFiles,
  PickFolder,
  RevealInExplorer,
  ScanFolder,
} from "../../wailsjs/go/main/App";
import { main, type model } from "../../wailsjs/go/models";
import { DropZone } from "../components/DropZone";
import { FileList } from "../components/FileList";
import { SectionBlock, ToolHero } from "../components/ToolLayout";
import {
  IconChecklist,
  IconFiles,
  IconFolder,
  IconPlay,
  IconResult,
  IconSliders,
} from "../components/ToolIcons";
import { t, type Locale } from "../i18n";
import { EventsOn, OnFileDrop, OnFileDropOff } from "../../wailsjs/runtime/runtime";
import { useSheetLists, type SheetMap } from "./useSheetLists";

type Phase = "input" | "running" | "result";
type ColumnAlign = "by-header" | "by-position" | "abort";
/** Mirrors files.SheetMode on the Go side. */
type SheetMode = "auto" | "index";
type ProgressEvent = {
  phase: string;
  current: number;
  total: number;
  label: string;
  percent: number;
};

type Props = { locale: Locale };

function mergeRefs(prev: model.FileRef[], next: model.FileRef[]): model.FileRef[] {
  const map = new Map<string, model.FileRef>();
  for (const f of prev) map.set(f.path, f);
  for (const f of next) map.set(f.path, f);
  return Array.from(map.values());
}

/** "data · #3 · 120 rows" style label for one pinned / loaded sheet. */
function sheetCaption(s: model.SheetRef, index: number, locale: Locale): string {
  const flags: string[] = [];
  if (s.empty) flags.push(t(locale, "emptySheet"));
  if (s.hidden) flags.push(t(locale, "hiddenSheet"));
  const base = `${s.name} · #${index} · ${s.rows} ${t(locale, "rows")}`;
  return flags.length ? `${base} · ${flags.join(" · ")}` : base;
}

export function FileCombinePage({ locale }: Props) {
  const [phase, setPhase] = useState<Phase>("input");
  const [files, setFiles] = useState<model.FileRef[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [folderPath, setFolderPath] = useState("");

  const [headerRows, setHeaderRows] = useState(1);
  const [sheetMode, setSheetMode] = useState<SheetMode>("auto");
  const [sheetIndex, setSheetIndex] = useState(1);
  const [columnAlign, setColumnAlign] = useState<ColumnAlign>("by-header");
  const [addSourceFile, setAddSourceFile] = useState(true);
  const [keepBaseOtherSheets, setKeepBaseOtherSheets] = useState(true);
  const [keepBlankRows, setKeepBlankRows] = useState(true);

  // path -> sheets (loaded), path -> error text (unreadable)
  const [sheetsByPath, setSheetsByPath] = useState<SheetMap>({});
  const [sheetErrors, setSheetErrors] = useState<Record<string, string>>({});
  // path -> 1-based sheet position pinned by the user; absent = follow the rule
  const [pins, setPins] = useState<Record<string, number>>({});

  const [progress, setProgress] = useState<ProgressEvent | null>(null);
  const [result, setResult] = useState<model.MergeResult | null>(null);
  const [running, setRunning] = useState(false);

  const addPaths = useCallback(async (paths: string[]) => {
    const absolute = (paths || []).filter((p) => /^[a-zA-Z]:[\\/]/.test(p) || p.startsWith("\\\\") || p.startsWith("/"));
    if (!absolute.length) {
      setError(t(locale, "errorFormat"));
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const refs = await ClassifyPaths(absolute);
      setFiles((prev) => mergeRefs(prev, refs));
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }, [locale]);

  useEffect(() => {
    try {
      EventsOn("filemerge:progress", (data) => {
        const ev = data as ProgressEvent;
        if (ev) setProgress(ev);
      });
    } catch {
      /* browser without wails */
    }
  }, []);

  const phaseRef = useRef(phase);
  phaseRef.current = phase;
  const addPathsRef = useRef(addPaths);
  addPathsRef.current = addPaths;

  useEffect(() => {
    try {
      OnFileDrop((_x, _y, paths) => {
        if (phaseRef.current === "input" && paths?.length) {
          void addPathsRef.current(paths);
        }
      }, true);
    } catch {
      /* ignore */
    }
    return () => {
      try {
        OnFileDropOff();
      } catch {
        /* ignore */
      }
    };
  }, []);

  const onPickFiles = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const refs = await PickExcelFiles();
      if (refs?.length) setFiles((prev) => mergeRefs(prev, refs));
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }, []);

  const onPickAndScanFolder = useCallback(async () => {
    setBusy(true);
    setError(null);
    try {
      const dir = await PickFolder();
      if (!dir) {
        return;
      }
      setFolderPath(dir);
      const refs = await ScanFolder(dir);
      if (refs?.length) setFiles(refs);
      else setError(t(locale, "errorEmpty"));
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }, [locale]);

  const onScanPath = useCallback(async () => {
    if (!folderPath.trim()) return;
    setBusy(true);
    setError(null);
    try {
      const refs = await ScanFolder(folderPath.trim());
      if (refs?.length) setFiles(refs);
      else setError(t(locale, "errorEmpty"));
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }, [folderPath, locale]);

  const onRemove = (path: string) => {
    setFiles((prev) => prev.filter((f) => f.path !== path));
    setPins((prev) => {
      if (!(path in prev)) return prev;
      const next = { ...prev };
      delete next[path];
      return next;
    });
  };

  const okPaths = useMemo(
    () => files.filter((f) => f.status === "ok").map((f) => f.path),
    [files],
  );
  const hasXls = files.some((f) => f.status === "ok" && f.ext === ".xls");

  const onSheetsLoaded = useCallback(
    (path: string, sheets: model.SheetRef[] | null, err?: string) => {
      if (sheets) {
        setSheetsByPath((prev) => ({ ...prev, [path]: sheets }));
      } else {
        setSheetErrors((prev) => ({ ...prev, [path]: err ?? "?" }));
      }
    },
    [],
  );
  useSheetLists(okPaths, sheetsByPath, onSheetsLoaded);

  /** The sheet a file contributes under the current uniform rule. */
  const followedIndex = useCallback(
    (sheets: model.SheetRef[]): number => {
      if (sheetMode === "index") return sheetIndex;
      const first = sheets.findIndex((s) => !s.empty && !s.hidden);
      return first >= 0 ? first + 1 : 1;
    },
    [sheetMode, sheetIndex],
  );

  const resultSheets = useMemo(() => {
    const map = new Map<string, model.FileSheet>();
    for (const s of result?.sheets ?? []) map.set(s.path, s);
    return map;
  }, [result]);

  const onRun = useCallback(async () => {
    if (okPaths.length === 0) {
      setError(t(locale, "errorEmpty"));
      return;
    }
    setRunning(true);
    setError(null);
    setResult(null);
    setProgress({ phase: "start", current: 0, total: 1, label: "", percent: 0 });
    setPhase("running");
    try {
      const overrides = Object.entries(pins)
        .filter(([path]) => okPaths.includes(path))
        .map(([path, idx]) => ({ path, name: "", index: idx }));
      const res = await MergeFiles(
        main.MergeFilesRequest.createFrom({
          paths: okPaths,
          headerRows,
          sheetMode,
          sheetIndex,
          sheetOverrides: overrides,
          columnAlign,
          addSourceFile,
          keepBaseOtherSheets,
          keepBlankRows,
          outputDir: "",
        }),
      );
      setResult(res);
      setPhase("result");
    } catch (e) {
      const msg = String(e);
      if (msg.includes("E_CANCEL")) {
        setError(t(locale, "cancelled"));
      } else {
        setError(msg);
      }
      setPhase("input");
    } finally {
      setRunning(false);
      setProgress(null);
    }
  }, [okPaths, headerRows, sheetMode, sheetIndex, pins, columnAlign, addSourceFile, keepBaseOtherSheets, keepBlankRows, locale]);

  const onCancel = useCallback(async () => {
    try {
      await CancelFileMerge();
    } catch {
      /* ignore */
    }
  }, []);

  const sheetCell = (path: string) => {
    const sheets = sheetsByPath[path];
    if (!sheets) {
      const err = sheetErrors[path];
      return err ? <span className="reason">{err}</span> : <span className="reason">{t(locale, "sheetsLoading")}</span>;
    }
    const pinned = pins[path];
    const idx = pinned && pinned <= sheets.length ? pinned : followedIndex(sheets);
    const s = sheets[idx - 1];
    if (!s) return <span className="reason">—</span>;
    return (
      <span className="reason">
        {pinned ? "📌 " : ""}
        {s.name} · #{idx}
      </span>
    );
  };

  const resultSheetCell = (path: string) => {
    const used = resultSheets.get(path);
    if (!used) return <span className="reason">—</span>;
    return (
      <span className="reason">
        {used.sheet} · #{used.index}
      </span>
    );
  };

  return (
    <div>
      <ToolHero
        icon={<IconFiles />}
        iconClass="tool-icon-files"
        title={t(locale, "toolFileTitle")}
        lead={t(locale, "toolFileDesc")}
      />

      {phase === "input" && (
        <>
          <SectionBlock icon={<IconChecklist />} label={t(locale, "fileStepInput")}>
          <DropZone
            title={t(locale, "fileDropTitle")}
            hint={t(locale, "fileDropHint")}
            multiple
            disabled={busy}
            onPaths={(paths) => void addPaths(paths)}
            onPickClick={() => void onPickFiles()}
          >
            <div className="btn-row" style={{ marginTop: "var(--space-2)" }}>
              <button
                type="button"
                className="btn"
                disabled={busy}
                onClick={(e) => {
                  e.stopPropagation();
                  void onPickFiles();
                }}
              >
                {t(locale, "chooseFiles")}
              </button>
              <button
                type="button"
                className="btn"
                disabled={busy}
                onClick={(e) => {
                  e.stopPropagation();
                  void onPickAndScanFolder();
                }}
              >
                {t(locale, "chooseFolder")}
              </button>
            </div>
          </DropZone>
          </SectionBlock>

          {error && (
            <div className="alert alert-error" role="alert" style={{ marginTop: "var(--space-4)" }}>
              <strong>{t(locale, "errorTitle")}</strong>
              {error}
            </div>
          )}
          {hasXls && (
            <div className="alert alert-warning" style={{ marginTop: "var(--space-4)" }}>
              {t(locale, "xlsNote")}
            </div>
          )}

          <SectionBlock icon={<IconFolder />} label={t(locale, "fileListTitle")}>
            <div className="path-row" style={{ marginBottom: "var(--space-3)" }}>
              <button
                type="button"
                className="path-field"
                disabled={busy}
                onClick={() => void onPickAndScanFolder()}
                title={t(locale, "chooseFolder")}
                aria-label={t(locale, "folderPath")}
              >
                <span className={`path-text${folderPath ? "" : " is-placeholder"}`}>
                  {folderPath || t(locale, "folderPathPlaceholder")}
                </span>
                <span className="path-browse">…</span>
              </button>
              <button
                type="button"
                className="btn"
                disabled={busy}
                onClick={() => void onPickAndScanFolder()}
              >
                {t(locale, "browse")}
              </button>
              <button
                type="button"
                className="btn"
                disabled={busy || !folderPath.trim()}
                onClick={() => void onScanPath()}
              >
                {t(locale, "scanFolder")}
              </button>
            </div>
            <FileList
              files={files}
              locale={locale}
              onRemove={onRemove}
              onClear={() => {
                setFiles([]);
                setPins({});
              }}
              sheetLabel={sheetCell}
            />
          </SectionBlock>

          <SectionBlock icon={<IconSliders />} label={t(locale, "options")}>
            <div className="panel">
            <label className="field-row">
              <span>{t(locale, "headerRows")}</span>
              <select
                className="select-control"
                value={headerRows}
                onChange={(e) => setHeaderRows(Number(e.target.value))}
              >
                <option value={0}>{t(locale, "headerRows0")}</option>
                <option value={1}>{t(locale, "headerRows1")}</option>
                <option value={2}>{t(locale, "headerRows2")}</option>
                <option value={3}>{t(locale, "headerRows3")}</option>
              </select>
            </label>
            <div className="field-row">
              <span>{t(locale, "sheetMode")}</span>
              <select
                className="select-control"
                value={sheetMode}
                onChange={(e) => setSheetMode(e.target.value as SheetMode)}
                aria-label={t(locale, "sheetMode")}
              >
                <option value="auto">{t(locale, "sheetModeAuto")}</option>
                <option value="index">{t(locale, "sheetModeIndex")}</option>
              </select>
              {sheetMode === "index" && (
                <input
                  className="select-control"
                  type="number"
                  min={1}
                  max={255}
                  value={sheetIndex}
                  aria-label={t(locale, "sheetIndex")}
                  onChange={(e) => setSheetIndex(Math.max(1, Number(e.target.value) || 1))}
                />
              )}
              <span className="reason">{t(locale, "sheetModeDesc")}</span>
            </div>
            <label className="field-row">
              <span>{t(locale, "columnAlign")}</span>
              <select
                className="select-control"
                value={columnAlign}
                onChange={(e) => setColumnAlign(e.target.value as ColumnAlign)}
              >
                <option value="by-header">{t(locale, "alignByHeader")}</option>
                <option value="by-position">{t(locale, "alignByPosition")}</option>
                <option value="abort">{t(locale, "alignAbort")}</option>
              </select>
              <span className="reason">{t(locale, "columnAlignDesc")}</span>
            </label>
            <label className="field-row checkbox-row">
              <input
                type="checkbox"
                className="sheet-check"
                checked={addSourceFile}
                onChange={(e) => setAddSourceFile(e.target.checked)}
              />
              <span>
                {t(locale, "sourceFileCol")}
                <span className="reason">{t(locale, "sourceFileColDesc")}</span>
              </span>
            </label>
            <label className="field-row checkbox-row">
              <input
                type="checkbox"
                className="sheet-check"
                checked={keepBaseOtherSheets}
                onChange={(e) => setKeepBaseOtherSheets(e.target.checked)}
              />
              <span>
                {t(locale, "keepBaseSheets")}
                <span className="reason">{t(locale, "keepBaseSheetsDesc")}</span>
              </span>
            </label>
            <label className="field-row checkbox-row">
              <input
                type="checkbox"
                className="sheet-check"
                checked={keepBlankRows}
                onChange={(e) => setKeepBlankRows(e.target.checked)}
              />
              <span>
                {t(locale, "keepBlankRows")}
                <span className="reason">{t(locale, "keepBlankRowsDesc")}</span>
              </span>
            </label>
            </div>
          </SectionBlock>

          {okPaths.length > 0 && (
            <SectionBlock icon={<IconFiles />} label={t(locale, "perFileSheets")}>
              <div className="panel">
                <p className="footer-note" style={{ marginTop: 0 }}>
                  {t(locale, "perFileSheetsDesc")}
                </p>
                <table className="list-table">
                  <thead>
                    <tr>
                      <th>{t(locale, "fileListTitle")}</th>
                      <th style={{ width: 360 }}>{t(locale, "sheetStepSheets")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {okPaths.map((path) => {
                      const sheets = sheetsByPath[path];
                      const file = files.find((f) => f.path === path);
                      const err = sheetErrors[path];
                      return (
                        <tr key={path}>
                          <td className="col-name">{file?.name ?? path}</td>
                          <td>
                            {err ? (
                              <span className="reason">{err}</span>
                            ) : !sheets ? (
                              <span className="reason">{t(locale, "sheetsLoading")}</span>
                            ) : sheets.length === 0 ? (
                              <span className="reason">{t(locale, "sheetListEmpty")}</span>
                            ) : (
                              <select
                                className="select-control"
                                value={pins[path] ?? ""}
                                aria-label={`${file?.name ?? path} ${t(locale, "sheetStepSheets")}`}
                                onChange={(e) => {
                                  const v = e.target.value;
                                  setPins((prev) => {
                                    const next = { ...prev };
                                    if (v === "") delete next[path];
                                    else next[path] = Number(v);
                                    return next;
                                  });
                                }}
                              >
                                <option value="">
                                  {t(locale, "perFileFollow")} · #{followedIndex(sheets)}
                                </option>
                                {sheets.map((s, i) => (
                                  <option key={`${s.name}-${i}`} value={i + 1}>
                                    {sheetCaption(s, i + 1, locale)}
                                  </option>
                                ))}
                              </select>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </SectionBlock>
          )}

          <div className="btn-row" style={{ marginTop: "var(--space-4)" }}>
            <button
              type="button"
              className="btn btn-primary"
              disabled={okPaths.length === 0}
              onClick={() => void onRun()}
            >
              {t(locale, "merge")}
            </button>
          </div>
        </>
      )}

      {phase === "running" && (
        <SectionBlock icon={<IconPlay />} label={t(locale, "merging")}>
        <div className="panel" aria-live="polite">
          <div className="progress-track">
            <div className="progress-bar" style={{ width: `${progress?.percent ?? 0}%` }} />
          </div>
          <p className="footer-note">
            {progress?.label
              ? `${progress.label} · ${progress.percent}%`
              : `${progress?.percent ?? 0}%`}
          </p>
          <button type="button" className="btn btn-danger" onClick={() => void onCancel()}>
            {t(locale, "cancel")}
          </button>
        </div>
        </SectionBlock>
      )}

      {phase === "result" && result && (
        <SectionBlock icon={<IconResult />} label={t(locale, "mergeDone")}>
        <div className="panel">
          <h2>
            {result.success
              ? result.partial
                ? t(locale, "partialDone")
                : t(locale, "mergeDone")
              : t(locale, "errorTitle")}
          </h2>
          <p>{result.message}</p>
          <p>
            {t(locale, "fileCount", { n: result.filesUsed })} · {t(locale, "rowsMerged")}:{" "}
            <strong>{result.rows}</strong> · {t(locale, "timeUsed")}:{" "}
            <strong>{((result.durationMs || 0) / 1000).toFixed(2)}</strong>s
          </p>
          {result.outputPath && (
            <dl className="kv">
              <dt>{t(locale, "outputPath")}</dt>
              <dd>{result.outputPath}</dd>
            </dl>
          )}
          {result.items?.length > 0 && (
            <ul className="item-list">
              {result.items.map((it, i) => (
                <li key={i}>
                  <span
                    className={`chip chip-${
                      it.status === "success"
                        ? "ok"
                        : it.status === "skipped"
                          ? "skipped"
                          : "failed"
                    }`}
                  >
                    {it.status}
                  </span>{" "}
                  {it.file}
                  {it.rows ? ` · ${it.rows} ${t(locale, "rows")}` : ""}
                  {it.message ? <span className="reason">{it.message}</span> : null}
                </li>
              ))}
            </ul>
          )}
          {(result.sheets?.length ?? 0) > 0 && (
            <>
              <h3 style={{ marginTop: "var(--space-4)" }}>{t(locale, "usedSheet")}</h3>
              <table className="list-table">
                <thead>
                  <tr>
                    <th>{t(locale, "fileListTitle")}</th>
                    <th style={{ width: 240 }}>{t(locale, "usedSheet")}</th>
                  </tr>
                </thead>
                <tbody>
                  {result.sheets.map((s) => (
                    <tr key={s.path}>
                      <td className="col-name">{s.name}</td>
                      <td>{resultSheetCell(s.path)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}
          <div className="btn-row" style={{ marginTop: "var(--space-4)" }}>
            {result.outputPath && (
              <button
                type="button"
                className="btn btn-primary"
                onClick={() => void RevealInExplorer(result.outputPath)}
              >
                {t(locale, "reveal")}
              </button>
            )}
            <button
              type="button"
              className="btn"
              onClick={() => {
                setResult(null);
                setPhase("input");
              }}
            >
              {t(locale, "mergeAnother")}
            </button>
            <button
              type="button"
              className="btn btn-quiet"
              onClick={() => {
                setFiles([]);
                setPins({});
                setResult(null);
                setPhase("input");
                setError(null);
              }}
            >
              {t(locale, "clearAll")}
            </button>
          </div>
        </div>
        </SectionBlock>
      )}
    </div>
  );
}
