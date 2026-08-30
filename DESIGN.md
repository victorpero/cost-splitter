---
name: Cost Splitter
description: A calm household reconciliation worksheet for inspectable grocery-cost splitting.
colors:
  canvas: "#f3f1eb"
  panel: "#fffefa"
  ink: "#17201d"
  muted: "#68736e"
  line: "#d8ddd8"
  accent: "#087f6a"
  accent-dark: "#056554"
  accent-pale: "#e1f2ed"
  danger: "#9c2f27"
  danger-bg: "#fff0ed"
typography:
  display:
    fontFamily: "Manrope Variable, ui-sans-serif, sans-serif"
    fontSize: "clamp(25px, 2vw, 34px)"
    fontWeight: 700
    lineHeight: 1.45
    letterSpacing: "-0.035em"
  headline:
    fontFamily: "Manrope Variable, ui-sans-serif, sans-serif"
    fontSize: "17px"
    fontWeight: 700
    lineHeight: 1.45
    letterSpacing: "-0.015em"
  title:
    fontFamily: "Manrope Variable, ui-sans-serif, sans-serif"
    fontSize: "15px"
    fontWeight: 700
    lineHeight: 1.45
  body:
    fontFamily: "Manrope Variable, ui-sans-serif, sans-serif"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.45
  label:
    fontFamily: "Manrope Variable, ui-sans-serif, sans-serif"
    fontSize: "10px"
    fontWeight: 800
    lineHeight: 1.45
    letterSpacing: "0.07em"
  data-entry:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.55
rounded:
  file-control: "5px"
  control: "7px"
  ledger-row: "8px"
  alert: "9px"
  panel: "12px"
  pill: "999px"
spacing:
  dense: "8px"
  compact: "12px"
  standard: "16px"
  section: "20px"
  roomy: "24px"
  empty: "40px"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "#ffffff"
    typography: "{typography.body}"
    rounded: "{rounded.control}"
    padding: "11px 14px"
  button-primary-hover:
    backgroundColor: "{colors.accent-dark}"
    textColor: "#ffffff"
    rounded: "{rounded.control}"
    padding: "11px 14px"
  button-secondary:
    backgroundColor: "transparent"
    textColor: "{colors.accent-dark}"
    typography: "{typography.body}"
    rounded: "{rounded.control}"
    padding: "11px 14px"
  button-text:
    backgroundColor: "transparent"
    textColor: "{colors.accent-dark}"
    rounded: "{rounded.control}"
    padding: "8px 10px"
    height: "40px"
  field:
    backgroundColor: "#ffffff"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.control}"
    padding: "9px 10px"
  panel:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.panel}"
  error-alert:
    backgroundColor: "{colors.danger-bg}"
    textColor: "{colors.danger}"
    typography: "{typography.body}"
    rounded: "{rounded.alert}"
    padding: "12px 15px"
  status-pill:
    backgroundColor: "transparent"
    textColor: "#486158"
    rounded: "{rounded.pill}"
    padding: "7px 11px"
---

# Design System: Cost Splitter

<!-- impeccable:direction-seed 25bb4377 -->

## Overview

**Creative North Star: "The Household Ledger"**

Cost Splitter is a calm household reconciliation worksheet: a practical paper-and-ink surface where imported charges become a compact, inspectable ledger. Warm paper, near-black ink, hairline rules, and one restrained grocery-green action color make the arithmetic feel trustworthy. The interface is made for finishing a real household task, not for resembling a generic finance dashboard.

The direction was resolved from concept seed `25bb4377` as the sixth grounded candidate. Its system-level comparator is the concept roll's deployable-sheet board: a finish bar for linked structure and responsive transformation, not a source of aerospace styling. The shipped system keeps imports, transactions, allocations, and totals visibly connected while remaining quiet enough for repeated use.

The experience tells a working story: set the import terms, reveal every piece of evidence, resolve exceptions, and close the ledger with exact totals. Progress stays visible without hiding the charges that produced it.

**Key Characteristics:**

- Warm, flat worksheet surfaces with precise hairline separation.
- Dense but readable transaction evidence and tabular monetary figures.
- Restrained green reserved for actions, focus, selection, and inclusion.
- Desktop accounting tables that become complete transaction cards on mobile.
- Explicit loading, error, empty, included, unmatched, and removed states.

## Colors

The palette behaves like paper, ink, and one grocery-green annotation pen: neutrals carry nearly all structure while color clarifies action and failure.

