# DESIGN_TOKENS

Convention-mode internal tool. Apple-inspired spacing and calm neutrals; no gradient CTAs, no decorative dashboard chrome.

## Color — Light

| Token | Value | Use |
|-------|-------|-----|
| `--bg` | `#F5F5F7` | window background |
| `--bg-elevated` | `#FFFFFF` | panels, inputs |
| `--text` | `#1D1D1F` | primary text |
| `--text-secondary` | `#6E6E73` | secondary |
| `--text-tertiary` | `#AEAEB2` | placeholder / disabled |
| `--border` | `#D2D2D7` | control borders |
| `--separator` | `#E5E5EA` | list separators |
| `--accent` | `#0071E3` | primary action |
| `--accent-hover` | `#0077ED` | hover |
| `--danger` | `#D70015` | errors |
| `--success` | `#1D7D3F` | success |
| `--warning` | `#B25000` | warnings / degradation |
| `--focus-ring` | `rgba(0,113,227,0.35)` | keyboard focus |

## Color — Dark

| Token | Value |
|-------|-------|
| `--bg` | `#1C1C1E` |
| `--bg-elevated` | `#2C2C2E` |
| `--text` | `#F5F5F7` |
| `--text-secondary` | `#A1A1A6` |
| `--text-tertiary` | `#636366` |
| `--border` | `#3A3A3C` |
| `--separator` | `#38383A` |
| `--accent` | `#0A84FF` |
| `--danger` | `#FF453A` |
| `--success` | `#30D158` |
| `--warning` | `#FF9F0A` |

## Type

| Token | Value |
|-------|-------|
| `--font-ui` | `"PingFang TC", "Microsoft YaHei", "Segoe UI", system-ui, sans-serif` |
| `--font-mono` | `"SF Mono", "Cascadia Code", Consolas, monospace` |
| title | 22px / 600 |
| section | 15px / 600 |
| body | 13px / 400 |
| caption | 11px / 400 |
| button | 13px / 500 |

## Space / shape

| Token | Value |
|-------|-------|
| space-1..6 | 4 / 8 / 12 / 16 / 20 / 24 px |
| radius-sm/md/lg | 6 / 8 / 12 px |
| control-height | 32 px |
| content-max | 720 px |

## Component states (required)

Default · Hover · Focus-visible · Active · Selected · Disabled · Loading · Error · Empty

Drop zone: idle / drag-over / disabled / error  
Progress: percent + cancel + current label  
Result: success / partial / failed / cancelled

## Layout rules

- Single column, max-width 720px, generous horizontal padding
- Tool list on home is two clear rows/cards with one-line purpose — not a marketing hero
- One primary button per view
- Tables/lists over nested cards for file/sheet lists
- Light and dark both first-class; follow system unless user overrides

## i18n

- MVP strings: zh-TW, zh-CN, en
- Keys reserved: ja, ko
- No hardcoded UI Chinese in components; all copy via i18n keys
- Excel user data is never translated
