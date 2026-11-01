# Design tokens — "Sunset Journey" (owner decision 2026-10-25)

Brand: **Sabai jai Travel**. Source of truth in code: `apps/mobile-flutter/lib/core/theme.dart` (`TepColors`).

| Token | Hex | Use |
|-------|-----|-----|
| primary | `#C2410C` | App bars, primary buttons, active icons. White text on it = 5.2:1 |
| primaryDeep | `#9A3412` | Gradient start, pressed state |
| secondary | `#FB923C` | Highlights, selected chips, nav indicator. **Use ink text, never white (2.3:1)** |
| background | `#FFF8F3` | Page background |
| surface / card | `#FFFFFF` | Cards, inputs, nav bar |
| accent | `#EAB308` | Rating stars |
| ink (text) | `#431D12` | Body text |
| muted *(derived)* | `#7C5A4B` | Secondary text, ≥ 4.5:1 on background |
| line *(derived)* | `#F1DFD3` | Borders, dividers |
| success / danger *(kept)* | `#23925D` / `#B91C1C` | Status; danger is a deeper red so it is not confused with the orange primary |

Derived/kept tokens are my additions to the five-colour palette the owner gave; confirm or replace.
