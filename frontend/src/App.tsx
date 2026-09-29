import { useCallback, useEffect, useMemo, useState } from "react";
import {
  CheckForUpdates,
  GetAppInfo,
  GetSettings,
  ResetSettings,
  SetSettings,
} from "../wailsjs/go/main/App";
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
      } catch {
        /* keep defaults */
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const persist = useCallback(async (next: { language?: string; theme?: string }) => {
    try {
      const current = await GetSettings();
      const merged: model.Settings = {
        language: next.language ?? current.language ?? "system",
        theme: next.theme ?? current.theme ?? "system",
        defaults: current.defaults ?? {},
      };
      await SetSettings(merged);
      setSettingsMsg("ok");
      window.setTimeout(() => setSettingsMsg(null), 1500);
    } catch {
      setSettingsMsg("fail");
    }
  }, []);

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
    try {
      const res = await CheckForUpdates();
      if (res?.currentVersion) setAppVersion(res.currentVersion);
      setUpdateMsg(
        res?.status === "up-to-date"
          ? t(locale, "updateUpToDate")
          : t(locale, "updateNoSource"),
      );
    } catch {
      setUpdateMsg(t(locale, "updateNoSource"));
    } finally {
      setCheckingUpdate(false);
      window.setTimeout(() => setUpdateMsg(null), 4000);
    }
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
              <div className="btn-row" style={{ marginTop: "var(--space-3)" }}>
                <button
                  type="button"
                  className="btn"
                  disabled={checkingUpdate}
                  onClick={() => void onCheckUpdate()}
                >
                  {checkingUpdate ? "…" : tr("checkUpdates")}
                </button>
              </div>
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
