'use strict';
const $ = id => document.getElementById(id);
const el = (tag, text, cls) => {
  const n = document.createElement(tag);
  if (text !== undefined)
    n.textContent = text;
  if (cls)
    n.className = cls;
  return n;
};
let examples = [], tools = {}, current = null, revision = 0, sequence = 0, controller = null;
function table(head, rows) {
  const t = el('table'), thead = el('thead'), tr = el('tr');
  for (const h of head)
    tr.append(el('th', h));
  thead.append(tr);
  t.append(thead);
  const b = el('tbody');
  for (const cells of rows) {
    const row = el('tr');
    for (const c of cells) {
      const td = el('td');
      if (c instanceof Node)
        td.append(c);
      else
        td.textContent = String(c ?? '');
      row.append(td);
    }
    b.append(row);
  }
  t.append(b);
  return t;
}
function badge(text, kind = '') {
  return el('span', text, 'badge ' + kind);
}
function status(text, kind = '') {
  $('status').textContent = text;
  $('status').className = kind;
}
function facts() {
  return Object.fromEntries([...document.querySelectorAll('#facts textarea')].map(t => [
    t.dataset.name,
    t.value
  ]));
}
function invalidate() {
  revision++;
  sequence++;
  if (controller)
    controller.abort();
  current = null;
  $('results').replaceChildren(el('p', 'Inputs changed \u2014 previous results cleared. Run these inputs to compare answers.', 'muted'));
  $('external').replaceChildren();
  $('c-source').textContent = '';
  $('snapshot').textContent = 'Inputs changed; no current execution snapshot.';
  for (const id of [
      'compile-status',
      'binding-flow',
      'queries',
      'dependencies',
      'guard',
      'decisions',
      'tokens',
      'ast-tree',
      'ast-node',
      'generated-rules',
      'rewrite',
      'selected-rule',
      'compaction'
    ])
    $(id).replaceChildren();
  $('compile-status').textContent = 'Inspection cleared \u2014 compile the current source.';
  $('rule').replaceChildren();
  $('compact-details').hidden = true;
  updateActions();
  status('Inputs changed \u2014 results and inspection invalidated.', 'stale');
}
function updateActions() {
  $('native').disabled = !current || current.status !== 'ok' || !tools.cc?.available;
  $('oracle').disabled = !current || current.status !== 'ok' || !tools.oracle?.available;
}
function factEditors(names, values) {
  $('facts').replaceChildren();
  for (const name of names) {
    const box = el('div', undefined, 'fact-editor');
    const label = el('label', name + '.facts');
    const area = el('textarea');
    area.dataset.name = name;
    area.setAttribute('aria-label', name + '.facts');
    area.spellcheck = false;
    area.value = values[name] ?? '';
    area.addEventListener('input', () => {
      invalidate();
      drawInputGraph();
    });
    label.append(area);
    box.append(label);
    $('facts').append(box);
  }
  drawInputGraph();
}
async function post(path, body, signal) {
  const response = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    signal
  });
  if (!response.ok)
    throw Error(await response.text());
  return response.json();
}
async function perform(inspectOnly) {
  if (controller)
    controller.abort();
  controller = new AbortController();
  const signal = controller.signal, seq = ++sequence, rev = revision, source = $('source').value, mode = $('mode').value, existing = facts();
  current = null;
  updateActions();
  $('external').replaceChildren();
  $('c-source').textContent = '';
  $('results').replaceChildren(el('p', inspectOnly ? 'Inspecting current source\u2026' : 'Running original and selected mode on the same snapshot\u2026', 'muted'));
  status(inspectOnly ? 'Inspecting compiler data\u2026' : 'Compiling and running actual inputs\u2026', 'busy');
  try {
    const discovery = await post('/api/run', {
      source,
      mode,
      facts: {},
      inspect_only: true
    }, signal);
    if (seq !== sequence || rev !== revision)
      return;
    if (!discovery.inspection || discovery.inspection.status !== 'ok') {
      apply(discovery);
      return;
    }
    const inputs = discovery.inspection.source.relations.filter(r => r.input).map(r => r.name);
    factEditors(inputs, existing);
    const result = await post('/api/run', {
      source,
      mode,
      facts: facts(),
      inspect_only: inspectOnly
    }, signal);
    if (seq !== sequence || rev !== revision)
      return;
    apply(result);
  } catch (e) {
    if (e.name === 'AbortError')
      return;
    if (seq !== sequence)
      return;
    status(e.message, 'error');
    $('results').replaceChildren(el('p', 'Request failed. No verified result.', 'diagnostic'));
  }
}
function apply(result) {
  current = result;
  const r = result.inspection;
  if (r)
    renderInspection(r);
  $('snapshot').dataset.runId = result.run_id;
  $('snapshot').dataset.sourceHash = result.source_sha256;
  $('snapshot').dataset.inputHash = result.input_sha256;
  $('snapshot').textContent = `Run ${ result.run_id }\nMode ${ result.mode } · ${ result.engine }\nSource SHA-256 ${ result.source_sha256 }\nInput snapshot SHA-256 ${ result.input_sha256 }\n${ (result.inputs || []).map(i => i.name + ': ' + i.bytes + ' bytes \xB7 ' + i.sha256).join('\n') }`;
  renderCommands(result.commands || []);
  renderResults(result);
  updateActions();
  status(result.status === 'ok' ? 'Completed real executions on this input snapshot.' : result.status === 'inspected' ? 'Inspection ready. Execute to measure work and compare answers.' : `${ result.status }: ${ result.message || 'See compiler diagnostics and command outcomes.' }`, [
    'ok',
    'inspected'
  ].includes(result.status) ? '' : 'error');
}
function renderInspection(r) {
  $('compile-status').replaceChildren(badge('source: ' + (r.source?.status || r.status), r.status === 'ok' ? '' : 'error'));
  for (const d of r.diagnostics || []) {
    const pos = d.span ? `${ d.span.start.line }:${ d.span.start.col }` : '';
    $('compile-status').append(el('p', `${ d.stage } / ${ d.category } ${ pos }: ${ d.message }`, 'diagnostic'));
  }
  if (!r.ast)
    return;
  $('rule').replaceChildren();
  for (const c of r.ast.clauses || []) {
    const o = el('option', `${ c.id } · ${ c.head.name } · line ${ c.span?.start.line ?? '?' }`);
    o.value = c.id;
    $('rule').append(o);
  }
  const recursive = (r.source?.dependencies || []).find(e => e.from === e.to);
  if (recursive)
    $('rule').value = recursive.rule_id;
  selectRule();
  $('queries').replaceChildren(el('h3', 'Actual query demand'));
  const q = r.transform?.queries || [];
  if (q.length)
    $('queries').append(table([
      'source predicate',
      'bindings',
      'seed arguments',
      'view / magic'
    ], q.map(x => [
      x.source_predicate,
      x.bindings.join(''),
      x.demand_arguments.map(a => `${ a.position }: ${ a.expression } (${ a.type })`).join(', '),
      `${ x.view_id }\n${ x.magic_id }`
    ])));
  else
    $('queries').append(el('p', r.transform?.no_query_reason || 'Unavailable after source errors.', 'muted'));
  $('guard').replaceChildren();
  for (const [label, p] of [
      [
        'source',
        r.source
      ],
      [
        'candidate',
        r.candidate
      ],
      [
        'final',
        r.final
      ]
    ])
    $('guard').append(badge(`${ label }: ${ p?.status || 'not attempted' }`, p?.status === 'unstratifiable' ? 'neutral' : ''));
  if (r.candidate?.invalid_sccs?.length)
    $('guard').append(el('pre', 'Candidate negative-cycle SCCs:\n' + r.candidate.invalid_sccs.map(s => s.join(', ')).join('\n')));
  $('decisions').replaceChildren(table([
    'source',
    'specialized',
    'full extent',
    'decision / reasons'
  ], (r.transform?.decisions || []).map(d => [
    d.source_predicate,
    d.specialized ? d.views.join(', ') : 'no',
    d.full_extent_retained ? 'retained' : 'not required',
    `${ d.fallback ? 'FALLBACK: ' + d.fallback_reason : 'original / demand views' }\n${ d.retained_reasons.join('; ') }`
  ])));
  $('rewrite').textContent = r.final?.printed || 'Unavailable';
  $('tokens').replaceChildren(table([
    'token',
    'text',
    'line:column'
  ], r.tokens.map(t => [
    t.kind,
    JSON.stringify(t.text),
    `${ t.span.start.line }:${ t.span.start.col }`
  ])));
  const comp = r.transform?.compaction || [];
  $('compact-details').hidden = !comp.length;
  $('compaction').replaceChildren(el('p', comp.some(a => a.action === 'remove') ? `${ comp.filter(a => a.action === 'remove').length } checkpoint relations removed.` : 'No checkpoints removed; this pass was a no-op.', 'muted'), table([
    'relation',
    'action',
    'reason'
  ], comp.map(a => [
    a.relation,
    a.action,
    a.reason
  ])));
  renderDependencies();
}
function selectRule() {
  const r = current?.inspection;
  if (!r?.ast)
    return;
  const c = r.ast.clauses.find(c => c.id === $('rule').value);
  if (!c)
    return;
  const text = sourceSlice(c.span);
  $('selected-rule').replaceChildren(el('div', text, 'rule-text'));
  $('binding-flow').replaceChildren();
  const sourcePlan = r.execution_plans?.find(p => p.program === 'source' && p.rule_id === c.id);
  if (sourcePlan) {
    $('binding-flow').append(el('h3', sourcePlan.lowered ? 'Executable plan \xB7 arithmetic columns lowered to equality filters' : 'Executable plan \xB7 actual scheduled order'));
    $('binding-flow').append(table([
      'step',
      'literal / filter',
      'bound before',
      'bound after'
    ], sourcePlan.steps.map((s, i) => [
      i + 1,
      `${ s.kind }\n${ nodeText(s.literal) }`,
      s.bound_before.join(', ') || '\u2205',
      s.bound_after.join(', ') || '\u2205'
    ])));
  }
  for (const s of r.transform?.sips || []) {
    if (s.source_rule_id !== c.id)
      continue;
    $('binding-flow').append(el('h3', `Demand SIPS · ${ s.view_id } (${ s.bindings.join('') })`));
    $('binding-flow').append(table([
      'literal',
      'before',
      'after'
    ], s.steps.map(x => [
      `${ x.kind }: ${ x.text }`,
      x.bound_before.join(', ') || '\u2205',
      x.bound_after.join(', ') || '\u2205'
    ])));
  }
  $('ast-tree').replaceChildren(astTree(c));
  $('ast-node').textContent = JSON.stringify(c, null, 2);
  const rules = (r.transform?.rules || []).filter(x => x.source_rule_id === c.id);
  $('generated-rules').replaceChildren();
  for (const rule of rules) {
    $('generated-rules').append(el('h3', `${ rule.role } · ${ rule.id }`));
    $('generated-rules').append(el('pre', rule.text || JSON.stringify(r.final?.ast.clauses[rule.index], null, 2)));
  }
  highlight(c.span);
}
function sourceSlice(span) {
  if (!span)
    return '';
  const b = new TextEncoder().encode($('source').value);
  return new TextDecoder().decode(b.slice(span.start.offset, span.end.offset));
}
function highlight(span) {
  if (!span)
    return;
  const b = new TextEncoder().encode($('source').value), d = new TextDecoder();
  $('source').setSelectionRange(d.decode(b.slice(0, span.start.offset)).length, d.decode(b.slice(0, span.end.offset)).length);
}
function astTree(n) {
  const d = el('details');
  d.open = true;
  const s = el('summary');
  const b = el('button', `${ n.kind } ${ n.name || n.op || n.value || '' }`);
  b.dataset.nodeId = n.id;
  b.addEventListener('click', e => {
    e.preventDefault();
    $('ast-node').textContent = JSON.stringify(n, null, 2);
    highlight(n.span);
  });
  s.append(b);
  d.append(s);
  for (const key of [
      'head',
      'body',
      'atom',
      'terms',
      'left',
      'right',
      'x'
    ]) {
    const value = n[key];
    if (Array.isArray(value))
      for (const x of value)
        d.append(astTree(x));
    else if (value && typeof value === 'object')
      d.append(astTree(value));
  }
  return d;
}
// This prints the supplied tagged node for display only. It is not a parser,
// planner or evaluator; the authoritative rewrite is compiler-printed source.
function nodeText(n) {
  switch (n.kind) {
  case 'Atom':
    return `${ n.name }(${ (n.terms || []).map(nodeText).join(', ') })`;
  case 'NegatedAtom':
    return '!' + nodeText(n.atom);
  case 'Constraint':
    return `${ nodeText(n.left) } ${ n.op } ${ nodeText(n.right) }`;
  case 'Var':
    return n.name;
  case 'NumberLit':
    return n.value;
  case 'StringLit':
    return JSON.stringify(n.value);
  case 'UnaryExpr':
    return n.op + nodeText(n.x);
  case 'BinaryExpr':
    return `(${ nodeText(n.left) }${ n.op }${ nodeText(n.right) })`;
  case 'Wildcard':
    return '_';
  default:
    return n.kind;
  }
}
const ns = 'http://www.w3.org/2000/svg';
function svgEl(tag, attrs, text) {
  const n = document.createElementNS(ns, tag);
  for (const [k, v] of Object.entries(attrs || {}))
    n.setAttribute(k, v);
  if (text !== undefined)
    n.textContent = text;
  return n;
}
function graph(nodes, edges, container, prefix) {
  container.replaceChildren();
  if (!nodes.length) {
    container.append(el('p', 'No graph for this stage.', 'muted'));
    return;
  }
  const width = Math.max(390, Math.ceil(Math.sqrt(nodes.length)) * 140), height = Math.max(260, Math.ceil(nodes.length / 3) * 85), svg = svgEl('svg', {
      viewBox: `0 0 ${ width } ${ height }`,
      role: 'img',
      'aria-label': prefix + ' graph'
    }), defs = svgEl('defs'), marker = svgEl('marker', {
      id: prefix + '-arrow',
      viewBox: '0 0 10 10',
      refX: 9,
      refY: 5,
      markerWidth: 6,
      markerHeight: 6,
      orient: 'auto-start-reverse'
    });
  marker.append(svgEl('path', {
    d: 'M 0 0 L 10 5 L 0 10 z',
    fill: '#6c8f79'
  }));
  defs.append(marker);
  svg.append(defs);
  const positions = new Map();
  nodes.forEach((n, i) => {
    const angle = 2 * Math.PI * i / nodes.length - Math.PI / 2;
    positions.set(n.id, {
      x: width / 2 + (width / 2 - 75) * Math.cos(angle),
      y: height / 2 + (height / 2 - 45) * Math.sin(angle)
    });
  });
  edges.forEach(e => {
    const a = positions.get(e.from), b = positions.get(e.to);
    if (!a || !b)
      return;
    let path;
    if (e.from === e.to)
      path = `M ${ a.x + 30 } ${ a.y - 10 } C ${ a.x + 95 } ${ a.y - 85 },${ a.x - 90 } ${ a.y - 85 },${ a.x - 30 } ${ a.y - 10 }`;
    else {
      const dx = b.x - a.x, dy = b.y - a.y, len = Math.hypot(dx, dy) || 1;
      path = `M ${ a.x + dx / len * 42 } ${ a.y + dy / len * 17 } L ${ b.x - dx / len * 46 } ${ b.y - dy / len * 20 }`;
    }
    const line = svgEl('path', {
      d: path,
      class: 'graph-edge' + (e.negative ? ' negative' : ''),
      'marker-end': `url(#${ prefix }-arrow)`,
      'data-from': e.from,
      'data-to': e.to,
      'data-negative': String(!!e.negative)
    });
    const hit = svgEl('path', {
      d: path,
      class: 'graph-hit',
      tabindex: 0,
      role: 'button',
      'aria-label': `${ e.from } reads ${ e.negative ? 'not ' : '' }${ e.to }`,
      'data-from': e.from,
      'data-to': e.to
    });
    if (e.rule_id) {
      const choose = () => {
        $('rule').value = e.rule_id;
        selectRule();
      };
      hit.addEventListener('click', choose);
      hit.addEventListener('keydown', ev => {
        if (ev.key === 'Enter' || ev.key === ' ') {
          ev.preventDefault();
          choose();
        }
      });
    }
    svg.append(hit);
    line.append(svgEl('title', {}, `${ e.from } → ${ e.negative ? '!' : '' }${ e.to }${ e.rule_id ? ' \xB7 ' + e.rule_id : '' }`));
    if (e.rule_id)
      line.addEventListener('click', () => {
        $('rule').value = e.rule_id;
        selectRule();
      });
    svg.append(line);
    if (e.negative)
      svg.append(svgEl('text', {
        x: (a.x + b.x) / 2 + 5,
        y: (a.y + b.y) / 2 - 5,
        class: 'graph-edge-label'
      }, '!'));
  });
  nodes.forEach(n => {
    const p = positions.get(n.id), g = svgEl('g', {
        class: 'graph-node',
        'data-relation': n.id
      });
    g.append(svgEl('rect', {
      x: p.x - 52,
      y: p.y - 19,
      width: 104,
      height: 39,
      rx: 4
    }));
    g.append(svgEl('text', {
      x: p.x,
      y: p.y - 3,
      'text-anchor': 'middle'
    }, n.id));
    g.append(svgEl('text', {
      x: p.x,
      y: p.y + 11,
      'text-anchor': 'middle',
      class: 'role'
    }, n.role));
    svg.append(g);
  });
  container.append(svg);
}
function renderDependencies() {
  const stage = $('graph-stage').value, p = current?.inspection?.[stage];
  if (!p) {
    $('dependencies').replaceChildren(el('p', 'This stage is unavailable.', 'muted'));
    return;
  }
  const roles = new Map((p.origins?.relations || []).map(r => [
    r.id,
    r.role
  ]));
  graph(p.relations.map(r => ({
    id: r.name,
    role: roles.get(r.name) || [
      r.input ? 'input' : '',
      r.output ? 'output' : '',
      r.scc !== undefined ? 'SCC ' + r.scc : 'source'
    ].filter(Boolean).join(' \xB7 ')
  })), p.dependencies.map(e => ({
    ...e,
    rule_id: e.source_rule_id || (stage === 'source' ? e.rule_id : undefined)
  })), $('dependencies'), 'dependencies');
  $('dependencies').append(table([
    'head',
    'reads',
    'polarity',
    'source rule'
  ], p.dependencies.map(e => [
    e.from,
    e.to,
    e.negative ? 'negative (!)' : 'positive',
    e.source_rule_id || (stage === 'source' ? e.rule_id : 'unavailable')
  ])));
}
function drawInputGraph() {
  const f = facts();
  const usable = Object.hasOwn(f, 'edge') && Object.hasOwn(f, 'node');
  $('input-graph-details').hidden = !usable;
  if (!usable)
    return;
  const rows = v => v.endsWith('\n') ? v.slice(0, -1).split('\n') : v ? v.split('\n') : [];
  const nodes = rows(f.node).filter(x => /^-?\d+$/.test(x)), edges = rows(f.edge).map(x => x.split('\t'));
  if (edges.some(x => x.length !== 2 || x.some(v => !nodes.includes(v)))) {
    $('input-graph').replaceChildren(el('p', 'Facts do not form this graph view. The raw facts remain authoritative.', 'muted'));
    return;
  }
  graph(nodes.map(n => ({
    id: n,
    role: 'input node'
  })), edges.map(([from, to]) => ({
    from,
    to
  })), $('input-graph'), 'input');
  for (const id of [
      'edge-from',
      'edge-to'
    ]) {
    const old = $(id).value;
    $(id).replaceChildren(...nodes.map(n => {
      const o = el('option', n);
      o.value = n;
      return o;
    }));
    if (nodes.includes(old))
      $(id).value = old;
  }
  $('edge-list').replaceChildren(...edges.map(([a, b], i) => {
    const button = el('button', `${ a } → ${ b } ×`);
    button.title = 'Remove this edge';
    button.addEventListener('click', () => {
      const area = document.querySelector('#facts textarea[data-name="edge"]');
      area.value = edges.filter((_, j) => i !== j).map(e => e.join('\t')).join('\n') + '\n';
      invalidate();
      drawInputGraph();
    });
    return button;
  }));
}
function renderOutputs(outputs) {
  const box = el('div');
  for (const f of outputs) {
    const d = el('details');
    d.open = true;
    d.append(el('summary', `${ f.name } · ${ f.rows.length } distinct serialized rows`));
    if (!f.rows.length)
      d.append(el('p', 'Empty file \xB7 no rows', 'muted'));
    else
      d.append(table([
        'escaped row bytes',
        'declared columns'
      ], f.rows.map(r => [
        r.escaped,
        r.ambiguous ? 'Ambiguous TSV: inspect escaped bytes' : r.columns?.map(x => JSON.stringify(x)).join(' | ') || 'nullary true'
      ])));
    box.append(d);
  }
  return box;
}
function renderComparison(c) {
  if (!c)
    return el('p', 'Comparison unavailable.', 'muted');
  const box = el('div');
  box.append(badge(c.available ? c.equal ? 'All requested output sets equal' : 'OUTPUT MISMATCH' : c.reason, c.available && c.equal ? '' : c.available ? 'error' : 'neutral'));
  for (const d of c.differences || [])
    box.append(el('pre', JSON.stringify(d, null, 2)));
  return box;
}
function renderResults(result) {
  $('results').replaceChildren();
  if (result.status === 'inspected') {
    $('results').append(el('p', 'Inspection complete; execution has not been run.', 'muted'));
    return;
  }
  if (result.comparison)
    $('results').append(renderComparison(result.comparison));
  if (result.evaluations?.length) {
    $('results').append(table([
      'mode',
      'Go evaluator',
      'derived materializations',
      'candidate-head attempts'
    ], result.evaluations.map(e => [
      e.mode,
      e.engine,
      e.materialized,
      e.attempts
    ])));
    const selected = result.evaluations.find(e => e.mode === result.mode && e.engine === 'semi-naive');
    if (selected)
      $('results').append(renderOutputs(selected.outputs));
    const d = el('details');
    d.append(el('summary', 'Per-relation counters \xB7 all modes'));
    for (const e of result.evaluations) {
      d.append(el('h3', e.mode + ' / ' + e.engine));
      d.append(table([
        'relation',
        'derived',
        'seed',
        'local buckets'
      ], e.relations.map(r => [
        r.relation,
        r.derived,
        r.seed,
        JSON.stringify(r.scc_local_buckets || {})
      ])));
    }
    $('results').append(d);
  }
  if (![
      'ok',
      'inspected'
    ].includes(result.status))
    $('results').prepend(el('p', `${ result.status }: ${ result.message || 'See diagnostics.' }`, 'diagnostic'));
}
function renderCommands(commands) {
  $('commands').replaceChildren();
  for (const c of commands) {
    $('commands').append(el('h3', `${ c.stage }: ${ c.status } (exit ${ c.exit_code })`));
    $('commands').append(el('pre', c.argv.join(' ') + '\n' + (c.capture_truncated ? '[capture truncated]\n' : '') + (c.stdout || '') + (c.stderr || '') + `\nprocess-group cleanup: ${ c.process_group_reaped ?? c.process_group_cleanup ?? false }`));
  }
}
async function external(kind) {
  if (!current || current.status !== 'ok')
    return;
  const run = current, rev = revision, seq = ++sequence;
  if (controller)
    controller.abort();
  controller = new AbortController();
  $('external').replaceChildren(el('p', kind === 'native' ? 'Generating, compiling and executing native C\u2026' : 'Running optional Soufflé answer check\u2026', 'muted'));
  try {
    const result = await post(`/api/${ kind }/${ encodeURIComponent(run.run_id) }`, {}, controller.signal);
    if (seq !== sequence || rev !== revision || current?.run_id !== run.run_id)
      return;
    $('external').replaceChildren(el('h3', `${ kind }: ${ result.status }`));
    $('external').append(renderComparison(result.comparison));
    if (result.message)
      $('external').append(el('p', result.message, 'diagnostic'));
    if (result.outputs?.length)
      $('external').append(renderOutputs(result.outputs));
    if (result.generated_c)
      $('c-source').textContent = result.generated_c;
    renderCommands([
      ...run.commands || [],
      ...result.commands
    ]);
  } catch (e) {
    if (e.name !== 'AbortError' && seq === sequence)
      $('external').replaceChildren(el('p', e.message, 'diagnostic'));
  }
}
async function loadExample() {
  const e = examples.find(x => x.id === $('example').value);
  if (!e)
    return;
  invalidate();
  $('source').value = e.source;
  $('example-description').textContent = e.description;
  factEditors(Object.keys(e.facts), e.facts);
  await perform(true);
}
$('source').addEventListener('input', invalidate);
$('mode').addEventListener('change', invalidate);
$('rule').addEventListener('change', selectRule);
$('graph-stage').addEventListener('change', renderDependencies);
$('load-example').addEventListener('click', loadExample);
$('inspect').addEventListener('click', () => perform(true));
$('run').addEventListener('click', () => perform(false));
$('native').addEventListener('click', () => external('native'));
$('oracle').addEventListener('click', () => external('oracle'));
$('add-edge').addEventListener('click', () => {
  const area = document.querySelector('#facts textarea[data-name="edge"]');
  if (!area)
    return;
  area.value += (area.value && !area.value.endsWith('\n') ? '\n' : '') + $('edge-from').value + '\t' + $('edge-to').value + '\n';
  invalidate();
  drawInputGraph();
});
(async () => {
  try {
    const [pre, ex] = await Promise.all([
      fetch('/api/preflight').then(r => r.json()),
      fetch('/api/examples').then(r => r.json())
    ]);
    tools = pre.tools;
    examples = ex;
    $('tools').replaceChildren(...Object.entries(tools).map(([n, t]) => el('span', `${ n }: ${ t.available ? 'available' : 'unavailable' }`, t.available ? '' : 'tool-missing')), el('span', `timeout: ${ pre.timeout_ms } ms / command`));
    $('mode').replaceChildren(...pre.modes.map(m => {
      const o = el('option', m);
      o.value = m;
      return o;
    }));
    $('run').disabled = !tools.compiler.available;
    $('inspect').disabled = !tools.compiler.available;
    if (!tools.cc.available)
      $('native').title = 'C compiler unavailable';
    if (!tools.oracle.available) {
      $('oracle').textContent = 'Soufflé unavailable';
      $('oracle').title = 'Optional oracle is missing; native C still works.';
    }
    $('example').replaceChildren(...examples.map(e => {
      const o = el('option', e.title);
      o.value = e.id;
      return o;
    }));
    await loadExample();
  } catch (e) {
    status(e.message, 'error');
  }
})();
