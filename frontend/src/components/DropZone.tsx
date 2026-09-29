import { useCallback, useRef, useState } from "react";

type Props = {
  title: string;
  hint: string;
  disabled?: boolean;
  multiple?: boolean;
  /** Receives filesystem paths. Only absolute paths are forwarded. */
  onPaths: (paths: string[]) => void;
  onPickClick?: () => void;
  children?: React.ReactNode;
};

/** True if path looks like a Windows drive, UNC, or POSIX absolute path. */
function isAbsolutePath(p: string): boolean {
  if (!p) return false;
  if (/^[a-zA-Z]:[\\/]/.test(p)) return true;
  if (p.startsWith("\\\\")) return true;
  if (p.startsWith("/")) return true;
  return false;
}

/**
 * Drop zone UI.
 * Absolute paths come from Wails EnableFileDrop + runtime OnFileDrop
 * (wired by pages). HTML5 DataTransfer often lacks a real path in WebView2 —
 * bare filenames must not be sent to the backend.
 */
export function DropZone({
  title,
  hint,
  disabled,
  multiple,
  onPaths,
  onPickClick,
  children,
}: Props) {
  const [dragOver, setDragOver] = useState(false);
  const depth = useRef(0);

  const handleFiles = useCallback(
    (files: FileList | null) => {
      if (!files || files.length === 0) return;
      const paths: string[] = [];
      for (let i = 0; i < files.length; i++) {
        const f = files[i] as File & { path?: string };
        const p = f.path;
        // WebView2 HTML5 drop often has no path — skip rather than send "file.xlsx"
        if (p && isAbsolutePath(p)) {
          paths.push(p);
        }
      }
      if (paths.length === 0) return;
      if (!multiple && paths.length > 1) {
        onPaths([paths[0]]);
        return;
      }
      onPaths(paths);
    },
    [multiple, onPaths],
  );

  const onDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    if (disabled) return;
    depth.current += 1;
    setDragOver(true);
  };
  const onDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    if (disabled) return;
  };
  const onDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    depth.current -= 1;
    if (depth.current <= 0) {
      depth.current = 0;
      setDragOver(false);
    }
  };
  const onDrop = (e: React.DragEvent) => {
    e.preventDefault();
    depth.current = 0;
    setDragOver(false);
    if (disabled) return;
    handleFiles(e.dataTransfer?.files ?? null);
  };

  return (
    <div
      className={`dropzone${dragOver ? " drag-over" : ""}${disabled ? " disabled" : ""}`}
      role="button"
      tabIndex={disabled ? -1 : 0}
      aria-disabled={disabled || undefined}
      onClick={() => {
        if (!disabled) onPickClick?.();
      }}
      onKeyDown={(e) => {
        if (disabled) return;
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onPickClick?.();
        }
      }}
      onDragEnter={onDragEnter}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
    >
      <div className="dropzone-title">{title}</div>
      <div className="dropzone-hint">{hint}</div>
      {children}
    </div>
  );
}
