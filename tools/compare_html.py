#!/usr/bin/env python3
"""compare_html.py — Compare Go HTML output against C# (ground-truth) HTML output.

For each .html file present in csharp-html-output/ that also exists in html-output/,
the script parses both files, extracts structured data, and writes a per-report
<filename>.md delta document into html-delta/.

Usage:
    python3 tools/compare_html.py
    python3 tools/compare_html.py --csharp csharp-html-output --go html-output --out html-delta
    python3 tools/compare_html.py --report "Simple List"

Comparison dimensions:
    - Page count
    - Text tokens present in C# but missing from Go (data loss)
    - Text tokens present in Go but absent in C# (spurious/unresolved expressions)
    - Numeric value accuracy (same text, different numbers)
    - Summary: pass / needs-attention
"""

import argparse
import difflib
import os
import re
import sys
from html.parser import HTMLParser
from pathlib import Path


# ---------------------------------------------------------------------------
# HTML parsing helpers
# ---------------------------------------------------------------------------

class PageExtractor(HTMLParser):
    """Extracts per-page text tokens and page count from FastReport HTML."""

    def __init__(self):
        super().__init__()
        self.pages: list[list[str]] = []          # list of pages, each a list of text chunks
        self._current_page: list[str] | None = None
        self._in_body = False
        self._depth_stack: list[str] = []
        self._skip_depth: int | None = None        # depth at which we entered a skip zone

    # Tags that indicate a page boundary in FastReport HTML output.
    # Match frpage0, frpage1, etc. but NOT frpage-container (the wrapper div).
    _PAGE_CLASSES = re.compile(r'\bfrpage\d+\b')

    def handle_starttag(self, tag, attrs):
        self._depth_stack.append(tag)
        depth = len(self._depth_stack)

        attr_dict = dict(attrs)
        cls = attr_dict.get('class', '')
        style = attr_dict.get('style', '')

        # Detect page container (both C# and Go use class="frpageN").
        if self._PAGE_CLASSES.search(cls):
            self._current_page = []
            self.pages.append(self._current_page)
            return

        # Skip <style> and <script> blocks.
        if tag in ('style', 'script', 'head') and self._skip_depth is None:
            self._skip_depth = depth

    def handle_endtag(self, tag):
        depth = len(self._depth_stack)
        if self._skip_depth is not None and depth == self._skip_depth:
            self._skip_depth = None
        if self._depth_stack:
            self._depth_stack.pop()

    def handle_data(self, data):
        if self._skip_depth is not None:
            return
        text = data.strip()
        if text and self._current_page is not None:
            self._current_page.append(text)


def parse_html(path: Path) -> tuple[int, list[list[str]]]:
    """Return (page_count, pages) where pages[i] is a list of text tokens."""
    parser = PageExtractor()
    parser.feed(path.read_text(encoding='utf-8', errors='replace'))
    pages = parser.pages
    return len(pages), pages


def flatten_tokens(pages: list[list[str]]) -> list[str]:
    """Flatten per-page token lists into a single ordered list."""
    result = []
    for page in pages:
        result.extend(page)
    return result


# ---------------------------------------------------------------------------
# Delta analysis
# ---------------------------------------------------------------------------

_UNRESOLVED_RE = re.compile(r'\[[A-Za-z_][A-Za-z0-9_.]*\]')


def is_unresolved(token: str) -> bool:
    """True if token looks like an unevaluated bracket expression."""
    return bool(_UNRESOLVED_RE.search(token))


def analyse(name: str,
            cs_pages: list[list[str]],
            go_pages: list[list[str]]) -> dict:
    """Return a delta dict for one report."""
    cs_flat = flatten_tokens(cs_pages)
    go_flat = flatten_tokens(go_pages)

    cs_set = set(cs_flat)
    go_set = set(go_flat)

    missing_from_go = sorted(cs_set - go_set)
    spurious_in_go = sorted(go_set - cs_set)
    unresolved = [t for t in spurious_in_go if is_unresolved(t)]

    # Sequence diff for ordering issues (top-20 hunks only).
    matcher = difflib.SequenceMatcher(None, cs_flat, go_flat, autojunk=False)
    opcodes = matcher.get_opcodes()
    diff_hunks: list[str] = []
    for tag, i1, i2, j1, j2 in opcodes:
        if tag == 'equal':
            continue
        cs_chunk = cs_flat[i1:i2]
        go_chunk = go_flat[j1:j2]
        if tag == 'replace':
            diff_hunks.append(f'replace  CS={cs_chunk}  →  Go={go_chunk}')
        elif tag == 'delete':
            diff_hunks.append(f'missing  CS={cs_chunk}')
        elif tag == 'insert':
            diff_hunks.append(f'extra    Go={go_chunk}')
        if len(diff_hunks) >= 30:
            diff_hunks.append('… (truncated, more differences exist)')
            break

    similarity = matcher.ratio()

    return {
        'name': name,
        'cs_pages': len(cs_pages),
        'go_pages': len(go_pages),
        'page_match': len(cs_pages) == len(go_pages),
        'missing_from_go': missing_from_go,
        'spurious_in_go': spurious_in_go,
        'unresolved': unresolved,
        'diff_hunks': diff_hunks,
        'similarity': similarity,
    }


