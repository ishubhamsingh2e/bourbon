# AGENTS.md - Engineering & Coding Guidelines

This document defines the strict engineering standards, coding principles, and behavioral rules for AI coding assistants working in this repository.

---

## 1. No Placeholder or Dummy Code
- **Complete Implementations Only**: Never generate stubbed functions, mock returns, `// TODO: implement later`, `// Add logic here`, or partial placeholders.
- **Production-Grade**: All written code must be syntactically valid, type-safe, fully wired, and immediately runnable.
- **No Incomplete Scaffolding**: If a function, method, or component is introduced, write its real, functioning implementation end-to-end.

---

## 2. Strict Commenting Rules
- **No Changelog Comments**: Never write conversational comments describing your actions (e.g., `// Updated this function`, `// Changed for new requirement`, `// Added by AI`).
- **No Diff Commentary**: Do not leave inline diff artifacts, history remarks, or author tags.
- **Documentation Only**: Comments must strictly be used for:
  - Exported Go documentation / Godoc comments (`// Package ...`, `// FunctionName does ...`).
  - Complex domain algorithms, non-obvious business invariants, or security constraints.
  - Structural definitions and interface contracts.
- **Self-Documenting Code**: Prefer clean, readable, and idiomatic naming over explanatory comments.

---

## 3. Skill & Tool Utilization
- **Skill-First Workflows**: Always identify and leverage the appropriate specialized agent skill before designing or modifying code.
- **Adhere to Skill Directives**: When a skill is activated, follow its constraints, checklists, and patterns precisely without skipping validation steps.

---

## 4. Code Quality & Minimalism (Ponytail Principles)
- **Zero Unrequested Abstractions**: Do not introduce single-implementation interfaces, premature factories, or speculative generic layers.
- **Standard Library First**: Prefer native Go standard library packages before adding external dependencies.
- **Root-Cause Fixes**: When resolving bugs, trace all callers and fix the issue at the common root rather than patching individual symptoms.
- **Clean Output**: Avoid unnecessary decorative fluff, stickers, or emoji clutter in source files, logs, and artifacts. Keep code and communication concise, professional, and focused.
