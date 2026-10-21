#!/usr/bin/env python3
"""Verify production guarded-compact with exact outputs and measured work.

No tool install/build, repository mutation, old evidence reuse, or timing claim.
"""
from pathlib import Path
import argparse, hashlib, json, os, random, shutil, subprocess, sys, tempfile
from datetime import datetime, timezone

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--repo',type=Path,default=Path.cwd())
p.add_argument('--output',type=Path)
p.add_argument('--require-oracle',action='store_true')
a=p.parse_args();root=a.repo.resolve();dest=(a.output or root/("measurements/compact-"+datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S.%fZ"))).resolve()
if dest.exists():raise SystemExit('Choose a fresh --output directory; evidence is never overwritten.')
binary=root/'bin/dlc';oracle=shutil.which(os.environ.get('DLC_SOUFFLE','souffle'));cc=shutil.which(os.environ.get('CC','cc'))
if not binary.is_file() or not cc:raise SystemExit('Build dlc and provide a C compiler first.')
if a.require_oracle and not oracle:raise SystemExit('Soufflé required; set DLC_SOUFFLE or install it.')
dest.mkdir(parents=True);sys.path.insert(0,str(root/'harness'))
from tuple_report import analyze
from answers import read_outputs
commands=[]
def run(args,timeout=60):
    args=list(map(str,args));r=subprocess.run(args,capture_output=True,text=True,timeout=timeout)
    commands.append(dict(argv=args,returncode=r.returncode,stdout=r.stdout,stderr=r.stderr))
    (dest/'commands.json').write_text(json.dumps(commands,indent=2)+'\n')
    if r.returncode:raise RuntimeError(f'command failed: {args}\n{r.stdout}\n{r.stderr}')
    return r.stdout
def answers(directory,expected):
    return read_outputs(directory,expected)
def equal(actual,want,label):
    if actual!=want:
        delta={}
        for n in actual.keys()|want.keys():
            got,expected=actual.get(n,set()),want.get(n,set())
            if n not in actual or n not in want or got!=expected:
                delta[n]={'extra':[repr(x) for x in sorted(got-expected)],'missing':[repr(x) for x in sorted(expected-got)]}
        raise RuntimeError(label+': '+json.dumps(delta))
def compact(source,guarded,target):
    target.write_text(json.loads(run([binary,'emit',source,'--transformer=guarded-compact']))['printed'])
def emit(source,mode,target):
    target.write_text(json.loads(run([binary,'emit',source,'--transformer='+mode]))['printed'])
def native(source,facts,out,directory,mode="passthrough"):
    # Native build products are reproducible scratch, not measurement evidence.
    with tempfile.TemporaryDirectory(prefix='dlc-compact-native-') as temp:
        temp=Path(temp);cfile,binary_c=temp/'query.c',temp/'query'
        run([binary,'codegen',source,cfile,'--transformer='+mode])
        run([cc,'-std=c11','-O2',cfile,'-o',binary_c])
        run([binary_c,facts,out])

records=[
 ('reachability_n250','tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl','fixtures/p2-scale-250'),
 ('ancestor_n500','tests/corpus/BENCHMARK_FAMILY/ancestor_nonancestor.dl','fixtures/benchmark-family/ancestor_nonancestor/n500'),
 ('same_generation_d4','tests/corpus/BENCHMARK_FAMILY/same_generation_negation.dl','fixtures/benchmark-family/same_generation_negation/d4_b4'),
 ('positive_n500','tests/corpus/BENCHMARK_FAMILY/transitive_closure_bound.dl','fixtures/benchmark-family/transitive_closure_bound/n500'),
 ('culprit_n20','tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl','fixtures/benchmark-family/culprit_cycle/n20'),
 *[(f'growing_n{n}','tests/corpus/CONE_CORPUS/cc_growing_sibling.dl',f'fixtures/cone_corpus/cc_growing_sibling/n{n}') for n in [20,50,100]],
]
summary={'mode':'integrated guarded-compact','oracle_available':bool(oracle),'benchmarks':[],'random_graphs':[],'projection_control':{},'browser_verified':False}
try:
    if oracle:summary['oracle_version']=run([oracle,'--version'])
    summary['cc_version']=run([cc,'--version']).split('\n')[0]
    for label,s,f in records:
        d=dest/label;d.mkdir();source=root/s;facts=root/f
        emit(source,'guarded',d/'guarded.dl');compact(source,d/'guarded.dl',d/'compact.dl')
        inspection=json.loads(run([binary,'inspect',source,'--transformer=guarded-compact']))
        (d/'inspection.json').write_text(json.dumps(inspection,indent=2)+'\n')
        expected=[r['name']+'.csv' for r in inspection['source']['relations'] if r['output']]
        row={'case':label,'source':s,'facts':f,'oracle_checked':bool(oracle)}
        if oracle:
            measured={}
            for mode,src in [('original',source),('guarded',d/'guarded.dl'),('compact',d/'compact.dl')]:
                out=d/mode;out.mkdir();run([oracle,'-F',facts,'-D',out,'-p',out/'profile.json',src])
                measured[mode]=answers(out,expected);row['T_'+mode]=analyze(out/'profile.json')['T_excl_copy']
            want=measured['original'];equal(measured['guarded'],want,label+'/guarded');equal(measured['compact'],want,label+'/compact')
        else:
            run([binary,'run-seminaive',source,facts,d/'reference']);want=answers(d/'reference',expected)
            row['materialization']='not measured: oracle unavailable'
        row['go_work']={}
        for mode in ['passthrough','guarded','guarded-compact']:
            for sub in ['run','run-seminaive']:
                out=d/(mode+'-'+sub)
                response=json.loads(run([binary,sub,source,facts,out,'--transformer='+mode]))
                equal(answers(out,expected),want,label+'/'+mode+'/'+sub)
                profile=json.loads((out/'profile.json').read_text())
                relations=profile['root']['program']['relation']
                totals={name:r.get('num-tuples',0)+sum(v.get('num-tuples',0) for v in r.get('iteration',{}).values()) for name,r in relations.items()}
                row['go_work'][mode+'/'+sub]={'materialized':sum(totals.values()),'attempts':response.get('derivation_attempts',0),'per_relation':totals}
        row['compaction_actions']=inspection['transform']['compaction']
        row['source_sha256']=hashlib.sha256(source.read_bytes()).hexdigest()
        row['fact_sha256']={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(facts.glob('*.facts'))}
        native(source,facts,d/'c',d,'guarded-compact');equal(answers(d/'c',expected),want,label+'/C')
        row['all_outputs_equal']=True;summary['benchmarks'].append(row);print(json.dumps({k:v for k,v in row.items() if k not in ['go_work','compaction_actions','fact_sha256']}),flush=True)

    for seed in range(24):
        rng=random.Random(seed);n=9;edges={(x,y) for x in range(1,n+1) for y in range(1,n+1) if rng.randrange(8)==0}
        reach=set(edges)
        while True:
            new=reach|{(x,z) for x,y in reach for yy,z in edges if y==yy}
            if new==reach:break
            reach=new
        d=dest/f'random-{seed:02}';d.mkdir();facts=d/'facts';facts.mkdir()
        (facts/'node.facts').write_text(''.join(f'{x}\n' for x in range(1,n+1)))
        (facts/'edge.facts').write_text(''.join(f'{x}\t{y}\n' for x,y in sorted(edges)))
        source=d/'source.dl'
        text='''.decl node(x:number)
.input node
.decl edge(x:number,y:number)
.input edge
.decl reach(x:number,y:number)
reach(x,y) :- edge(x,y).
reach(x,y) :- reach(x,z),edge(z,y).
.decl missing(x:number,y:number)
missing(x,y) :- node(x),node(y),!reach(x,y).
.decl out(y:number)
.output out
out(y) :- missing(1,y).
.decl other(y:number)
.output other
other(y) :- missing(3,y).
'''
        want={'out.csv':{str(y).encode() for y in range(1,n+1) if (1,y) not in reach},'other.csv':{str(y).encode() for y in range(1,n+1) if (3,y) not in reach}}
        if seed%2==0:text+='.output reach\n';want['reach.csv']={f'{x}\t{y}'.encode() for x,y in reach}
        if seed%3==0:
            text+='.decl untouched(x:number,y:number)\n.output untouched\nuntouched(x,y) :- node(x),node(y),!reach(x,y).\n'
            want['untouched.csv']={f'{x}\t{y}'.encode() for x in range(1,n+1) for y in range(1,n+1) if (x,y) not in reach}
        source.write_text(text);emit(source,'guarded',d/'guarded.dl');compact(source,d/'guarded.dl',d/'compact.dl')
        paths=[]
        for label,src,flag,sub in [('naive',source,'passthrough','run'),('semi',source,'passthrough','run-seminaive'),('guarded-naive',source,'guarded','run'),('guarded-semi',source,'guarded','run-seminaive'),('compact',source,'guarded-compact','run-seminaive'),('compact-naive',source,'guarded-compact','run')]:
            run([binary,sub,src,facts,d/label,'--transformer='+flag]);equal(answers(d/label,want),want,f'{seed}/{label}');paths.append(label)
        if seed<6:native(source,facts,d/'c',d,'guarded-compact');equal(answers(d/'c',want),want,f'{seed}/C');paths.append('C')
        if oracle:
            out=d/'oracle';out.mkdir();run([oracle,'-F',facts,'-D',out,d/'compact.dl']);equal(answers(out,want),want,f'{seed}/oracle');paths.append('oracle')
        summary['random_graphs'].append({'seed':seed,'nodes':n,'edges':len(edges),'checked_paths':paths,'independent_answers_equal':True})

    d=dest/'projection-control';d.mkdir();facts=d/'facts';facts.mkdir()
    (facts/'edge.facts').write_text(''.join(f'1\t{y}\n' for y in range(100)));(facts/'node.facts').write_text(''.join(f'{z}\n' for z in range(20)))
    source=d/'source.dl';source.write_text('''.decl edge(x:number,y:number)
.input edge
.decl node(z:number)
.input node
.decl p(x:number,z:number)
p(x,z) :- edge(x,y),node(z).
.decl out(z:number)
.output out
out(z) :- p(1,z).
''')
    emit(source,'guarded',d/'guarded.dl');compact(source,d/'guarded.dl',d/'compact.dl')
    for mode in ['guarded','compact']:
        j=json.loads(run([binary,'run-seminaive',d/(mode+'.dl'),facts,d/mode]))
        equal(answers(d/mode,['out.csv']),{'out.csv':{str(z).encode() for z in range(20)}},'projection/'+mode)
        summary['projection_control'][mode+'_attempts']=j.get('derivation_attempts',0)
    want={'out.csv':{str(z).encode() for z in range(20)}}
    native(source,facts,d/'native',d,'guarded-compact');equal(answers(d/'native',want),want,'projection/native')
    if oracle:
        out=d/'oracle';out.mkdir();run([oracle,'-F',facts,'-D',out,source]);equal(answers(out,want),want,'projection/oracle')
    if summary['projection_control']['compact_attempts']>summary['projection_control']['guarded_attempts']:
        raise RuntimeError('Projection barrier removed: extra derivation work on the supplied control.')
    summary['status']='PASS' if oracle else 'PARTIAL: oracle unavailable; no oracle/materialization claims'
    print(summary['status'],flush=True)
except Exception as exc:
    summary['status']='FAIL';summary['error']=str(exc);print(str(exc),file=sys.stderr)
finally:
    (dest/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
raise SystemExit(1 if summary['status']=='FAIL' else 0)