### Primary

- **Grocery Green:** The sole positive action and confirmed-inclusion color; use it for primary buttons, checkboxes, carets, focus borders, and reversible row actions.
- **Deep Grocery Green:** The higher-contrast green for hover states, secondary action text, and the file-picker control.
- **Washed Grocery Green:** The quiet interaction wash behind file-picker and secondary/text-button hover states.

### Neutral

- **Warm Canvas:** The broad page field, softened by the implemented pale-green radial wash near the upper left.
- **Ledger Paper:** The panel and ledger surface, held slightly warmer and lighter than the canvas.
- **Accounting Ink:** Primary headings, values, transaction names, and neutral control text.
- **Provenance Gray:** Supporting explanations, source-file metadata, counts, and secondary labels.
- **Hairline Gray:** Dividers, table boundaries, and transaction-card edges.

### Semantic

- **Correction Red:** Reserved for concise, recoverable error messages.
- **Correction Wash:** The error alert's quiet paper-red background.

### Named Rules

**The One Annotation Pen Rule.** Green is the only routine accent; it signals available action, focus, selection, or confirmed inclusion and never decorates neutral data.

**The Evidence Stays Ink Rule.** Totals and transaction amounts remain neutral ink, including negative values. Meaning comes from labels, signs, and position rather than financial red/green coding.

## Typography

**Display Font:** Manrope Variable with a system sans-serif fallback.

**Body Font:** Manrope Variable with a system sans-serif fallback.
**Data-entry Font:** The platform monospace stack is used only for the merchant-prefix editor.

**Character:** A single variable sans-serif keeps the tool compact and contemporary without separating “brand” typography from working typography. Weight, case, and tabular numerals create the hierarchy; decoration does not.

### Hierarchy

- **Display:** The product name only; fluid, tightly tracked, and confidently heavy.
- **Headline:** Panel titles and the main empty-state title.
- **Title:** Transaction-section headings.
- **Body:** Controls, transaction content, descriptions, and helper copy.
- **Label:** Uppercase metric and table labels with deliberate tracking.
- **Data entry:** Merchant prefixes only, where monospaced alignment improves editing.

Currency, dates, imported amounts, split amounts, and participant totals use tabular numerals. Supporting provenance remains visually subordinate but readable.

### Named Rules

**The Arithmetic Holds Still Rule.** Monetary figures use tabular numerals and never animate through intermediate values that could imply a false calculation.

**The One Sans Voice Rule.** Manrope carries headings, labels, controls, records, and totals; monospace is an editing aid, not a second brand voice.

## Layout

The working canvas is centered and capped at `1480px`, with `40px` of total horizontal viewport inset on desktop. The main work area is a two-column grid: a `326px` sticky import panel and a flexible ledger separated by a `24px` gap. At `1120px`, the controls narrow to `285px` and the six summary metrics reflow into two rows of three.

The workflow defines the topology: import files and confirm currency and merchant prefixes; review the matched and unmatched evidence; adjust split amounts and allocations; then read the totals as the consequence of those rows. Summary metrics act as an accounting header above the ledger rather than a detached dashboard.

The first desktop viewport establishes the product identity, one-sentence context, import controls, configuration fields, primary action, summary, and start of the transaction evidence without navigating away. The controls remain sticky while longer ledgers scroll.

At `780px` and below, the page becomes a single column with `12px` side insets. The status pill disappears, the controls return to normal flow, summary metrics form two columns, and each transaction row unfolds into a complete stacked record. The mobile result must show merchant, date, imported amount, editable split amount, allocation, provenance, its included-or-unmatched group context, and the include/remove control without document-level horizontal overflow; the production target is a `390px` viewport with touch targets at least `44px` high.

**The Phase Spacing Rule.** Space expands between import, summary, included, and unmatched phases; rows remain compact so comparisons are fast.

**The Ledger Must Unfold Rule.** Mobile changes the transaction grammar instead of clipping or horizontally scrolling the desktop table.

## Elevation & Depth

The system is flat by default. Depth comes from tonal separation, borders, sticky positioning, and the faint radial canvas wash—not from resting drop shadows. A field receives a tight green halo only while focused, and button hover uses a one-pixel upward movement as transient feedback. Reduced-motion mode removes that movement while preserving the color response.

### Named Rules

**The Flat Ledger Rule.** Panels, tables, metrics, and transaction cards rest on the page without shadows; borders and paper tones carry structure.

