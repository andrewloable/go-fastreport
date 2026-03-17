#!/usr/bin/env python3
"""compare_html_bytes.py — Byte-level comparison of Go vs C# HTML output.

For each .html file in csharp-html-output/ that also exists in html-output/,
produce a detailed character-level diff and write it to html-delta/<filename>.md.

The C# output is the ground-truth (expected). The Go output is the actual.

Usage:
    python3 tools/compare_html_bytes.py
    python3 tools/compare_html_bytes.py --csharp csharp-html-output --go html-output --out html-delta
    python3 tools/compare_html_bytes.py --report "Simple List"
"""

import argparse
import difflib
import os
import sys
from pathlib import Path


def byte_stats(expected: bytes, actual: bytes) -> dict:
    """Compare two byte sequences and return stats."""
    if expected == actual:
        return {
            'identical': True,
            'expected_len': len(expected),
            'actual_len': len(actual),
            'len_diff': 0,
        }

    # Find first differing byte position.
    first_diff = -1
    for i in range(min(len(expected), len(actual))):
        if expected[i] != actual[i]:
            first_diff = i
            break
    if first_diff == -1:
        # One is a prefix of the other.
        first_diff = min(len(expected), len(actual))

    return {
        'identical': False,
        'expected_len': len(expected),
        'actual_len': len(actual),
        'len_diff': len(actual) - len(expected),
        'first_diff_byte': first_diff,
    }


def unified_diff_lines(expected_text: str, actual_text: str, max_lines: int = 300) -> list[str]:
    """Return a unified diff between expected and actual as a list of lines."""
    exp_lines = expected_text.splitlines(keepends=True)
    act_lines = actual_text.splitlines(keepends=True)
    diff = list(difflib.unified_diff(
        exp_lines, act_lines,
        fromfile='expected (C#)',
        tofile='actual (Go)',
        lineterm='',
    ))
    truncated = False
    if len(diff) > max_lines:
        diff = diff[:max_lines]
        truncated = True
    return diff, truncated


def context_around(text: str, byte_pos: int, radius: int = 80) -> str:
    """Return a snippet of text around the given byte position."""
    start = max(0, byte_pos - radius)
    end = min(len(text), byte_pos + radius)
    snippet = text[start:end]
    # Mark the diff position.
    marker_pos = byte_pos - start
    before = snippet[:marker_pos]
    after = snippet[marker_pos:]
    return before, after


def render_md(name: str, stats: dict, expected_text: str, actual_text: str) -> str:
    """Generate a markdown report for one file comparison."""
    lines = []
    lines.append(f'# {name}')
    lines.append('')

    if stats['identical']:
        lines.append('**Status:** PASS (identical)')
        lines.append(f'**Size:** {stats["expected_len"]:,} bytes')
        lines.append('')
        return '\n'.join(lines)

    lines.append('**Status:** FAIL (differences found)')
    lines.append(f'**Expected size (C#):** {stats["expected_len"]:,} bytes')
    lines.append(f'**Actual size (Go):** {stats["actual_len"]:,} bytes')
    lines.append(f'**Size difference:** {stats["len_diff"]:+,} bytes')
    lines.append('')

    if 'first_diff_byte' in stats:
        pos = stats['first_diff_byte']
        lines.append(f'**First difference at byte:** {pos:,}')
        lines.append('')

        # Show context around the first difference.
        lines.append('## Context Around First Difference')
        lines.append('')

        if pos < len(expected_text):
            before_exp, after_exp = context_around(expected_text, pos, 120)
            lines.append('**Expected (C#):**')
            lines.append('```')
            lines.append(f'{before_exp}>>>HERE<<<{after_exp}')
            lines.append('```')
            lines.append('')

        if pos < len(actual_text):
            before_act, after_act = context_around(actual_text, pos, 120)
            lines.append('**Actual (Go):**')
            lines.append('```')
            lines.append(f'{before_act}>>>HERE<<<{after_act}')
            lines.append('```')
            lines.append('')

    # Line-level diff summary.
    exp_lines = expected_text.splitlines()
    act_lines = actual_text.splitlines()
    lines.append(f'**Expected lines:** {len(exp_lines):,}')
    lines.append(f'**Actual lines:** {len(act_lines):,}')
    lines.append('')

    # Compute similarity ratio on lines.
    matcher = difflib.SequenceMatcher(None, exp_lines, act_lines, autojunk=False)
    ratio = matcher.ratio()
    lines.append(f'**Line similarity:** {ratio:.1%}')
    lines.append('')

    # Unified diff.
    diff, truncated = unified_diff_lines(expected_text, actual_text, max_lines=500)
    if diff:
        lines.append('## Unified Diff')
        lines.append('')
        if truncated:
            lines.append('_(Truncated to first 500 lines of diff)_')
            lines.append('')
        lines.append('```diff')
        for d in diff:
            lines.append(d.rstrip('\n'))
        lines.append('```')
        lines.append('')

    return '\n'.join(lines)


