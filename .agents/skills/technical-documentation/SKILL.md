---
name: technical-documentation
description: Skill for writing high-quality, professional technical documentation following Write the Docs (WTD) standards, Docs-as-Code approach, and accessibility guidelines.
---

# Technical Documentation Guidelines & Standards

This skill provides comprehensive, practitioner-grade guidelines and methodologies for writing, structuring, and maintaining software documentation based on the collective wisdom of the **Write the Docs (WTD)** community.

---

## 1. General Documentation Principles

### Precursory (Documentation-Driven Design)

- **Write Docs Before Code**: Document requirements, APIs, and configuration schemas _before_ beginning implementation. This serves as the first draft and guides design decisions.
- **Clarify Intent Early**: Front-loading documentation exposes edge cases, dependency bottlenecks, and integration issues early.

### Participatory (Democratic Ownership)

- **Democratic Contributions**: Integrate documentation writing directly into the standard engineering workflow. Avoid isolating documentation work to specialized writers.
- **Empower User Feedback**: Provide readers with clear pathways to comment, flag out-of-date content, or suggest edits to maintain documentation health.

---

## 2. Principles for Great Content

### ARID (Accept Repetition In Documentation)

- **Balance DRY vs. WET**: While code should be DRY (Don't Repeat Yourself), documentation inevitably requires some repetition ("moisture") to explain the code's logic.
- **Minimize Redundant Maintenance**: Try to keep details as single-source as possible, but prioritize reader completion and clarity over rigid code-like DRY styling.

### Skimmability & Newspaper Style

- **Newspaper Layout**: Write for readers who skim. Put the most important information first.
- **Front-Loaded Paragraphs**: Begin paragraphs and list items with clear, identifiable concepts.
- **Descriptive Hyperlinks**: Link text must name the specific destination (e.g., _"refer to the [RabbitMQ Routing Guide](url)"_). Never use generic phrases like _"click here"_ or _"this link"_.
- **Descriptive Headers**: Avoid generic headers (e.g., use _"Configuring Redis Connections"_ instead of _"Getting Started"_).

### Accuracy, Currency, & Examples

- **Incorrect Docs are Worse than Missing Docs**: Out-of-date or incorrect documentation degrades developer trust. Maintain accuracy strictly.
- **Version-Agnostic Patterns**: Avoid hardcoding specific patch versions in tutorial command snippets (e.g., generalize download targets using variable placeholders).
- **Common-Path Examples**: Include code samples for the 80% use cases, and clearly separate these examples from dense API reference details.

---

## 3. Principles for Content Sources & Publications

### nearby & Unique Sources

- **Keep Sources Close**: Store markdown/MDX files within the same repository as the source code they document to merge development and writing workflows.
- **Disjoint Scopes**: Define clear, non-overlapping boundaries for different documentation pages to prevent parallel out-of-sync edits.

### Publication Standards

- **Discoverable**: Ensure entry points are highly visible. Document where sidebars, API indexes, and guides live.
- **Typography & Readability**: Ensure documentation features premium styling: clean typography, comfortable line-heights, high contrast, and responsive viewports.
- **Integrated Search**: All documentation portals must feature functional full-text search systems.
- **Versioned Docs**: Support separate docs spaces for active releases/versions.

---

## 4. Docs-as-Code & DocOps best practices

Treat documentation with the same engineering rigor as code:

- **Version Control**: Manage all documentation files in git alongside the source.
- **CI Checks**: Set up automated tests during builds to:
  - Find and flag broken URLs and file paths.
  - Validate YAML configuration schema structures.
  - Verify markdown format linting constraints.
- **Peer Review**: Code reviews for feature PRs should enforce matching documentation updates.

---

## 5. Accessibility (a11y) & Bias Reduction

- **Heading Hierarchy**: Maintain strict heading progression (`# h1` -> `## h2` -> `### h3`). Never skip heading levels for size styling.
- **Descriptive Alt Text**: All media, diagrams, and figures must have meaningful alt text descriptions. Avoid naming paths or using file names.
- **Spatial Cautions**: Avoid visual directional references (e.g., do not say _"see menu on the right"_ or _"click the blue button below"_ as they break for screen readers and mobile views).
- **Inclusive Terminology**: Use inclusive, neutral naming conventions in configurations and code examples (e.g., use `alice@example.com` or standard placeholder domains like `example.com`).
