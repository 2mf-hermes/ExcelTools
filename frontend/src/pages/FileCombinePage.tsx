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
import type { model } from "../../wailsjs/go/models";
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

type Phase = "input" | "running" | "result";
type ColumnAlign = "by-header" | "by-position" | "abort";
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

export function FileCombinePage({ locale }: Props) {
  const [phase, setPhase] = useState<Phase>("input");
  const [files, setFiles] = useState<model.FileRef[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [folderPath, setFolderPath] = useState("");

  const [headerRows, setHeaderRows] = useState(1);
  const [sheetIndex, setSheetIndex] = useState(1);
  const [columnAlign, setColumnAlign] = useState<ColumnAlign>("by-header");
  const [addSourceFile, setAddSourceFile] = useState(true);
  const [keepBaseOtherSheets, setKeepBaseOtherSheets] = useState(true);
  const [keepBlankRows, setKeepBlankRows] = useState(true);

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
  };

  const okPaths = useMemo(
    () => files.filter((f) => f.status === "ok").map((f) => f.path),
    [files],
  );
  const hasXls = files.some((f) => f.status === "ok" && f.ext === ".xls");

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
      const res = await MergeFiles({
        paths: okPaths,
        headerRows,
        sheetIndex,
        columnAlign,
        addSourceFile,
        keepBaseOtherSheets,
        keepBlankRows,
        outputDir: "",
      });
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
  }, [okPaths, headerRows, sheetIndex, columnAlign, addSourceFile, keepBaseOtherSheets, keepBlankRows, locale]);

  const onCancel = useCallback(async () => {
    try {
      await CancelFileMerge();
    } catch {
      /* ignore */
    }
  }, []);

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
              onClear={() => setFiles([])}
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
            <label className="field-row">
              <span>{t(locale, "sheetIndex")}</span>
              <input
                className="select-control"
                type="number"
                min={1}
                max={255}
                value={sheetIndex}
                onChange={(e) => setSheetIndex(Math.max(1, Number(e.target.value) || 1))}
              />
              <span className="reason">{t(locale, "sheetIndexDesc")}</span>
            </label>
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
