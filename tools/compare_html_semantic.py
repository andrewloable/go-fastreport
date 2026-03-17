#!/usr/bin/env python3
"""compare_html_semantic.py — Semantic comparison of Go vs C# HTML output.

Parses both HTML files, extracts per-page elements (position, size, text, style),
normalises number formatting, and compares structurally.

The C# output is ground-truth (expected). The Go output is the actual.

Usage:
    python3 tools/compare_html_semantic.py
    python3 tools/compare_html_semantic.py --report "Simple List"
"""

import argparse
import difflib
import re
import sys
from html.parser import HTMLParser
from pathlib import Path


# ---------------------------------------------------------------------------
# HTML parsing
# ---------------------------------------------------------------------------

_PAGE_RE = re.compile(r'\bfrpage(\d+)\b')
_NUM_RE = re.compile(r'(-?\d+(?:\.\d+)?)')


def _parse_style(style: str) -> dict[str, str]:
    """Parse a CSS style string into a dict of property→value."""
    result = {}
    for part in style.split(';'):
        part = part.strip()
        if ':' in part:
            k, v = part.split(':', 1)
            result[k.strip().lower()] = v.strip()
    return result


def _norm_num(s: str) -> str:
    """Normalise a CSS value: strip trailing zeros from numbers."""
    def _repl(m):
        n = m.group(0)
        if '.' in n:
            n = n.rstrip('0').rstrip('.')
        return n
    return _NUM_RE.sub(_repl, s)


def _norm_style(style_dict: dict) -> dict:
    """Normalise style values for comparison."""
    return {k: _norm_num(v) for k, v in style_dict.items()}


class Element:
    """A positioned element extracted from an HTML page."""
    __slots__ = ('tag', 'cls', 'style', 'left', 'top', 'width', 'height',
                 'text', 'raw_style', 'bg_color', 'color', 'font', 'border',
                 'text_align', 'vert_align')

    def __init__(self, tag, cls, style_str, text=''):
        self.tag = tag
        self.cls = cls
        self.raw_style = style_str
        self.text = text
        sd = _parse_style(style_str)
        nsd = _norm_style(sd)
        self.left = nsd.get('left', '')
        self.top = nsd.get('top', '')
        self.width = nsd.get('width', '')
        self.height = nsd.get('height', '')
        self.bg_color = nsd.get('background-color', '')
        self.color = nsd.get('color', '')
        self.font = nsd.get('font', nsd.get('font-family', ''))
        self.border = nsd.get('border', '')
        self.text_align = nsd.get('text-align', '')
        self.vert_align = nsd.get('vertical-align', '')

    def pos_key(self) -> str:
        return f'{self.left},{self.top}'

    def geom_key(self) -> str:
        return f'{self.left},{self.top},{self.width},{self.height}'

    def __repr__(self):
        t = self.text[:30] if self.text else ''
        return f'Elem({self.left},{self.top} {self.width}x{self.height} "{t}")'


class Page:
    """A parsed page from HTML output."""
    def __init__(self, idx, cls, style_str):
        self.idx = idx
        self.cls = cls
        sd = _parse_style(style_str)
        nsd = _norm_style(sd)
        self.width = nsd.get('width', '')
        self.height = nsd.get('height', '')
        self.bg_color = nsd.get('background-color', '')
        self.elements: list[Element] = []
        self.texts: list[str] = []

    def __repr__(self):
        return f'Page({self.idx}, {self.width}x{self.height}, {len(self.elements)} elems)'


