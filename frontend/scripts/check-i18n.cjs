// Fallback i18n parity check without TS loader.
const fs = require("fs");
const path = require("path");
const file = path.join(__dirname, "..", "src", "i18n", "index.ts");
const src = fs.readFileSync(file, "utf8");

function extractDict(name) {
  const re = new RegExp(`const ${name}: Dict = \\{([\\s\\S]*?)\\n\\};`);
  const m = src.match(re);
  if (!m) throw new Error("dict not found: " + name);
  const keys = new Set();
  for (const km of m[1].matchAll(/^\s{2}([A-Za-z0-9_]+):/gm)) {
    keys.add(km[1]);
  }
  return keys;
}

const locales = {
  zhTW: extractDict("zhTW"),
  zhCN: extractDict("zhCN"),
  en: extractDict("en"),
  ja: extractDict("ja"),
  ko: extractDict("ko"),
};
const base = locales.en;
let fail = false;
for (const [name, keys] of Object.entries(locales)) {
  const missing = [...base].filter((k) => !keys.has(k));
  const extra = [...keys].filter((k) => !base.has(k));
  if (missing.length || extra.length) {
    fail = true;
    console.error(`${name}: missing=[${missing}] extra=[${extra}]`);
  }
}
if (fail) process.exit(1);
console.log(`i18n OK: 5 locales x ${base.size} keys`);