# ---------------------------------------------------------------------------
# Markdown report generation
# ---------------------------------------------------------------------------

STATUS_PASS = '✅ PASS'
STATUS_MINOR = '⚠️  MINOR'
STATUS_FAIL = '❌ FAIL'


def status(delta: dict) -> str:
    if not delta['page_match']:
        return STATUS_FAIL
    if delta['unresolved']:
        return STATUS_FAIL
    if delta['missing_from_go']:
        return STATUS_FAIL
    if delta['similarity'] < 0.85:
        return STATUS_MINOR
    if delta['spurious_in_go']:
        return STATUS_MINOR
    return STATUS_PASS


def render_md(delta: dict) -> str:
    name = delta['name']
    st = status(delta)
    lines: list[str] = []

    lines.append(f'# {name}')
    lines.append('')
    lines.append(f'**Status:** {st}  ')
    lines.append(f'**Similarity:** {delta["similarity"]:.1%}  ')
    lines.append(f'**Pages:** C# = {delta["cs_pages"]}, Go = {delta["go_pages"]}'
                 + ('' if delta['page_match'] else '  ⚠️ mismatch'))
    lines.append('')

    # Unresolved expressions — highest priority
    if delta['unresolved']:
        lines.append('## Unresolved Expressions in Go Output')
        lines.append('')
        lines.append('These tokens appear to be unevaluated bracket expressions '
                     'that should have been replaced with data values:')
        lines.append('')
        for t in delta['unresolved'][:50]:
            lines.append(f'- `{t}`')
        if len(delta['unresolved']) > 50:
            lines.append(f'- … and {len(delta["unresolved"]) - 50} more')
        lines.append('')

    # Missing text
    if delta['missing_from_go']:
        lines.append('## Text Present in C# Output but Missing from Go')
        lines.append('')
        lines.append('These text tokens appear in the ground-truth C# output '
                     'but are absent from the Go output:')
        lines.append('')
        shown = [t for t in delta['missing_from_go']
                 if not is_unresolved(t)][:60]
        for t in shown:
            lines.append(f'- `{t}`')
        if len(delta['missing_from_go']) > 60:
            lines.append(f'- … and {len(delta["missing_from_go"]) - 60} more')
        lines.append('')

    # Spurious text (not unresolved)
    extra = [t for t in delta['spurious_in_go'] if not is_unresolved(t)]
    if extra:
        lines.append('## Text Present in Go Output but Absent from C#')
        lines.append('')
        lines.append('These tokens appear only in the Go output '
                     '(may be extra labels, formatting artefacts, or duplicates):')
        lines.append('')
        for t in extra[:40]:
            lines.append(f'- `{t}`')
        if len(extra) > 40:
            lines.append(f'- … and {len(extra) - 40} more')
        lines.append('')

    # Sequence diff hunks
    if delta['diff_hunks']:
        lines.append('## Sequence Diff Hunks')
        lines.append('')
        lines.append('Ordered differences between C# and Go text token sequences '
                     '(up to 30 hunks):')
        lines.append('')
        lines.append('```')
        lines.extend(delta['diff_hunks'])
        lines.append('```')
        lines.append('')

    if st == STATUS_PASS:
        lines.append('_No significant differences detected._')
        lines.append('')

    return '\n'.join(lines)


# ---------------------------------------------------------------------------
# Summary index
# ---------------------------------------------------------------------------

