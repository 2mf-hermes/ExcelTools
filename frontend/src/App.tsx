import { useCallback, useEffect, useMemo, useState } from "react";
import {
  CheckForUpdates,
  CheckForUpdatesOnStartup,
  GetAppInfo,
  GetSettings,
  InstallUpdate,
  OpenReleasePage,
  ResetSettings,
  SetSettings,
} from "../wailsjs/go/main/App";
import { EventsOn } from "../wailsjs/runtime/runtime";
import type { model } from "../wailsjs/go/models";
import {
  LOCALES,
  LOCALE_LABELS,
  detectLocale,
  t,
  type Locale,
} from "./i18n";
import {
  IconCloudDown,
  IconFiles,
  IconGlobe,
  IconMoon,
  IconRefresh,
  IconSheets,
} from "./components/ToolIcons";
import { SheetCombinePage } from "./pages/SheetCombinePage";
import { FileCombinePage } from "./pages/FileCombinePage";
import "./App.css";

type ThemeMode = "system" | "light" | "dark";
type Screen = "home" | "settings" | "sheet" | "file";

/**
 * Payload of the "update:progress" event. Declared here because it only ever
 * crosses the bridge as an event, so Wails does not emit it into models.ts.
 * Keep in sync with model.UpdateProgress in core/model/types.go.
 */
type UpdateProgress = { done: number; total: number; pct: number };

