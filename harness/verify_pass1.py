#!/usr/bin/env python3
"""Portable source/backend/oracle verification; no /root/compiler paths.

Run `make verify` first, then `make oracle` (DLC_SOUFFLE may name a binary).
Profiles are collected on the SAME Souffle engine for all three variants.
This script checks outputs and tuple materialization, never timing claims.
"""
from pathlib import Path
import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile

from tuple_report import analyze
from answers import read_outputs, differences
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--require-oracle', action='store_true')
    parser.add_argument('--output', type=Path, default=None)
    args = parser.parse_args()
    oracle = shutil.which(os.environ.get('DLC_SOUFFLE', 'souffle'))
    if not oracle:
        print('Souffle is required for oracle verification; set DLC_SOUFFLE.', file=sys.stderr)
        return 1 if args.require_oracle else 0
    cc = shutil.which(os.environ.get('CC', 'cc'))
    if not cc:
        print('A C compiler is required.', file=sys.stderr)
        return 1
    dlc = ROOT / 'bin/dlc'
    if not dlc.is_file():
        print('Build the compiler first: make build', file=sys.stderr)
        return 1
    dest = (args.output or ROOT / ("measurements/oracle-" + datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S.%fZ"))).resolve()
    dest.mkdir(parents=True, exist_ok=False)
    commands = []
    summary = []

    def call(cmd, *, timeout=60):
        cmd = [str(x) for x in cmd]
        run = subprocess.run(cmd, capture_output=True, text=True, timeout=timeout)
        commands.append({'argv': cmd, 'returncode': run.returncode, 'stdout': run.stdout, 'stderr': run.stderr})
        (dest / 'commands.json').write_text(json.dumps(commands, indent=2) + '\n')
        if run.returncode:
            raise RuntimeError(f'command failed ({run.returncode}): {cmd}\n{run.stdout}\n{run.stderr}')
        return run.stdout

    def compare(actual, expected, label):
        if actual != expected:
            raise RuntimeError(f'{label}: answer mismatch: {json.dumps(differences(actual,expected))}')

    def io_names(source):
        report = json.loads(call([dlc, 'inspect', source]))
        relations = report['source']['relations']
        return ([r['name']+'.csv' for r in relations if r['output']],
                [r['name'] for r in relations if r['input']])

    def check(source, facts, label):
        work = dest / label
        work.mkdir(exist_ok=True)
        row = {'case': label, 'source': str(source.relative_to(ROOT)), 'facts': str(facts.relative_to(ROOT))}
        expected, _ = io_names(source)
        def answers(directory): return read_outputs(directory, expected)
        modes = {}
        for mode, path, extra in [('none', source, []), ('souffle_magic', source, ['--magic-transform=*'])]:
            out = work / mode
            out.mkdir(exist_ok=True)
            call([oracle, '-F', facts, '-D', out, '-p', out/'profile.json', *extra, path])
            modes[mode] = answers(out)
            row['T_'+mode] = analyze(out/'profile.json')['T_excl_copy']
        compare(modes['souffle_magic'], modes['none'], label+'/souffle_magic')
        emitted = json.loads(call([dlc, 'emit', source, '--transformer=guarded']))
        transformed = work/'transformed.dl'
        transformed.write_text(emitted['printed'])
        out = work/'guarded_souffle'
        out.mkdir(exist_ok=True)
        call([oracle, '-F', facts, '-D', out, '-p', out/'profile.json', transformed])
        compare(answers(out), modes['none'], label+'/guarded_souffle')
        row['T_guarded'] = analyze(out/'profile.json')['T_excl_copy']
        row['original_over_guarded'] = row['T_none']/row['T_guarded'] if row['T_guarded'] else None
        for sub, flag, mode in [('run', 'passthrough', 'naive'), ('run-seminaive', 'passthrough', 'semi'), ('run-seminaive', 'guarded', 'guarded_go')]:
            out = work/mode
            call([dlc, sub, source, facts, out, '--transformer='+flag])
            compare(answers(out), modes['none'], label+'/'+mode)
        with tempfile.TemporaryDirectory(prefix='dlc-c-') as td:
            td = Path(td)
            cfile, binary, out = td/'program.c', td/'program', work/'guarded_c'
            call([dlc, 'codegen', source, cfile, '--transformer=guarded'])
            call([cc, '-std=c11', '-O2', cfile, '-o', binary])
            call([binary, facts, out])
            compare(answers(out), modes['none'], label+'/guarded_c')
        row['answers_identical'] = True
        summary.append(row)
        print(json.dumps(row), flush=True)
        (dest/'summary.json').write_text(json.dumps(summary, indent=2)+'\n')

    try:
        call([oracle, '--version'])
        call([cc, '--version'])
        cases = [
            ('reachability_complement', ROOT/'fixtures/p2-scale-250', 'reachability_n250'),
            ('ancestor_nonancestor', ROOT/'fixtures/benchmark-family/ancestor_nonancestor/n500', 'ancestor_n500'),
            ('same_generation_negation', ROOT/'fixtures/benchmark-family/same_generation_negation/d4_b4', 'same_generation_d4'),
            ('transitive_closure_bound', ROOT/'fixtures/benchmark-family/transitive_closure_bound/n500', 'positive_n500'),
            ('culprit_cycle', ROOT/'fixtures/benchmark-family/culprit_cycle/n20', 'culprit_n20'),
        ]
        for name, facts, label in cases:
            check(ROOT/f'tests/corpus/BENCHMARK_FAMILY/{name}.dl', facts, label)
        for n in (20, 50, 100):
            check(ROOT/'tests/corpus/CONE_CORPUS/cc_growing_sibling.dl', ROOT/f'fixtures/cone_corpus/cc_growing_sibling/n{n}', f'growing_n{n}')
        # Every supported hostile fixture must round-trip and match the oracle.
        # Unsupported/malformed sources are counted separately, never as passes.
        hostile = {'accepted': [], 'rejected_by_subset': [], 'oracle_rejected': [], 'runtime_limited': []}
        for source in sorted((ROOT/'tests/hostile').glob('*.dl')):
            run = subprocess.run([str(dlc), 'check', str(source)], capture_output=True, text=True)
            if run.returncode:
                hostile['rejected_by_subset'].append(source.name)
                continue
            with tempfile.TemporaryDirectory(prefix='dlc-hostile-') as td:
                td=Path(td)
                expected, input_names = io_names(source)
                names = [name[:-4] for name in expected] + input_names
                if any(len(name.encode())+6>os.pathconf(td,'PC_NAME_MAX') for name in names):
                    hostile['runtime_limited'].append({'file':source.name,'reason':'relation name exceeds filesystem filename limit; parser-only fixture'})
                    continue
                for name in input_names:
                    (td/(name+'.facts')).write_text('')
                out=td/'oracle';out.mkdir()
                run=subprocess.run([oracle,'-F',str(td),'-D',str(out),str(source)],capture_output=True,text=True,timeout=15)
                if run.returncode:
                    # These must be explained, not hidden in a success count.
                    hostile['oracle_rejected'].append({'file':source.name,'stderr':run.stderr})
                    continue
                got=td/'dlc'
                call([dlc,'run-seminaive',source,td,got,'--transformer=guarded'])
                compare(read_outputs(got,expected),read_outputs(out,expected),'hostile/'+source.name)
                hostile['accepted'].append(source.name)
        (dest/'hostile.json').write_text(json.dumps(hostile,indent=2)+'\n')
        if hostile['oracle_rejected']:
            raise RuntimeError('dlc accepted sources rejected by Souffle; inspect hostile.json')
        print(f"PASS: {len(summary)} benchmark configurations; {len(hostile['accepted'])} supported hostile sources; {len(hostile['rejected_by_subset'])} explicitly rejected sources.")
        return 0
    except (RuntimeError, subprocess.TimeoutExpired) as exc:
        print(str(exc), file=sys.stderr)
        return 1


if __name__ == '__main__':
    raise SystemExit(main())
