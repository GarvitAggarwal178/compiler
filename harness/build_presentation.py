#!/usr/bin/env python3
"""
PUNCH-LIST.md P6 (per `docs/m2 m3.md` section 10 and NIGHT-BATCH-04 G,
unchanged): one static HTML presentation artifact, reading committed
measurement JSON plus `dlc explain` output. No server, no framework, no
build step, no state, no JavaScript beyond a few inert tab toggles.

Four things visible, per instruction:
1. The analyzer rejecting programs, all four grounds, with spans.
2. Three-column metric per shape, incl-sup headline.
3. The guard firing and declining, reason named, cone shown.
4. Two findings with mechanisms: bb->bf demand relaxation before/after,
   and cone behaviour.

Every number here is read from an already-committed file -- this script
computes nothing new, it only renders.
"""
import html
import json
import re
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
EXPLAIN_DIR = REPO / "docs" / "reports" / "explain-samples"
HEADLINE_JSON = REPO / "measurements" / "m3-5-headline-m4" / "summary.json"
CONE_JSON = REPO / "measurements" / "night04-b-cone" / "summary.json"
BLAST_DIR = REPO / "measurements" / "night04-b-cone" / "blast_radius"
BEFORE_DIR = REPO / "measurements" / "m4-sips" / "before"
AFTER_DIR = REPO / "measurements" / "m4-sips" / "after"
DECOMPOSE_JSON = REPO / "measurements" / "punch-list-2" / "p1-decompose" / "summary.json"
GROWING_JSON = REPO / "measurements" / "punch-list-2" / "item2-growing-sibling" / "summary.json"
OUT_PATH = REPO / "docs" / "reports" / "presentation.html"

REJECTION_SAMPLES = [
    ("arity", "rejection_arity"),
    ("type", "rejection_type"),
    ("allowedness", "rejection_allowedness"),
    ("unstratifiable negation", "rejection_stratification"),
]

SHAPE_ORDER = [
    "reachability_complement", "ancestor_nonancestor",
    "same_generation_negation", "transitive_closure_bound", "culprit_cycle",
]

ADORN_RE = re.compile(r"\.decl ([a-zA-Z_][a-zA-Z0-9_]*_[bf]+)\(")


def esc(s):
    return html.escape(str(s))


def adornment_set(path: Path):
    text = path.read_text()
    return sorted(set(m for m in ADORN_RE.findall(text) if not m.startswith("magic_") and not m.startswith("sup_")))


def build_rejection_section():
    rows = []
    for ground, base in REJECTION_SAMPLES:
        dl_text = (EXPLAIN_DIR / f"{base}.dl").read_text()
        explain_text = (EXPLAIN_DIR / f"{base}.explain.txt").read_text().strip()
        m = re.search(r"REJECT ground=(\S+) span=(\S+) message=\"(.*)\"", explain_text)
        g, span, message = (m.group(1), m.group(2), m.group(3)) if m else (ground, "?", explain_text)
        rows.append(f"""
        <div class="card">
          <h4>{esc(ground)}</h4>
          <pre class="code">{esc(dl_text.strip())}</pre>
          <div class="verdict">
            <span class="tag reject">REJECT</span>
            <code>ground={esc(g)}</code>
            <code>span={esc(span)}</code>
          </div>
          <p class="msg">&ldquo;{esc(message)}&rdquo;</p>
        </div>""")
    return "\n".join(rows)


def build_metric_section():
    data = json.loads(HEADLINE_JSON.read_text())
    blocks = []
    for shape in SHAPE_ORDER:
        rows = data.get(shape, [])
        trs = []
        for r in rows:
            if r.get("status_dlc") != "ok":
                continue
            t_none, t_souffle, t_dlc = r["T_none"], r["T_souffle"], r["T_dlc"]
            contribution = t_souffle / t_dlc if t_dlc else float("inf")
            ident = "&check;" if r.get("answers_identical") else "&cross;"
            fired = "yes" if r.get("guard_fired") else "no"
            trs.append(f"""<tr>
                <td>{esc(r['tag'])}</td><td>{t_none:,}</td><td>{t_souffle:,}</td>
                <td>{t_dlc:,}</td><td class="ratio">{contribution:,.1f}&times;</td>
                <td class="center">{ident}</td><td class="center">{fired}</td></tr>""")
        blocks.append(f"""
        <div class="card">
          <h4>{esc(shape)}</h4>
          <table>
            <thead><tr><th>scale</th><th>T_none</th><th>T_souffle</th><th>T_dlc</th>
            <th>T_souffle/T_dlc</th><th>answers=</th><th>guard fired</th></tr></thead>
            <tbody>{''.join(trs)}</tbody>
          </table>
        </div>""")
    return "\n".join(blocks)