class HtmlExtractor(HTMLParser):
    """Extract pages and positioned elements from FastReport HTML."""

    def __init__(self):
        super().__init__()
        self.pages: list[Page] = []
        self._page: Page | None = None
        self._skip = 0
        self._stack: list[str] = []
        self._text_parts: list[str] = []
        self._cur_elem: Element | None = None
        self._depth = 0

    def handle_starttag(self, tag, attrs):
        self._depth += 1
        self._stack.append(tag)
        if tag in ('style', 'script', 'head'):
            self._skip += 1
            return
        if self._skip:
            return

        d = dict(attrs)
        cls = d.get('class', '')
        style = d.get('style', '')

        m = _PAGE_RE.search(cls)
        if m:
            self._page = Page(int(m.group(1)), cls, style)
            self.pages.append(self._page)
            return

        if self._page is not None and style and ('left:' in style or 'left:' in style.lower()):
            self._flush_text()
            self._cur_elem = Element(tag, cls, style)
            self._page.elements.append(self._cur_elem)

    def handle_endtag(self, tag):
        if self._stack and self._stack[-1] == tag:
            self._stack.pop()
        if tag in ('style', 'script', 'head') and self._skip > 0:
            self._skip -= 1
        self._depth -= 1
        # If we're closing the element that held positioned content, flush text.
        if self._cur_elem is not None and tag == self._cur_elem.tag:
            self._flush_text()
            self._cur_elem = None

    def handle_data(self, data):
        if self._skip:
            return
        text = data.strip()
        if not text:
            return
        if self._page is not None:
            self._text_parts.append(text)
            self._page.texts.append(text)

    def _flush_text(self):
        if self._cur_elem and self._text_parts:
            self._cur_elem.text = ' '.join(self._text_parts)
        self._text_parts = []


def parse_html(path: Path) -> list[Page]:
    p = HtmlExtractor()
    p.feed(path.read_text(encoding='utf-8', errors='replace'))
    return p.pages


# ---------------------------------------------------------------------------
# CSS class extraction (for style comparison)
# ---------------------------------------------------------------------------

_CSS_CLASS_RE = re.compile(r'\.(s\d+)\s*\{([^}]+)\}')


def extract_css_classes(html_text: str) -> dict[str, dict[str, str]]:
    """Extract .sN { ... } CSS class definitions into a dict of class→properties."""
    result = {}
    for m in _CSS_CLASS_RE.finditer(html_text):
        cls_name = m.group(1)
        props = _parse_style(m.group(2))
        result[cls_name] = _norm_style(props)
    return result


# ---------------------------------------------------------------------------
# Comparison
# ---------------------------------------------------------------------------

class Delta:
    """Holds all differences between expected and actual for one report."""
    def __init__(self, name):
        self.name = name
        self.page_count_exp = 0
        self.page_count_act = 0
        self.page_dim_diffs: list[str] = []       # per-page dimension mismatches
        self.missing_texts: list[tuple[int, str]] = []   # (page, text) present in C# but not Go
        self.extra_texts: list[tuple[int, str]] = []     # (page, text) present in Go but not C#
        self.position_diffs: list[str] = []        # element position/size mismatches
        self.element_count_diffs: list[str] = []   # per-page element count differences
        self.text_content_diffs: list[str] = []    # text value mismatches at same position
        self.style_diffs: list[str] = []           # color, font, border differences
        self.text_order_diffs: list[str] = []      # text ordering differences (sequence)
        self.css_class_diffs: list[str] = []       # CSS class definition differences

    def is_pass(self) -> bool:
        return (self.page_count_exp == self.page_count_act
                and not self.missing_texts
                and not self.extra_texts
                and not self.position_diffs
                and not self.text_content_diffs
                and not self.style_diffs
                and not self.element_count_diffs
                and not self.text_order_diffs)

    def severity(self) -> str:
        if self.page_count_exp != self.page_count_act:
            return 'PAGE_COUNT'
        if self.missing_texts or self.extra_texts:
            return 'TEXT_CONTENT'
        if self.text_content_diffs:
            return 'TEXT_VALUES'
        if self.position_diffs:
            return 'POSITIONING'
        if self.style_diffs:
            return 'STYLING'
        if self.element_count_diffs:
            return 'ELEMENTS'
        if self.text_order_diffs:
            return 'TEXT_ORDER'
        return 'PASS'


