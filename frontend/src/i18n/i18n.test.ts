/**
 * Localization key parity check. Run: npm run test:i18n
 * Fails if any locale misses a key present in English.
 */
import { I18N_KEYS, LOCALES, missingI18nKeys } from "./index.ts";

const missing = missingI18nKeys();
if (missing.length > 0) {
  console.error("Missing i18n keys:");
  for (const m of missing) {
    console.error(`  ${m.locale}: ${m.keys.join(", ")}`);
  }
  process.exit(1);
}
console.log(`i18n OK: ${LOCALES.length} locales × ${I18N_KEYS.length} keys`);
