# Design tokens — TEP "Sunset Journey"

Brand: **TEP — Travel Super App**. Source of truth: `apps/mobile-flutter/lib/core/theme.dart` (`TepColors`).

| Token | Hex | Use |
|---|---:|---|
| Primary | `#C2410C` | App bars, primary buttons, important actions |
| Primary Deep | `#9A3412` | Gradients, pressed states |
| Primary Light | `#FFF0E8` | Brand badge backgrounds and light highlights |
| Secondary | `#FB923C` | Decorative highlights and selected surfaces |
| Coral | `#E85D75` | Lifestyle content and recommendations |
| Highlight | `#EAB308` | Stars and promotions; never as white-text background |
| Text Primary | `#18392B` | Headings and primary content |
| Text Secondary | `#647067` | Supporting text |
| Background | `#F3FBF5` | Page canvas |
| Surface | `#FFFFFF` | Cards, dialogs and inputs |
| Border | `#E2E8E3` | Dividers and card borders |
| Success | `#15803D` | Confirmed and completed states |
| Warning | `#B45309` | Pending and caution states |
| Error | `#B91C1C` | Cancellation and validation errors |
| Info | `#1D4ED8` | Informational messages |

## Accessibility rule

Normal text must achieve at least WCAG AA `4.5:1` contrast against its background. Secondary and highlight colors are decorative unless they are paired with `ink` or another verified dark text color. White text is reserved for dark brand and system surfaces.
