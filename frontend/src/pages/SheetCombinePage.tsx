import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ClassifyPaths,
  ListSheets,
  MergeSheets,
  CancelSheetMerge,
  PickExcelFile,
  RevealInExplorer,
} from "../../wailsjs/go/main/App";
import type { model } from "../../wailsjs/go/models";
import { DropZone } from "../components/DropZone";
import { SheetList } from "../components/SheetList";
import { SectionBlock, ToolHero } from "../components/ToolLayout";
import { IconChecklist, IconPlay, IconResult, IconSheets, IconSliders } from "../components/ToolIcons";
import { t, type Locale } from "../i18n";
import { EventsOn, OnFileDrop, OnFileDropOff } from "../../wailsjs/runtime/runtime";

type Phase = "input" | "options" | "running" | "result";
type HeaderMode = "each" | "first" | "none";
type ProgressEvent = {
  phase: string;
  current: number;
  total: number;
  label: string;
  percent: number;
};

type Props = { locale: Locale };

export function SheetCombinePage({ locale }: Props) {
  const [phase, setPhase] = useState<Phase>("input");
  const [file, setFile] = useState<model.FileRef | null>(null);
  const [sheets, setSheets] = useState<model.SheetRef[]>([]);
  const [order, setOrder] = useState<string[]>([]);
  const [selected, setSelected] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [note, setNote] = useState<string | null>(null);

  const [headerMode, setHeaderMode] = useState<HeaderMode>("each");
  const [addSourceSheet, setAddSourceSheet] = useState(true);
  const [skipEmpty, setSkipEmpty] = useState(true);

  const [progress, setProgress] = useState<ProgressEvent | null>(null);
  const [result, setResult] = useState<model.MergeResult | null>(null);
  const [running, setRunning] = useState(false);

  const loadSheets = useCallback(
    async (path: string) => {
      setLoading(true);
      setError(null);
      setNote(null);
      setResult(null);
      setPhase("input");
      setSheets([]);
      setOrder([]);
      setSelected({});
      try {
        const refs = await ClassifyPaths([path]);
        const ref = refs[0];
        if (!ref) {
          setError(t(locale, "errorGeneric"));
          setFile(null);
          return;
        }
        setFile(ref);
        if (ref.status !== "ok") {
          setError(ref.reason || t(locale, "errorFormat"));
          return;
        }
        const listResult = await ListSheets(ref.path);
        const list = listResult?.sheets || [];
        setSheets(list);
        setOrder(list.map((s) => s.name));
        const sel: Record<string, boolean> = {};
        for (const s of list) {
          sel[s.name] = !s.empty;
        }
        setSelected(sel);
        if (listResult?.note) setNote(listResult.note);
        if (ref.ext === ".xls") setNote(t(locale, "xlsNote"));
        setPhase("options");
      } catch (e) {
        setError(String(e));
        setFile(null);
      } finally {
        setLoading(false);
      }
    },
    [locale],
  );

  // Wails native file drop (absolute paths; EnableFileDrop in main.go)
  const phaseRef = useRef(phase);
  phaseRef.current = phase;
  const loadSheetsRef = useRef(loadSheets);
  loadSheetsRef.current = loadSheets;

  useEffect(() => {
    try {
      OnFileDrop((_x, _y, paths) => {
        if (!paths || paths.length === 0) return;
        const p = paths.find(
          (x) => /^[a-zA-Z]:[\\/]/.test(x) || x.startsWith("\\\\") || x.startsWith("/"),
        );
        if (!p) return;
        if (phaseRef.current === "input" || phaseRef.current === "options") {
          void loadSheetsRef.current(p);
        }
      }, true);
    } catch {
      // browser dev without wails runtime
    }
    return () => {
      try {
        OnFileDropOff();
      } catch {
        /* ignore */
      }
    };
  }, []);

  useEffect(() => {
    try {
      EventsOn("merge:progress", (data) => {
        const ev = data as ProgressEvent;
        if (ev) setProgress(ev);
      });
    } catch {
      /* browser without wails */
    }
  }, []);

  const onPaths = useCallback(
    (paths: string[]) => {
      const p = paths?.find(
        (x) => /^[a-zA-Z]:[\\/]/.test(x) || x.startsWith("\\\\") || x.startsWith("/"),
      );
      if (p) void loadSheets(p);
    },
    [loadSheets],
  );

  const onPick = useCallback(async () => {
    try {
      const ref = await PickExcelFile();
      if (ref?.path) await loadSheets(ref.path);
    } catch (e) {
      setError(String(e));
    }
  }, [loadSheets]);

  const onToggle = (name: string) => {
    setSelected((s) => ({ ...s, [name]: !s[name] }));
  };
  const onSelectAll = (on: boolean) => {
    const next: Record<string, boolean> = {};
    for (const s of sheets) next[s.name] = on;
    setSelected(next);
  };
  const onMove = (name: string, dir: -1 | 1) => {
    setOrder((ord) => {
      const i = ord.indexOf(name);
      const j = i + dir;
      if (i < 0 || j < 0 || j >= ord.length) return ord;
      const next = ord.slice();
      [next[i], next[j]] = [next[j], next[i]];
      return next;
    });
  };

  const clear = () => {
    setFile(null);
    setSheets([]);
    setOrder([]);
    setSelected({});
    setError(null);
    setNote(null);
    setResult(null);
    setProgress(null);
    setPhase("input");
  };

  const selectedNames = useMemo(
    () => order.filter((n) => selected[n]),
    [order, selected],
  );

  const onRun = useCallback(async () => {
    if (!file || selectedNames.length === 0) {
      setError(t(locale, "errorEmpty"));
      return;
    }
    setRunning(true);
    setError(null);
    setResult(null);
    setProgress({ phase: "start", current: 0, total: 1, label: "", percent: 0 });
    setPhase("running");
    try {
      const res = await MergeSheets({
        sourcePath: file.path,
        sheetNames: selectedNames,
        headerMode,
        addSourceSheet,
        skipEmpty,
        outputSheetName: "",
        sourceColumnLabel: "",
        outputPath: "",
      });
      setResult(res);
      setPhase("result");
    } catch (e) {
      const msg = String(e);
      if (msg.includes("E_CANCEL")) {
        setError(t(locale, "cancelled"));
        setPhase("options");
      } else {
        setError(msg);
        setPhase("options");
      }
    } finally {
      setRunning(false);
      setProgress(null);
    }
  }, [file, selectedNames, headerMode, addSourceSheet, skipEmpty, locale]);

  const onCancel = useCallback(async () => {
    try {
      await CancelSheetMerge();
    } catch {
      /* ignore */
    }
  }, []);

  return (
    <div>
      <ToolHero
        icon={<IconSheets />}
        iconClass="tool-icon-sheets"
        title={t(locale, "toolSheetTitle")}
        lead={t(locale, "toolSheetDesc")}
      />

      {phase === "input" && (
        <SectionBlock icon={<IconChecklist />} label={t(locale, "sheetStepInput")}>
          <DropZone
            title={t(locale, "sheetDropTitle")}
            hint={t(locale, "sheetDropHint")}
            multiple={false}
            onPaths={onPaths}
            onPickClick={() => void onPick()}
          >
            <button
              type="button"
              className="btn"
              onClick={(e) => {
                e.stopPropagation();
                void onPick();
              }}
            >
              {file ? t(locale, "changeFile") : t(locale, "chooseFile")}
            </button>
          </DropZone>
        </SectionBlock>
      )}

      {error && (
        <div className="alert alert-error" role="alert" style={{ marginTop: "var(--space-4)" }}>
          <strong>{t(locale, "errorTitle")}</strong>
          {error}
          <div className="btn-row" style={{ marginTop: "var(--space-2)" }}>
            <button type="button" className="btn" onClick={() => void onPick()}>
              {t(locale, "errorRetry")}
            </button>
          </div>
        </div>
      )}
      {note && (
        <div className="alert alert-warning" style={{ marginTop: "var(--space-4)" }}>
          {note}
        </div>
      )}

      {(phase === "options" || phase === "running") && file && (
        <div>
          <SectionBlock icon={<IconChecklist />} label={t(locale, "sheetStepSheets")}>
          <div className="file-meta">
            <span>
              {t(locale, "selectedFile")}: <strong>{file.name}</strong>
            </span>
            {sheets.length > 0 && (
              <span>
                {sheets.length} {t(locale, "sheetsFound")}
              </span>
            )}
            <button type="button" className="btn btn-quiet" onClick={clear} disabled={running}>
              {t(locale, "clearFile")}
            </button>
          </div>

          {loading ? (
            <div className="panel">
              <p className="loading-row">{t(locale, "loadingSheets")}</p>
            </div>
          ) : (
            <>
              <SheetList
                sheets={sheets}
                selected={selected}
                locale={locale}
                order={order}
                onToggle={onToggle}
                onSelectAll={onSelectAll}
                onMove={onMove}
              />

              <SectionBlock icon={<IconSliders />} label={t(locale, "options")}>
                <div className="panel">
                <label className="field-row">
                  <span>{t(locale, "headerMode")}</span>
                  <select
                    className="select-control"
                    value={headerMode}
                    disabled={running}
                    onChange={(e) => setHeaderMode(e.target.value as HeaderMode)}
                  >
                    <option value="each">{t(locale, "headerEach")}</option>
                    <option value="first">{t(locale, "headerFirst")}</option>
                    <option value="none">{t(locale, "headerNone")}</option>
                  </select>
                </label>
                <label className="field-row checkbox-row">
                  <input
                    type="checkbox"
                    className="sheet-check"
                    checked={addSourceSheet}
                    disabled={running}
                    onChange={(e) => setAddSourceSheet(e.target.checked)}
                  />
                  <span>
                    {t(locale, "sheetNameCol")}
                    <span className="reason">{t(locale, "sheetNameColDesc")}</span>
                  </span>
                </label>
                <label className="field-row checkbox-row">
                  <input
                    type="checkbox"
                    className="sheet-check"
                    checked={skipEmpty}
                    disabled={running}
                    onChange={(e) => setSkipEmpty(e.target.checked)}
                  />
                  <span>
                    {t(locale, "skipEmpty")}
                    <span className="reason">{t(locale, "skipEmptyDesc")}</span>
                  </span>
                </label>
              </div>
              </SectionBlock>
            </>
          )}

          {phase === "running" && (
            <SectionBlock icon={<IconPlay />} label={t(locale, "merging")}>
            <div className="panel" aria-live="polite">
              <div className="progress-track">
                <div
                  className="progress-bar"
                  style={{ width: `${progress?.percent ?? 0}%` }}
                />
              </div>
              <p className="footer-note">
                {progress?.label ? `${progress.label} · ${progress.percent}%` : `${progress?.percent ?? 0}%`}
              </p>
              <button type="button" className="btn btn-danger" onClick={() => void onCancel()}>
                {t(locale, "cancel")}
              </button>
            </div>
            </SectionBlock>
          )}

          {phase === "options" && !loading && (
            <div className="btn-row" style={{ marginTop: "var(--space-4)" }}>
              <button
                type="button"
                className="btn btn-primary"
                disabled={selectedNames.length === 0}
                onClick={() => void onRun()}
              >
                {t(locale, "merge")}
              </button>
              <button type="button" className="btn" onClick={() => setPhase("input")}>
                {t(locale, "back")}
              </button>
            </div>
          )}
          </SectionBlock>
        </div>
      )}

      {phase === "result" && result && (
        <SectionBlock icon={<IconResult />} label={t(locale, "mergeDone")}>
        <div className="panel">
          <h2>{result.success ? t(locale, "mergeDone") : t(locale, "errorTitle")}</h2>
          <p>
            {t(locale, "rowsMerged")}: <strong>{result.rows}</strong> · {t(locale, "cols")}:{" "}
            <strong>{result.cols}</strong> · {t(locale, "timeUsed")}:{" "}
            <strong>{((result.durationMs || 0) / 1000).toFixed(2)}</strong>s
          </p>
          <dl className="kv">
            <dt>{t(locale, "outputPath")}</dt>
            <dd>{result.outputPath}</dd>
          </dl>
          {result.items?.length > 0 && (
            <ul className="item-list">
              {result.items.map((it, i) => (
                <li key={i}>
                  <span className={`chip chip-${it.status === "success" ? "ok" : it.status === "skipped" ? "skipped" : "failed"}`}>
                    {it.status}
                  </span>{" "}
                  {it.file}
                  {it.message ? <span className="reason">{it.message}</span> : null}
                </li>
              ))}
            </ul>
          )}
          <div className="btn-row" style={{ marginTop: "var(--space-4)" }}>
            <button
              type="button"
              className="btn btn-primary"
              onClick={() => void RevealInExplorer(result.outputPath || "")}
            >
              {t(locale, "reveal")}
            </button>
            <button type="button" className="btn" onClick={() => setPhase("options")}>
              {t(locale, "mergeAnother")}
            </button>
            <button type="button" className="btn btn-quiet" onClick={clear}>
              {t(locale, "clearFile")}
            </button>
          </div>
        </div>
        </SectionBlock>
      )}
    </div>
  );
}
