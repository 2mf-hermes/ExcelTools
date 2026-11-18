import { useEffect } from "react";
import type { model } from "../../wailsjs/go/models";
import { ListSheets } from "../../wailsjs/go/main/App";

/** Sheets per file path. */
export type SheetMap = Record<string, model.SheetRef[]>;

/**
 * Reads the sheet list of every path that is not in `known` yet, sequentially
 * (so a hundred-file drop does not spawn a hundred Excel readers at once).
 * `onLoaded` is called after each file, with that file's sheets, or with the
 * error text as the second argument when the workbook could not be read.
 *
 * `known` is only read when the effect starts; it is deliberately not a
 * dependency, otherwise every cache update would re-run the effect.
 */
export function useSheetLists(
  paths: string[],
  known: SheetMap,
  onLoaded: (path: string, sheets: model.SheetRef[] | null, error?: string) => void,
) {
  const knownRef = known;
  const cbRef = onLoaded;

  useEffect(() => {
    let cancelled = false;
    const todo = paths.filter((p) => !knownRef[p]);
    if (!todo.length) return;

    void (async () => {
      for (const path of todo) {
        if (cancelled) return;
        try {
          const res = await ListSheets(path);
          if (cancelled) return;
          cbRef(path, res?.sheets ?? []);
        } catch (e) {
          if (cancelled) return;
          cbRef(path, null, String(e));
        }
      }
    })();

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [paths.join("\n")]);
}