def compare_pages(name: str, exp_pages: list[Page], act_pages: list[Page],
                  exp_css: dict, act_css: dict) -> Delta:
    d = Delta(name)
    d.page_count_exp = len(exp_pages)
    d.page_count_act = len(act_pages)

    n = min(len(exp_pages), len(act_pages))
    for i in range(n):
        ep = exp_pages[i]
        ap = act_pages[i]

        # Page dimensions.
        if ep.width != ap.width or ep.height != ap.height:
            d.page_dim_diffs.append(
                f'Page {i}: dim C#={ep.width}x{ep.height} Go={ap.width}x{ap.height}')

        # Element counts.
        if len(ep.elements) != len(ap.elements):
            d.element_count_diffs.append(
                f'Page {i}: C#={len(ep.elements)} elements, Go={len(ap.elements)} elements')

        # Text content (set comparison).
        exp_texts = set(ep.texts)
        act_texts = set(ap.texts)
        for t in sorted(exp_texts - act_texts):
            d.missing_texts.append((i, t))
        for t in sorted(act_texts - exp_texts):
            d.extra_texts.append((i, t))

        # Text sequence comparison.
        if ep.texts != ap.texts:
            sm = difflib.SequenceMatcher(None, ep.texts, ap.texts, autojunk=False)
            for tag, i1, i2, j1, j2 in sm.get_opcodes():
                if tag == 'equal':
                    continue
                cs_chunk = ep.texts[i1:i2]
                go_chunk = ap.texts[j1:j2]
                if tag == 'replace':
                    d.text_order_diffs.append(
                        f'Page {i}: replace C#={cs_chunk[:3]} -> Go={go_chunk[:3]}')
                elif tag == 'delete':
                    d.text_order_diffs.append(f'Page {i}: missing C#={cs_chunk[:3]}')
                elif tag == 'insert':
                    d.text_order_diffs.append(f'Page {i}: extra Go={go_chunk[:3]}')
                if len(d.text_order_diffs) > 30:
                    d.text_order_diffs.append('... (truncated)')
                    break

        # Element-by-element comparison (match by position).
        exp_by_pos = {}
        for e in ep.elements:
            key = e.pos_key()
            if key not in exp_by_pos:
                exp_by_pos[key] = []
            exp_by_pos[key].append(e)

        act_by_pos = {}
        for e in ap.elements:
            key = e.pos_key()
            if key not in act_by_pos:
                act_by_pos[key] = []
            act_by_pos[key].append(e)

        # Compare elements at matching positions.
        all_positions = set(exp_by_pos.keys()) | set(act_by_pos.keys())
        for pos in sorted(all_positions):
            exp_elems = exp_by_pos.get(pos, [])
            act_elems = act_by_pos.get(pos, [])
            for j in range(min(len(exp_elems), len(act_elems))):
                ee = exp_elems[j]
                ae = act_elems[j]
                # Size comparison.
                if ee.width != ae.width or ee.height != ae.height:
                    d.position_diffs.append(
                        f'Page {i} @{pos}: size C#={ee.width}x{ee.height} Go={ae.width}x{ae.height}')
                # Text at same position.
                if ee.text and ae.text and ee.text != ae.text:
                    d.text_content_diffs.append(
                        f'Page {i} @{pos}: C#="{ee.text[:60]}" Go="{ae.text[:60]}"')
                # Background color.
                if ee.bg_color and ae.bg_color and ee.bg_color != ae.bg_color:
                    d.style_diffs.append(
                        f'Page {i} @{pos}: bg C#={ee.bg_color} Go={ae.bg_color}')

            # Elements at a position in C# but not in Go.
            if len(exp_elems) > len(act_elems) and len(d.position_diffs) < 50:
                for j in range(len(act_elems), len(exp_elems)):
                    d.position_diffs.append(
                        f'Page {i} @{pos}: missing in Go: {exp_elems[j]}')
            if len(act_elems) > len(exp_elems) and len(d.position_diffs) < 50:
                for j in range(len(exp_elems), len(act_elems)):
                    d.position_diffs.append(
                        f'Page {i} @{pos}: extra in Go: {act_elems[j]}')

    # Truncate large diffs.
    for attr in ('position_diffs', 'text_content_diffs', 'style_diffs',
                 'text_order_diffs', 'element_count_diffs'):
        lst = getattr(d, attr)
        if len(lst) > 50:
            n_more = len(lst) - 50
            del lst[50:]
            lst.append(f'... and {n_more} more')
    # Truncate tuple lists separately.
    for attr in ('missing_texts', 'extra_texts'):
        lst = getattr(d, attr)
        if len(lst) > 50:
            n_more = len(lst) - 50
            del lst[50:]
            lst.append((-1, f'... and {n_more} more'))

    return d