def render_summary(results: list[tuple[str, dict]]) -> str:
    """Generate the summary README.md."""
    lines = []
    lines.append('# HTML Byte-Level Comparison Summary')
    lines.append('')
    lines.append('Character-level comparison of Go HTML output (`html-output/`) against '
                 'C# ground-truth (`csharp-html-output/`).')
    lines.append('')

    pass_list = []
    fail_list = []
    missing_list = []
    for name, stats in results:
        if stats is None:
            missing_list.append(name)
        elif stats['identical']:
            pass_list.append((name, stats))
        else:
            fail_list.append((name, stats))

    total = len(results)
    lines.append('| Category | Count |')
    lines.append('|---|---|')
    lines.append(f'| Identical | {len(pass_list)} |')
    lines.append(f'| Different | {len(fail_list)} |')
    lines.append(f'| Go output missing | {len(missing_list)} |')
    lines.append(f'| **Total** | **{total}** |')
    lines.append('')

    if fail_list:
        lines.append('## Different')
        lines.append('')
        lines.append('| Report | C# Size | Go Size | Size Diff | First Diff Byte |')
        lines.append('|---|---|---|---|---|')
        for name, stats in fail_list:
            link = name.replace(' ', '%20')
            cs_sz = f'{stats["expected_len"]:,}'
            go_sz = f'{stats["actual_len"]:,}'
            sz_diff = f'{stats["len_diff"]:+,}'
            first = f'{stats.get("first_diff_byte", "?"):,}' if 'first_diff_byte' in stats else '?'
            lines.append(f'| [{name}]({link}.md) | {cs_sz} | {go_sz} | {sz_diff} | {first} |')
        lines.append('')

    if pass_list:
        lines.append('## Identical')
        lines.append('')
        lines.append('| Report | Size |')
        lines.append('|---|---|')
        for name, stats in pass_list:
            lines.append(f'| {name} | {stats["expected_len"]:,} |')
        lines.append('')

    if missing_list:
        lines.append('## Go Output Missing')
        lines.append('')
        for name in missing_list:
            lines.append(f'- {name}')
        lines.append('')

    return '\n'.join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__,
                                     formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument('--csharp', default='csharp-html-output',
                        help='Directory with C# ground-truth HTML files')
    parser.add_argument('--go', default='html-output',
                        help='Directory with Go HTML files')
    parser.add_argument('--out', default='html-delta',
                        help='Output directory for delta .md files')
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
            md = f'# {name}\n\n**Status:** MISSING — Go output file not found.\n'
            (out_dir / f'{name}.md').write_text(md, encoding='utf-8')
            continue

        try:
            expected_bytes = cs_path.read_bytes()
            actual_bytes = go_path.read_bytes()
        except Exception as exc:
            print(f'  ERROR    {name}: {exc}')
            results.append((name, None))
            continue

        stats = byte_stats(expected_bytes, actual_bytes)

        if stats['identical']:
            print(f'  PASS     {name:<55} {stats["expected_len"]:>10,} bytes')
            # Remove any old delta file for this report.
            old_md = out_dir / f'{name}.md'
            if old_md.exists():
                old_md.unlink()
        else:
            expected_text = expected_bytes.decode('utf-8', errors='replace')
            actual_text = actual_bytes.decode('utf-8', errors='replace')
            md = render_md(name, stats, expected_text, actual_text)
            (out_dir / f'{name}.md').write_text(md, encoding='utf-8')
            first_diff = stats.get('first_diff_byte', '?')
            print(f'  FAIL     {name:<55} C#={stats["expected_len"]:>10,}  Go={stats["actual_len"]:>10,}  diff@{first_diff}')

        results.append((name, stats))

    # Write summary.
    if not args.report:
        summary_md = render_summary(results)
        summary_path = out_dir / 'README.md'
        summary_path.write_text(summary_md, encoding='utf-8')
        print(f'\nSummary written to {summary_path}')

    pass_n = sum(1 for _, s in results if s and s['identical'])
    fail_n = sum(1 for _, s in results if s and not s['identical'])
    miss_n = sum(1 for _, s in results if s is None)
    print(f'\n{pass_n} identical, {fail_n} different, {miss_n} missing — delta docs in "{out_dir}/"')


if __name__ == '__main__':
    main()