def build_guard_section():
    cone = json.loads(CONE_JSON.read_text())
    blast_files = sorted(BLAST_DIR.glob("*.json"))
    blast_rows = []
    for f in blast_files:
        d = json.loads(f.read_text())
        name = f.stem
        idb = len(d["idb_relations"])
        culprit = d["culprit_predicates"]
        cone_rel = d["cone_relations"]
        declined = d["declined_relations"]
        frac = len(declined) / idb if idb else 0.0
        fires = "yes" if culprit else "no"
        blast_rows.append(f"""<tr>
            <td>{esc(name)}</td><td class="center">{idb}</td>
            <td class="center">{fires}</td>
            <td><code>{{{', '.join(culprit)}}}</code></td>
            <td><code>{{{', '.join(cone_rel)}}}</code></td>
            <td class="center">{len(declined)}/{idb}</td>
            <td class="ratio">{frac:.2f}</td></tr>""")

    cross_rows = []
    for name, v in cone.get("cone_cross_check", {}).items():
        agree = "&check; exact" if v["cone_agrees"] else "&cross; MISMATCH"
        cross_rows.append(f"""<tr>
            <td>{esc(name)}</td>
            <td><code>{{{', '.join(v['go_cone'])}}}</code></td>
            <td><code>{{{', '.join(v['py_cone'])}}}</code></td>
            <td class="center">{agree}</td></tr>""")

    return f"""
    <div class="card">
      <h4>Blast radius, per program</h4>
      <table>
        <thead><tr><th>program</th><th>IDB</th><th>fires?</th><th>culprit SCC</th>
        <th>cone</th><th>declined/total</th><th>fraction</th></tr></thead>
        <tbody>{''.join(blast_rows)}</tbody>
      </table>
    </div>
    <div class="card">
      <h4>Cone cross-check: <code>guard.Decide</code> (Go) vs <code>cone_metric.py</code> (independent Python)</h4>
      <table>
        <thead><tr><th>program</th><th>Go cone</th><th>Python cone</th><th>agreement</th></tr></thead>
        <tbody>{''.join(cross_rows)}</tbody>
      </table>
    </div>"""