# ---------------------------------------------------------------------------
# Markdown report
# ---------------------------------------------------------------------------

def render_md(d: Delta) -> str:
    lines = [f'# {d.name}', '']

    status = 'PASS' if d.is_pass() else f'FAIL ({d.severity()})'
    lines.append(f'**Status:** {status}')
    lines.append(f'**Pages:** C# = {d.page_count_exp}, Go = {d.page_count_act}'
                 + ('' if d.page_count_exp == d.page_count_act
                    else f'  (mismatch: {d.page_count_act - d.page_count_exp:+d})'))
    lines.append('')

    if d.page_dim_diffs:
        lines.append('## Page Dimension Differences')
        lines.append('')
        for s in d.page_dim_diffs:
            lines.append(f'- {s}')
        lines.append('')

    if d.element_count_diffs:
        lines.append('## Element Count Differences')
        lines.append('')
        for s in d.element_count_diffs:
            lines.append(f'- {s}')
        lines.append('')

    if d.missing_texts:
        lines.append('## Text Present in C# but Missing from Go')
        lines.append('')
        for pg, t in d.missing_texts:
            lines.append(f'- Page {pg}: `{t[:80]}`')
        lines.append('')

    if d.extra_texts:
        lines.append('## Text Present in Go but Absent from C#')
        lines.append('')
        for pg, t in d.extra_texts:
            lines.append(f'- Page {pg}: `{t[:80]}`')
        lines.append('')

    if d.text_content_diffs:
        lines.append('## Text Value Differences (Same Position)')
        lines.append('')
        for s in d.text_content_diffs:
            lines.append(f'- {s}')
        lines.append('')

    if d.position_diffs:
        lines.append('## Position / Size Differences')
        lines.append('')
        for s in d.position_diffs:
            lines.append(f'- {s}')
        lines.append('')

    if d.style_diffs:
        lines.append('## Style Differences')
        lines.append('')
        for s in d.style_diffs:
            lines.append(f'- {s}')
        lines.append('')

    if d.text_order_diffs:
        lines.append('## Text Order Differences')
        lines.append('')
        lines.append('```')
        for s in d.text_order_diffs:
            lines.append(s)
        lines.append('```')
        lines.append('')

    if d.is_pass():
        lines.append('_No significant differences detected._')
        lines.append('')

    return '\n'.join(lines)