function resolveTheme(mode: ThemeMode): "light" | "dark" {
  if (mode !== "system") return mode;
  return window.matchMedia?.("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function App() {
  const [screen, setScreen] = useState<Screen>("home");
  const [locale, setLocale] = useState<Locale>(() => detectLocale(navigator.language));
  const [theme, setTheme] = useState<ThemeMode>("system");
  const [settingsMsg, setSettingsMsg] = useState<string | null>(null);
  const [appVersion, setAppVersion] = useState<string>("");
  const [updateMsg, setUpdateMsg] = useState<string | null>(null);
  const [checkingUpdate, setCheckingUpdate] = useState(false);
  const [update, setUpdate] = useState<model.UpdateCheckResult | null>(null);
  const [installing, setInstalling] = useState(false);
  const [progress, setProgress] = useState<UpdateProgress | null>(null);
  const [autoCheck, setAutoCheck] = useState(true);

  useEffect(() => {
    document.documentElement.dataset.theme = resolveTheme(theme);
  }, [theme]);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => {
      if (theme === "system") {
        document.documentElement.dataset.theme = resolveTheme("system");
      }
    };
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, [theme]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const [appInfo, s] = await Promise.all([GetAppInfo(), GetSettings()]);
        if (cancelled) return;
        setAppVersion(appInfo?.version || "");
        if (s?.language && s.language !== "system") {
          setLocale(s.language as Locale);
        } else {
          setLocale(detectLocale(navigator.language));
        }
        if (s?.theme) setTheme(s.theme as ThemeMode);
        setAutoCheck(s?.autoCheckUpdates ?? true);
      } catch {
        /* keep defaults */
      }
      // Automatic check. It stays silent unless a newer release actually exists:
      // being offline must not greet the user with an error.
      try {
        const res = await CheckForUpdatesOnStartup();
        if (cancelled) return;
        if (res?.status === "available") setUpdate(res);
      } catch {
        /* offline: stay quiet */
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const off = EventsOn("update:progress", (p: UpdateProgress) => {
      setProgress(p);
    });
    return () => {
      if (typeof off === "function") off();
    };
  }, []);

  const persist = useCallback(
    async (next: { language?: string; theme?: string; autoCheckUpdates?: boolean }) => {
      try {
        const current = await GetSettings();
        const merged: model.Settings = {
          language: next.language ?? current.language ?? "system",
          theme: next.theme ?? current.theme ?? "system",
          autoCheckUpdates: next.autoCheckUpdates ?? current.autoCheckUpdates ?? true,
          defaults: current.defaults ?? {},
        };
        await SetSettings(merged);
        setSettingsMsg("ok");
        window.setTimeout(() => setSettingsMsg(null), 1500);
      } catch {
        setSettingsMsg("fail");
      }
    },
    [],
  );

  const onChangeLocale = (l: Locale) => {
    setLocale(l);
    void persist({ language: l });
  };

  const onChangeTheme = (m: ThemeMode) => {
    setTheme(m);
    void persist({ theme: m });
  };

  const onCheckUpdate = async () => {
    setCheckingUpdate(true);
    setUpdateMsg(null);
    setUpdate(null);
    try {
      const res = await CheckForUpdates();
      if (res?.currentVersion) setAppVersion(res.currentVersion);
      setUpdate(res);
      if (res?.status === "up-to-date") {
        setUpdateMsg(t(locale, "updateUpToDate"));
        window.setTimeout(() => setUpdateMsg(null), 4000);
      } else if (res?.status === "error") {
        setUpdateMsg(t(locale, "updateOffline"));
      } else if (res?.status === "available" && !res.canAutoInstall) {
        setUpdateMsg(t(locale, "updateNoDigest"));
      }
    } catch {
      setUpdateMsg(t(locale, "updateOffline"));
    } finally {
      setCheckingUpdate(false);
    }
  };

  const onInstall = async () => {
    setInstalling(true);
    setProgress(null);
    setUpdateMsg(null);
    try {
      const res = await InstallUpdate();
      if (res?.status === "installed") {
        // This process is about to exit in favour of the new build, so the
        // notice stays on screen and `installing` deliberately stays true.
        setUpdateMsg(t(locale, "updateRestarting"));
        return;
      }
      setUpdateMsg(`${t(locale, "updateFailed")}: ${res?.message ?? ""}`);
    } catch (e) {
      setUpdateMsg(`${t(locale, "updateFailed")}: ${String(e)}`);
    }
    setInstalling(false);
    setProgress(null);
  };

  const onToggleAutoCheck = (value: boolean) => {
    setAutoCheck(value);
    void persist({ autoCheckUpdates: value });
  };

  const onReset = async () => {
    try {
      const s = await ResetSettings();
      setLocale(
        s.language === "system" ? detectLocale(navigator.language) : (s.language as Locale),
      );
      setTheme((s.theme as ThemeMode) || "system");
      setSettingsMsg("ok");
      window.setTimeout(() => setSettingsMsg(null), 1500);
    } catch {
      setSettingsMsg("fail");
    }
  };

  const tr = useCallback((k: string) => t(locale, k), [locale]);

  const headerTitle = useMemo(() => {
    if (screen === "sheet") return tr("toolSheetTitle");
    if (screen === "file") return tr("toolFileTitle");
    if (screen === "settings") return tr("settings");
    return tr("appTitle");
  }, [screen, tr]);

  return (
    <div className="app-shell">
      <header className="app-header">
        <h1 className="app-title">{headerTitle}</h1>
        <div className="app-header-actions">
          {screen !== "home" && (
            <button type="button" className="btn btn-quiet" onClick={() => setScreen("home")}>
              {tr("home")}
            </button>
          )}
          {screen !== "settings" && (
            <button type="button" className="btn btn-quiet" onClick={() => setScreen("settings")}>
              {tr("settings")}
            </button>
          )}
        </div>
      </header>

      <main className="app-main">
        {screen === "home" && (
          <>
            <h2 className="page-title">{tr("homeTitle")}</h2>
            <p className="page-lead">{tr("homeLead")}</p>
            <div className="privacy-note">{tr("privacy")}</div>

            <p className="section-label">{tr("toolsSection")}</p>
            <ul className="tool-list">
              <li>
                <button type="button" className="tool-card" onClick={() => setScreen("sheet")}>
                  <span className="tool-icon tool-icon-sheets" aria-hidden="true">
                    <IconSheets />
                  </span>
                  <span className="tool-body">
                    <h2>{tr("toolSheetTitle")}</h2>
                    <p>{tr("toolSheetDesc")}</p>
                  </span>
                  <span className="badge">{tr("badgeMergeReady")}</span>
                  <span className="tool-chevron" aria-hidden="true">
                    ›
                  </span>
                </button>
              </li>
              <li>
                <button type="button" className="tool-card" onClick={() => setScreen("file")}>
                  <span className="tool-icon tool-icon-files" aria-hidden="true">
                    <IconFiles />
                  </span>
                  <span className="tool-body">
                    <h2>{tr("toolFileTitle")}</h2>
                    <p>{tr("toolFileDesc")}</p>
                  </span>
                  <span className="badge">{tr("badgeMergeReady")}</span>
                  <span className="tool-chevron" aria-hidden="true">
                    ›
                  </span>
                </button>
              </li>
            </ul>
          </>
        )}

        {screen === "sheet" && <SheetCombinePage locale={locale} />}
        {screen === "file" && <FileCombinePage locale={locale} />}

        {screen === "settings" && (
          <>
            <h2 className="page-title">{tr("settings")}</h2>

            <div className="panel">
              <div className="panel-title-row">
                <span className="section-block-icon" aria-hidden="true">
                  <IconGlobe />
                </span>
                <h2 style={{ margin: 0 }}>{tr("language")}</h2>
              </div>
              <div className="lang-list" role="listbox" aria-label={tr("language")}>
                {LOCALES.map((l) => (
                  <button
                    key={l}
                    type="button"
                    className="lang-option"
                    role="option"
                    aria-selected={locale === l}
                    onClick={() => onChangeLocale(l)}
                  >
                    <span>{LOCALE_LABELS[l]}</span>
                    {locale === l && (
                      <span className="check" aria-hidden="true">
                        ✓
                      </span>
                    )}
                  </button>
                ))}
              </div>
            </div>

            <div className="panel">
              <div className="panel-title-row">
                <span className="section-block-icon" aria-hidden="true">
                  <IconMoon />
                </span>
                <h2 style={{ margin: 0 }}>{tr("theme")}</h2>
              </div>
              <div className="segmented" role="group" aria-label={tr("theme")}>
                {(["system", "light", "dark"] as ThemeMode[]).map((m) => (
                  <button
                    key={m}
                    type="button"
                    aria-pressed={theme === m}
                    onClick={() => onChangeTheme(m)}
                  >
                    {m === "system"
                      ? tr("themeSystem")
                      : m === "light"
                        ? tr("themeLight")
                        : tr("themeDark")}
                  </button>
                ))}
              </div>
            </div>

            <div className="panel">
              <div className="panel-title-row">
                <span className="section-block-icon" aria-hidden="true">
                  <IconRefresh />
                </span>
                <h2 style={{ margin: 0 }}>{tr("resetSettings")}</h2>
              </div>
              <button type="button" className="btn" onClick={() => void onReset()}>
                {tr("resetSettings")}
              </button>
              {settingsMsg && (
                <p className="footer-note" role="status">
                  {settingsMsg === "ok" ? tr("settingsSaved") : tr("errorGeneric")}
                </p>
              )}
            </div>

            <div className="panel">
              <div className="panel-title-row">
                <span className="section-block-icon" aria-hidden="true">
                  <IconCloudDown />
                </span>
                <h2 style={{ margin: 0 }}>{tr("currentVersion")}</h2>
              </div>
              <dl className="kv">
                <dt>{tr("version")}</dt>
                <dd>{appVersion || "—"}</dd>
              </dl>

              {update?.status === "available" && (
                <div className="update-available">
                  <p className="update-headline">
                    {t(locale, "updateAvailable", { version: update.latestVersion })}
                  </p>

                  {update.releaseNotes && (
                    <details className="update-notes">
                      <summary>{tr("updateNotes")}</summary>
                      <pre>{update.releaseNotes}</pre>
                    </details>
                  )}

                  {installing && (
                    <div
                      className="progress-track"
                      role="progressbar"
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-valuenow={progress?.pct ?? 0}
                    >
                      <div
                        className="progress-bar"
                        style={{ width: `${progress?.pct ?? 0}%` }}
                      />
                    </div>
                  )}
                  {installing && (
                    <p className="footer-note" role="status">
                      {progress && progress.pct > 0
                        ? tr("updateInstalling")
                        : tr("updateDownloading")}
                    </p>
                  )}
                </div>
              )}

              <div className="btn-row" style={{ marginTop: "var(--space-3)" }}>
                <button
                  type="button"
                  className="btn"
                  disabled={checkingUpdate || installing}
                  onClick={() => void onCheckUpdate()}
                >
                  {checkingUpdate ? "…" : tr("checkUpdates")}
                </button>

                {update?.status === "available" && update.canAutoInstall && (
                  <button
                    type="button"
                    className="btn btn-primary"
                    disabled={installing}
                    onClick={() => void onInstall()}
                  >
                    {tr("updateDownload")}
                  </button>
                )}

                {update?.status === "available" && !update.canAutoInstall && (
                  <button
                    type="button"
                    className="btn"
                    disabled={installing}
                    onClick={() => void OpenReleasePage(update.releaseUrl || "")}
                  >
                    {tr("updateManual")}
                  </button>
                )}
              </div>

              <label className="field-row checkbox-row">
                <input
                  type="checkbox"
                  checked={autoCheck}
                  disabled={installing}
                  onChange={(e) => onToggleAutoCheck(e.target.checked)}
                />
                <span>
                  {tr("updateAutoCheck")}
                  <br />
                  <small className="footer-note">{tr("updateAutoCheckHint")}</small>
                </span>
              </label>

              {updateMsg && (
                <p className="footer-note" role="status">
                  {updateMsg}
                </p>
              )}
            </div>
          </>
        )}
      </main>
    </div>
  );
}

export default App;