def build_findings_section():
    shapes = ["p2", "reachability_complement", "ancestor_nonancestor", "same_generation_negation"]
    rows = []
    for s in shapes:
        before = adornment_set(BEFORE_DIR / f"{s}.transformed.dl")
        after = adornment_set(AFTER_DIR / f"{s}.transformed.dl")
        rows.append(f"""<tr>
            <td>{esc(s)}</td>
            <td><code>{{{', '.join(before)}}}</code></td>
            <td><code>{{{', '.join(after)}}}</code></td></tr>""")

    cone_data = json.loads(CONE_JSON.read_text())
    cone_rows = []
    for r in cone_data.get("points", []):
        lt = r.get("T_guarded_lt_T_none")
        mark = "&check;" if lt else ("&mdash;" if lt is False else "?")
        cone_rows.append(f"""<tr>
            <td>{esc(r['program'])}</td><td>{r['n']}</td>
            <td>{r.get('T_none', '-'):,}</td><td>{r.get('T_dlc', '-'):,}</td>
            <td class="center">{mark}</td></tr>""")

    decompose = json.loads(DECOMPOSE_JSON.read_text()) if DECOMPOSE_JSON.is_file() else []
    decompose_rows = []
    for r in decompose:
        ratio = r["T_none"] / r["T_guarded"] if r["T_guarded"] else 0
        decompose_rows.append(f"""<tr>
            <td>{esc(r['program'])}</td><td>{r['n']}</td>
            <td>{r['declined_portion']:,}</td><td>{r['transformed_portion']:,}</td>
            <td class="ratio">{ratio:.2f}&times;</td></tr>""")

    growing = json.loads(GROWING_JSON.read_text()) if GROWING_JSON.is_file() else []
    growing_rows = []
    for r in growing:
        ratio = r["ratio_T_none_over_T_guarded"]
        growing_rows.append(f"""<tr>
            <td>{r['n']}</td><td>{r['T_none']:,}</td><td>{r['T_guarded']:,}</td>
            <td>{r['declined_portion']:,}</td><td>{r['transformed_portion']:,}</td>
            <td class="ratio">{ratio:.2f}&times;</td></tr>""")

    return f"""
    <div class="card">
      <h4>Finding 1 &mdash; the <code>bb</code>&rarr;<code>bf</code> demand relaxation, before and after</h4>
      <p>Adornment sets actually generated for each negated occurrence's target
      relation, before M4-SIPS's relaxation and after. Two shapes fully collapse
      to a single adornment; <code>same_generation_negation</code> does not
      (a structural fact about its own recursive rule, disclosed in
      <code>docs/reports/m4-sips.md</code>, not a bug).</p>
      <table>
        <thead><tr><th>program</th><th>before (adornments)</th><th>after (adornments)</th></tr></thead>
        <tbody>{''.join(rows)}</tbody>
      </table>
    </div>
    <div class="card">
      <h4>Finding 2 &mdash; cone behaviour: <code>T_guarded</code> vs <code>T_none</code></h4>
      <p>Task B constructed the first programs in this project with a genuinely
      non-empty fallback cone. PUNCH-LIST P1's multi-query seeding fix then
      made the guard's own contribution measurable for the first time &mdash;
      <code>T_guarded &lt; T_none</code> now holds on 9 of 12 points (was 0/12).</p>
      <table>
        <thead><tr><th>program</th><th>n</th><th>T_none</th><th>T_guarded (T_dlc)</th><th>T_guarded &lt; T_none?</th></tr></thead>
        <tbody>{''.join(cone_rows)}</tbody>
      </table>
    </div>
    <div class="card">
      <h4>Finding 2b &mdash; why the margin shrinks here, decomposed (PUNCH-LIST-2 item 1)</h4>
      <p>The declined portion (always full extent) is bit-for-bit identical to the
      untransformed baseline's own cost at every point &mdash; it grows with
      <code>n</code> by design. The transformed sibling portion does <em>not</em>
      stay constant as first guessed: it shrinks, because the sibling's own
      fixture was never scaled with <code>n</code> in this construction. Both
      effects push the ratio toward 1.0.</p>
      <table>
        <thead><tr><th>program</th><th>n</th><th>declined portion</th><th>transformed portion</th><th>T_none/T_guarded</th></tr></thead>
        <tbody>{''.join(decompose_rows)}</tbody>
      </table>
    </div>
    <div class="card">
      <h4>Finding 2c &mdash; the opposite construction: the margin grows instead (PUNCH-LIST-2 item 2)</h4>
      <p><code>cc_growing_sibling.dl</code> pins the culprit core fixed and lets the
      sibling's own reachable set grow linearly with <code>n</code> &mdash; the
      deliberate opposite of the fixture above. Predicted (Q13) 2&times;&ndash;5&times;
      by n=100; measured direction and mechanism confirmed, magnitude underestimated.</p>
      <table>
        <thead><tr><th>n</th><th>T_none</th><th>T_guarded</th><th>declined portion</th><th>transformed portion</th><th>T_none/T_guarded</th></tr></thead>
        <tbody>{''.join(growing_rows)}</tbody>
      </table>
      <p class="msg">Whether the guard's contribution grows or shrinks with scale is a
      property of the shape being transformed, not of the guard mechanism itself.</p>
    </div>"""