def render_summary(results: list[tuple[str, Delta | None]]) -> str:
    lines = ['# HTML Semantic Comparison Summary', '']
    lines.append('Semantic comparison of Go HTML output vs C# ground-truth.')
    lines.append('')

    pass_list, fail_list, missing_list = [], [], []
    severity_counts: dict[str, int] = {}
    for name, d in results:
        if d is None:
            missing_list.append(name)
        elif d.is_pass():
            pass_list.append((name, d))
        else:
            fail_list.append((name, d))
            sev = d.severity()
            severity_counts[sev] = severity_counts.get(sev, 0) + 1

    total = len(results)
    lines.append('| Category | Count |')
    lines.append('|---|---|')
    lines.append(f'| Pass | {len(pass_list)} |')
    lines.append(f'| Fail | {len(fail_list)} |')
    lines.append(f'| Go missing | {len(missing_list)} |')
    lines.append(f'| **Total** | **{total}** |')
    lines.append('')

    if severity_counts:
        lines.append('### Failure Breakdown')
        lines.append('')
        lines.append('| Severity | Count |')
        lines.append('|---|---|')
        for sev in sorted(severity_counts, key=lambda s: -severity_counts[s]):
            lines.append(f'| {sev} | {severity_counts[sev]} |')
        lines.append('')

    if fail_list:
        lines.append('## Failures')
        lines.append('')
        lines.append('| Report | Severity | Pages (C#/Go) | Missing Texts | Extra Texts | Pos Diffs |')
        lines.append('|---|---|---|---|---|---|')
        for name, d in fail_list:
            link = name.replace(' ', '%20')
            pg = f'{d.page_count_exp}/{d.page_count_act}'
            if d.page_count_exp != d.page_count_act:
                pg += ' !'
            miss = len(d.missing_texts)
            extra = len(d.extra_texts)
            pos = len(d.position_diffs)
            lines.append(f'| [{name}]({link}.md) | {d.severity()} | {pg} | {miss} | {extra} | {pos} |')
        lines.append('')

    if pass_list:
        lines.append('## Passing')
        lines.append('')
        lines.append('| Report | Pages |')
        lines.append('|---|---|')
        for name, d in pass_list:
            lines.append(f'| {name} | {d.page_count_exp} |')
        lines.append('')

    if missing_list:
        lines.append('## Go Output Missing')
        lines.append('')
        for name in missing_list:
            lines.append(f'- {name}')
        lines.append('')

    return '\n'.join(lines)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(description=__doc__,
                                     formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument('--csharp', default='csharp-html-output')
    parser.add_argument('--go', default='html-output')
    parser.add_argument('--out', default='html-delta')
    parser.add_argument('--report', default='')
    args = parser.parse_args()

    cs_dir = Path(args.csharp)
    go_dir = Path(args.go)
    out_dir = Path(args.out)

    if not cs_dir.is_dir():
        sys.exit(f'error: not found: {cs_dir}')
    if not go_dir.is_dir():
        sys.exit(f'error: not found: {go_dir}')

    out_dir.mkdir(parents=True, exist_ok=True)

    cs_files = sorted(cs_dir.glob('*.html'))
    if args.report:
        cs_files = [f for f in cs_files if f.stem.lower() == args.report.lower()]
        if not cs_files:
            sys.exit(f'error: no match for "{args.report}"')

    results: list[tuple[str, Delta | None]] = []

    for cs_path in cs_files:
        name = cs_path.stem
        go_path = go_dir / cs_path.name

        if not go_path.exists():
            print(f'  MISSING  {name}')
            results.append((name, None))
            md = f'# {name}\n\n**Status:** MISSING\n'
            (out_dir / f'{name}.md').write_text(md, encoding='utf-8')
            continue

        try:
            cs_html = cs_path.read_text(encoding='utf-8', errors='replace')
            go_html = go_path.read_text(encoding='utf-8', errors='replace')
            exp_pages = parse_html(cs_path)
            act_pages = parse_html(go_path)
            exp_css = extract_css_classes(cs_html)
            act_css = extract_css_classes(go_html)
        except Exception as exc:
            print(f'  ERROR    {name}: {exc}')
            results.append((name, None))
            continue

        delta = compare_pages(name, exp_pages, act_pages, exp_css, act_css)

        if delta.is_pass():
            print(f'  PASS     {name:<55} pages={delta.page_count_exp}')
            old = out_dir / f'{name}.md'
            if old.exists():
                old.unlink()
        else:
            sev = delta.severity()
            miss = len(delta.missing_texts)
            extra = len(delta.extra_texts)
            print(f'  FAIL     {name:<55} {sev:<14} '
                  f'pages={delta.page_count_exp}/{delta.page_count_act}  '
                  f'miss={miss} extra={extra}')
            md = render_md(delta)
            (out_dir / f'{name}.md').write_text(md, encoding='utf-8')

        results.append((name, delta))

    if not args.report:
        summary = render_summary(results)
        (out_dir / 'README.md').write_text(summary, encoding='utf-8')
        print(f'\nSummary written to {out_dir}/README.md')

    pass_n = sum(1 for _, d in results if d and d.is_pass())
    fail_n = sum(1 for _, d in results if d and not d.is_pass())
    miss_n = sum(1 for _, d in results if d is None)
    print(f'\n{pass_n} pass, {fail_n} fail, {miss_n} missing')


if __name__ == '__main__':
    main()