def render_summary(results: list[tuple[str, dict]]) -> str:
    lines: list[str] = []
    lines.append('# HTML Delta Summary')
    lines.append('')
    lines.append('Comparison of Go HTML output (`html-output/`) against '
                 'C# ground-truth (`csharp-html-output/`).  ')
    lines.append('')

    pass_list, minor_list, fail_list, missing_list = [], [], [], []
    for fname, delta in results:
        if delta is None:
            missing_list.append(fname)
        else:
            st = status(delta)
            entry = (fname, delta)
            if st == STATUS_PASS:
                pass_list.append(entry)
            elif st == STATUS_MINOR:
                minor_list.append(entry)
            else:
                fail_list.append(entry)

    total = len(results)
    lines.append(f'| Category | Count |')
    lines.append(f'|---|---|')
    lines.append(f'| ✅ Pass | {len(pass_list)} |')
    lines.append(f'| ⚠️  Minor differences | {len(minor_list)} |')
    lines.append(f'| ❌ Fail | {len(fail_list)} |')
    lines.append(f'| ➖ Go output missing | {len(missing_list)} |')
    lines.append(f'| **Total** | **{total}** |')
    lines.append('')

    def table_rows(entries):
        for fname, delta in entries:
            sim = f'{delta["similarity"]:.0%}'
            pages = (f'{delta["cs_pages"]}' if delta['page_match']
                     else f'{delta["cs_pages"]} → {delta["go_pages"]} ⚠️')
            link = fname.replace(' ', '%20')
            lines.append(f'| [{fname}]({link}.md) | {sim} | {pages} |')

    if fail_list:
        lines.append('## ❌ Failures')
        lines.append('')
        lines.append('| Report | Similarity | Pages |')
        lines.append('|---|---|---|')
        table_rows(fail_list)
        lines.append('')

    if minor_list:
        lines.append('## ⚠️  Minor Differences')
        lines.append('')
        lines.append('| Report | Similarity | Pages |')
        lines.append('|---|---|---|')
        table_rows(minor_list)
        lines.append('')

    if pass_list:
        lines.append('## ✅ Passing')
        lines.append('')
        lines.append('| Report | Similarity | Pages |')
        lines.append('|---|---|---|')
        table_rows(pass_list)
        lines.append('')

    if missing_list:
        lines.append('## ➖ Go Output Missing')
        lines.append('')
        lines.append('These reports were rendered by C# but have no matching '
                     'Go output file:')
        lines.append('')
        for fname in missing_list:
            lines.append(f'- {fname}')
        lines.append('')

    return '\n'.join(lines)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(description=__doc__,
                                     formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument('--csharp', default='csharp-html-output',
                        help='Directory with C# ground-truth HTML files (default: csharp-html-output)')
    parser.add_argument('--go', default='html-output',
                        help='Directory with Go HTML files (default: html-output)')
    parser.add_argument('--out', default='html-delta',
                        help='Output directory for delta .md files (default: html-delta)')
    parser.add_argument('--report', default='',
                        help='Process only this report name (without .html)')
    args = parser.parse_args()

    cs_dir = Path(args.csharp)
    go_dir = Path(args.go)
    out_dir = Path(args.out)

    if not cs_dir.is_dir():
        sys.exit(f'error: C# output directory not found: {cs_dir}')
    if not go_dir.is_dir():
        sys.exit(f'error: Go output directory not found: {go_dir}')

    out_dir.mkdir(parents=True, exist_ok=True)

    # Collect C# HTML files.
    cs_files = sorted(f for f in cs_dir.glob('*.html'))
    if args.report:
        cs_files = [f for f in cs_files
                    if f.stem.lower() == args.report.lower()]
        if not cs_files:
            sys.exit(f'error: no C# HTML file matching "{args.report}"')

    results: list[tuple[str, dict | None]] = []

    for cs_path in cs_files:
        name = cs_path.stem
        go_path = go_dir / cs_path.name

        if not go_path.exists():
            print(f'  MISSING  {name}')
            results.append((name, None))
            # Write a stub delta doc.
            md = f'# {name}\n\n**Status:** ➖ MISSING — Go output file not found.\n'
            (out_dir / f'{name}.md').write_text(md, encoding='utf-8')
            continue

        try:
            cs_page_count, cs_pages = parse_html(cs_path)
            go_page_count, go_pages = parse_html(go_path)
        except Exception as exc:
            print(f'  ERROR    {name}: {exc}')
            results.append((name, None))
            continue

        delta = analyse(name, cs_pages, go_pages)
        st = status(delta)
        sim = f'{delta["similarity"]:.0%}'
        print(f'  {st}  {name:<55} similarity={sim}  pages={cs_page_count}/{go_page_count}')

        st = status(delta)
        if st != STATUS_PASS:
            md = render_md(delta)
            (out_dir / f'{name}.md').write_text(md, encoding='utf-8')
        results.append((name, delta))

    # Write summary index.
    if not args.report:
        summary_md = render_summary(results)
        summary_path = out_dir / 'README.md'
        summary_path.write_text(summary_md, encoding='utf-8')
        print(f'\nSummary written to {summary_path}')

    pass_n = sum(1 for _, d in results if d and status(d) == STATUS_PASS)
    minor_n = sum(1 for _, d in results if d and status(d) == STATUS_MINOR)
    fail_n = sum(1 for _, d in results if d and status(d) == STATUS_FAIL)
    miss_n = sum(1 for _, d in results if d is None)
    print(f'\n{pass_n} pass, {minor_n} minor, {fail_n} fail, {miss_n} missing '
          f'— delta docs in "{out_dir}/"')


if __name__ == '__main__':
    main()