PAGE_TEMPLATE = """<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>dlc &mdash; presentation artifact</title>
<style>
  :root {{
    --bg: #0f1115; --panel: #171a21; --border: #2a2f3a; --text: #e6e8eb;
    --muted: #9aa3af; --accent: #6ee7b7; --accent2: #93c5fd; --bad: #f87171;
  }}
  * {{ box-sizing: border-box; }}
  body {{
    background: var(--bg); color: var(--text); margin: 0; padding: 0 0 4rem;
    font: 15px/1.5 -apple-system, "Segoe UI", Roboto, sans-serif;
  }}
  header {{ padding: 2.5rem 2rem 1rem; border-bottom: 1px solid var(--border); }}
  header h1 {{ margin: 0 0 0.3rem; font-size: 1.8rem; }}
  header p {{ color: var(--muted); margin: 0.2rem 0; max-width: 62rem; }}
  nav {{ display: flex; gap: 0.5rem; padding: 1rem 2rem; flex-wrap: wrap; }}
  nav a {{
    color: var(--text); text-decoration: none; padding: 0.4rem 0.9rem;
    border: 1px solid var(--border); border-radius: 999px; font-size: 0.85rem;
  }}
  nav a:hover {{ border-color: var(--accent2); color: var(--accent2); }}
  main {{ padding: 0 2rem; max-width: 72rem; margin: 0 auto; }}
  section {{ margin: 2.5rem 0; }}
  section > h2 {{
    font-size: 1.3rem; border-left: 4px solid var(--accent2); padding-left: 0.7rem;
    margin-bottom: 0.3rem;
  }}
  section > p.lead {{ color: var(--muted); margin-top: 0; max-width: 60rem; }}
  .card {{
    background: var(--panel); border: 1px solid var(--border); border-radius: 10px;
    padding: 1.1rem 1.3rem; margin: 1rem 0;
  }}
  .card h4 {{ margin: 0 0 0.6rem; font-size: 1rem; }}
  .code, pre.code {{
    background: #0b0d11; border: 1px solid var(--border); border-radius: 6px;
    padding: 0.7rem; font: 12.5px/1.4 "SF Mono", Consolas, monospace;
    white-space: pre-wrap; overflow-x: auto; margin: 0.4rem 0;
  }}
  table {{ width: 100%; border-collapse: collapse; margin-top: 0.5rem; font-size: 0.88rem; }}
  th, td {{ text-align: left; padding: 0.35rem 0.6rem; border-bottom: 1px solid var(--border); }}
  th {{ color: var(--muted); font-weight: 600; font-size: 0.78rem; text-transform: uppercase; letter-spacing: 0.03em; }}
  td.center, th.center {{ text-align: center; }}
  td.ratio {{ color: var(--accent); font-weight: 600; }}
  code {{ font: 12.5px/1.3 "SF Mono", Consolas, monospace; color: var(--accent2); }}
  .tag {{
    display: inline-block; padding: 0.15rem 0.5rem; border-radius: 4px;
    font-size: 0.75rem; font-weight: 700; letter-spacing: 0.04em;
  }}
  .tag.reject {{ background: rgba(248,113,113,0.15); color: var(--bad); }}
  .verdict {{ display: flex; gap: 0.6rem; align-items: center; margin: 0.4rem 0; }}
  .msg {{ color: var(--muted); font-style: italic; margin: 0.3rem 0 0; }}
  footer {{ text-align: center; color: var(--muted); font-size: 0.8rem; padding: 2rem; }}
</style>
</head>
<body>
<header>
  <h1>dlc &mdash; a magic-set Datalog compiler for stratified negation</h1>
  <p>Static presentation artifact, generated by <code>harness/build_presentation.py</code>
  from committed measurement JSON and <code>dlc explain</code> output only &mdash;
  every number here also appears in <code>docs/reports/FINAL.md</code>, with its
  own citation. No server, no framework, no state.</p>
</header>
<nav>
  <a href="#rejection">1. Rejecting programs</a>
  <a href="#metric">2. Three-column metric</a>
  <a href="#guard">3. Guard firing &amp; cone</a>
  <a href="#findings">4. Findings with mechanisms</a>
</nav>
<main>

<section id="rejection">
  <h2>1. The analyzer rejecting programs</h2>
  <p class="lead">All four rejection grounds, each on a minimal sample program,
  <code>dlc explain</code>'s own output (unmodified) below each one.</p>
  {rejection}
</section>

<section id="metric">
  <h2>2. Three-column metric, per shape</h2>
  <p class="lead"><code>T_none</code> / <code>T_souffle</code> / <code>T_dlc</code>
  (incl-sup, the project's headline convention), contribution =
  <code>T_souffle/T_dlc</code>. Full 32-point sweep,
  <code>measurements/m3-5-headline-m4/summary.json</code>.</p>
  {metric}
</section>

<section id="guard">
  <h2>3. The guard firing and declining</h2>
  <p class="lead">Per-program culprit SCC, cone, and declined fraction
  (<code>bin/conecheck</code>, cross-checked exactly against an independent
  Python implementation of the same cone computation).</p>
  {guard}
</section>

<section id="findings">
  <h2>4. Two findings with mechanisms</h2>
  <p class="lead">The two results <code>docs/reports/FINAL.md</code> section 6
  names as carrying the presentation.</p>
  {findings}
</section>

</main>
<footer>Generated from committed material only &mdash; see <code>docs/reports/FINAL.md</code> for the full report.</footer>
</body>
</html>
"""


def main():
    html_out = PAGE_TEMPLATE.format(
        rejection=build_rejection_section(),
        metric=build_metric_section(),
        guard=build_guard_section(),
        findings=build_findings_section(),
    )
    OUT_PATH.write_text(html_out)
    print(f"wrote {OUT_PATH} ({len(html_out):,} bytes)")


if __name__ == "__main__":
    main()