**The State-Only Lift Rule.** Elevation appears only as focus or hover feedback and must disappear when the interaction ends.

## Shapes

Shapes are gently practical rather than ornamental. Fields and buttons use compact rounded corners; table frames and mobile transaction cards are slightly softer; the major import and result panels carry the largest radius. The API status is the only pill. Hairline borders define nearly every container, and dashed borders are reserved for an empty transaction group.

**The One Pill Rule.** Pill geometry belongs to the compact API status only; ordinary buttons, fields, panels, and records keep the ledger's modest corner language.

## Components

### Buttons

- **Primary:** A full grocery-green fill, white text, heavy label, and compact control corners. It terminates the import setup sequence and changes from “Analyze CSV Files” to “Recalculate” once records exist.
- **Secondary:** Transparent paper with a thin green-gray border and deep-green text. Hover fills with the pale accent wash.
- **Text action:** A quiet deep-green include/remove action with a `40px` minimum height; hover adds only the pale wash.
- **States:** Hover darkens or washes the background and lifts the control by one pixel; active returns it to rest. Keyboard focus uses a visible three-pixel outline with offset. Disabled actions remain legible, reduce opacity, and use a wait cursor while work is active.

### Inputs and Fields

- **Text, amount, and select fields:** White controls with a neutral one-pixel stroke, compact corners, dark ink, and dense internal padding.
- **Focus:** The border shifts to grocery green and gains a tight translucent green halo; the global keyboard outline remains visible.
- **File input:** Preserves the native file affordance while giving its selector button the pale/deep-green treatment.
- **Checkbox:** Uses the native control with the accent color and a direct text label.
- **Merchant-prefix editor:** Uses the monospace stack, a `150px` minimum height, vertical resize, and compact leading.

### Panels and Containers

- **Panels:** Ledger-paper surfaces with a one-pixel translucent gray-green border, modest large corners, and no resting shadow.
- **Panel headings:** A `20px` inset and a hairline lower rule connect the heading to the content below.
- **Summary strip:** Six equal accounting columns on wide screens, three at intermediate widths, and two on mobile. Each metric uses an uppercase label above a clipped, tabular value.
- **Table frame:** A bordered, softly rounded frame on desktop; the frame disappears on mobile so each row can become its own card.

### Transaction Ledger

Desktop uses a dense table with uppercase micro-labels, tabular dates and amounts, restrained hover wash, inline editing, and the action aligned at the far right. Description and provenance travel together. On mobile, the same semantic row becomes a bordered card: date, imported amount, and action form the first line; description and source occupy the second; split amount and allocation complete the record below.

Included and unmatched groups remain adjacent and reversible. Removing and re-including a transaction restores its original amount and default even split so the visual state continues to describe the calculation state.

### Feedback and Empty States

Errors appear as a bordered correction-red alert above the work area and retain the API's validation meaning in plain language. The result panel reserves at least `470px` for empty and loading states so calculation content does not cause an abrupt page lurch. Successful calculations remain matter-of-fact and do not add celebratory decoration.

## Do's and Don'ts

### Do:

- **Do** keep imported charges inspectable, editable, attributable to their source file, and reversible.
- **Do** preserve exact currency and half-cent display behavior with tabular numerals.
- **Do** keep semantic headings, direct labels, fieldsets where grouping applies, native form affordances, keyboard operation, visible focus, and at least `4.5:1` contrast for normal text.
- **Do** keep loading, validation, empty, included, unmatched, removed, and re-included states explicit.
- **Do** verify that a `390px` viewport has no document-level horizontal overflow and contains at least one complete transaction record.
- **Do** verify production captures without framework chrome; desktop must show the full working hierarchy and dense ledger without clipped labels or values, and the import, edit, remove, and re-include flow must produce no application console warnings or errors.

### Don't:

- **Don't** turn the worksheet into a generic finance dashboard, decorative marketing page, or aerospace interface.
- **Don't** hide the evidence behind an opaque total or detach summary figures from the records that produce them.
- **Don't** use color alone to communicate destructive, included, matched, or unmatched state; actions keep explicit words.
- **Don't** force the desktop table into a clipped or document-scrolling mobile viewport.
- **Don't** add resting shadows, gratuitous pills, celebratory effects, or animated arithmetic.
- **Don't** imply that uploaded files are stored; the implemented flow keeps them in memory for the requested calculation.
