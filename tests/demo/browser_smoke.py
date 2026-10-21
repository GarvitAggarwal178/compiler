#!/usr/bin/env python3
"""Exercise the real local workbench through Chromium. Requires Playwright.

Run make demo first. No recorded responses or substituted compiler outputs.
"""
import argparse
import hashlib
import gzip
import json
import os
from pathlib import Path
import socket
import subprocess
import time
from urllib.request import urlopen
from playwright.sync_api import sync_playwright

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--url',default='http://127.0.0.1:8080')
p.add_argument('--repo',type=Path,default=Path.cwd())
p.add_argument('--output',type=Path,required=True)
p.add_argument('--stage',choices=['review2','final'],default='review2')
a=p.parse_args();root=a.repo.resolve();a.output.mkdir(parents=True,exist_ok=False)
checks=[];records=[];errors=[];servers=[];browser_version=None

def record(response):
    if '/api/run' not in response.url and '/api/native/' not in response.url and '/api/oracle/' not in response.url:return
    try:
        if response.status==200:
            data=response.json();records.append({'url':response.url,'request':response.request.post_data_json,'response':data})
    except Exception:
        pass  # A cancelled HTTP response has no JSON document; UI state is checked below.

def last_result(page):
    run=page.locator('#snapshot').get_attribute('data-run-id')
    return next(x['response'] for x in reversed(records) if x['response'].get('run_id')==run and 'evaluations' in x['response'])

def wait_state(page,text,timeout=15000):
    page.wait_for_function('(text)=>document.querySelector("#status").textContent.includes(text)',arg=text,timeout=timeout)

def run(page):
    page.locator('#run').click();wait_state(page,'Completed real executions');return last_result(page)

def rows(page,name):
    d=page.locator(f'xpath=//div[@id="results"]//details[summary[starts-with(text(), "{name} ·")]]')
    return set(d.locator('tbody tr td:first-child').all_text_contents())

def check(name,fn):
    fn();checks.append({'check':name,'status':'PASS'});print('PASS',name,flush=True)

def controlled(extra):
    with socket.socket() as sock:
        sock.bind(('127.0.0.1',0));port=sock.getsockname()[1]
    log=(a.output/('controlled-'+str(port)+'.log')).open('w')
    cmd=[str(root/'bin/dlc-demo'),'-repo',str(root),'-addr',f'127.0.0.1:{port}',*extra]
    proc=subprocess.Popen(cmd,stdout=log,stderr=subprocess.STDOUT);servers.append((proc,log));url=f'http://127.0.0.1:{port}'
    for _ in range(100):
        try:
            with urlopen(url+'/api/preflight',timeout=1) as r:r.read()
            return url
        except Exception:
            if proc.poll() is not None:raise RuntimeError('controlled server exited')
            time.sleep(.03)
    raise RuntimeError('controlled server did not start')

try:
 with sync_playwright() as pw:
    browser=pw.chromium.launch(headless=True,args=['--no-sandbox','--disable-dev-shm-usage'])
    browser_version=browser.version
    page=browser.new_page(viewport={'width':1536,'height':1100});page.on('response',record);page.on('pageerror',lambda e:errors.append(str(e)))
    page.goto(a.url);wait_state(page,'Inspection ready')
    original=page.locator('#source').input_value()
    def initial():
        r=run(page);assert rows(page,'out.csv')=={'"1"','"5"','"6"','"7"','"8"'}
        assert [(e['materialized'],e['attempts']) for e in r['evaluations'][:2]]==[(69,69),(38,46)]
        edge=page.locator('#dependencies .graph-hit[data-from="reach"][data-to="reach"]')
        point=edge.evaluate('(p)=>{const a=p.getPointAtLength(p.getTotalLength()/2),m=p.getScreenCTM();return {x:a.x*m.a+a.y*m.c+m.e,y:a.x*m.b+a.y*m.d+m.f}}')
        page.mouse.click(point['x'],point['y'])
        assert page.locator('#rule').input_value()=='rule_1'
        page.locator('#ast-tree').locator('..').evaluate('(d)=>d.open=true')
        page.locator('#ast-tree button[data-node-id="rule_1/head"]').click()
        assert '"kind": "Atom"' in page.locator('#ast-node').inner_text()
        assert page.locator('#source').evaluate('(x)=>x.selectionEnd>x.selectionStart')
        assert page.locator('#dependencies [data-negative="true"]').count()>0
        if a.stage=='review2':page.screenshot(path=str(a.output/'review2-live.png'),full_page=True)
    check('real AST, dependencies, graph answers and live counters',initial)
    first=last_result(page)
    def query_edit():
        page.locator('#source').fill(original.replace('missing(1,y)','missing(5,y)'))
        assert page.locator('#results').get_by_text('All requested output sets equal',exact=True).count()==0
        assert page.locator('#native').is_disabled()
        r=run(page);assert rows(page,'out.csv')=={'"1"','"2"','"3"','"4"','"5"'}
        assert r['source_sha256']!=first['source_sha256'] and r['input_sha256']!=first['input_sha256']
    check('query mutation and immediate stale-result invalidation',query_edit)
    def graph_edit():
        page.locator('#example').select_option('reachability');page.locator('#load-example').click();wait_state(page,'Inspection ready')
        page.locator('#edge-from').select_option('4');page.locator('#edge-to').select_option('5');page.locator('#add-edge').click()
        assert '4\t5\n' in page.locator('textarea[data-name="edge"]').input_value()
        r=run(page);assert rows(page,'out.csv')=={'"1"'} and r['input_sha256']!=first['input_sha256']
    check('input graph edits update facts, snapshot and answers',graph_edit)
    def full_output():
        page.locator('#example').select_option('reachability');page.locator('#load-example').click();wait_state(page,'Inspection ready')
        page.locator('#source').fill(original+'\n.output reach\n');r=run(page)
        assert len(rows(page,'reach.csv'))==12 and len(rows(page,'out.csv'))==5
        d=next(d for d in r['inspection']['transform']['decisions'] if d['source_predicate']=='reach')
        assert d['specialized'] and d['full_extent_retained'];assert r['evaluations'][1]['materialized']==50
    check('direct full output preserves all 12 reach pairs',full_output)
    def second_query():
        page.locator('#source').fill(original+'\n.decl other(y:number)\n.output other\nother(y) :- missing(5,y).\n');r=run(page)
        assert len(r['inspection']['transform']['queries'])==2
        assert rows(page,'other.csv')=={'"1"','"2"','"3"','"4"','"5"'} and len(rows(page,'out.csv'))==5
    check('second bound query has a real seed and a separate output',second_query)
    def no_query():
        page.locator('#source').fill(original.replace('missing(1,y)','missing(x,y)'));r=run(page)
        assert not r['inspection']['transform']['queries'] and r['inspection'].get('candidate') is None
        assert r['inspection']['transform']['no_query_reason'] and not any(d['specialized'] for d in r['inspection']['transform']['decisions'])
        assert r['evaluations'][0]['materialized']==r['evaluations'][1]['materialized']
        assert len(rows(page,'out.csv'))==8
    check('removing constants is an honest no-op specialization',no_query)
    custom='.decl result(v:number)\n.output result\nresult(41+1).\n'
    def custom_error_fix():
        page.locator('#source').fill(custom);r=run(page);assert rows(page,'result.csv')=={'"42"'}
        assert r['source_sha256']==hashlib.sha256(custom.encode()).hexdigest()
        page.locator('#source').fill(custom.replace('41+1','"wrong"'));page.locator('#run').click();wait_state(page,'rejected:')
        assert 'type_mismatch' in page.locator('#compile-status').inner_text()
        assert not page.locator('#results .badge').count() and page.locator('#native').is_disabled()
        page.locator('#source').fill(custom);run(page);assert rows(page,'result.csv')=={'"42"'}
    check('new custom source, located semantic failure and successful repair',custom_error_fix)
    def fallback():
        page.locator('#example').select_option('mixed-fallback');page.locator('#load-example').click();wait_state(page,'Inspection ready');r=run(page)
        j=r['inspection'];assert j['source']['status']=='valid' and j['candidate']['status']=='unstratifiable' and j['final']['status']=='valid'
        d={d['source_predicate']:d for d in j['transform']['decisions']};assert all(d[n]['fallback'] for n in ['p','q','s']);assert d['tc']['specialized'] and not d['tc']['fallback']
        assert 'candidate: unstratifiable' in page.locator('#guard').inner_text()
        if a.stage=='review2':page.screenshot(path=str(a.output/'mixed-fallback-live.png'),full_page=True)
    check('mixed fallback shows valid source, invalid candidate and valid final',fallback)
    def native():
        page.locator('#native').click();page.wait_for_function('()=>document.querySelector("#external").textContent.includes("native: ok")',timeout=15000)
        assert 'All requested output sets equal' in page.locator('#external').inner_text()
        details=page.locator('#c-source').locator('..')
        if details.get_attribute('open') is None:details.locator('summary').click()
        c=page.locator('#c-source').inner_text();assert 'source SHA-256' in c and 'while (changed)' in c and 'source rule_' in c, 'displayed C lacks actual provenance or loop'
    check('generated C compiles, executes and matches every output',native)
    def oracle():
        if page.locator('#oracle').is_disabled():raise RuntimeError('Souffle not available for requested oracle browser check')
        page.locator('#oracle').click();page.wait_for_function('()=>document.querySelector("#external").textContent.includes("oracle: ok")',timeout=15000)
        assert 'All requested output sets equal' in page.locator('#external').inner_text()
    check('optional live Souffle answer verification',oracle)
    def timeout():
        page.locator('#source').fill('.decl p(x:number)\n.output p\np(0).\np(x+1) :- p(x).\n');page.locator('#run').click();wait_state(page,'timeout:',10000)
        r=last_result(page);timed=[c for c in r['commands'] if c['status']=='timeout'];assert timed and not r.get('comparison')
        for c in timed:
            assert c['process_group_cleanup'] and c['process_reaped']
            try:os.kill(c['pid'],0)
            except ProcessLookupError:pass
            else:raise AssertionError('timed-out compiler remains live')
        page.locator('#source').fill(custom);run(page);assert rows(page,'result.csv')=={'"42"'}
    check('bounded recursion timeout, process cleanup and successful next run',timeout)
    def rapid():
        page.locator('#source').fill('.decl p(x:number) .output p p(0). p(x+1) :- p(x).');page.locator('#run').click()
        page.wait_for_timeout(150)
        latest=custom.replace('41+1','7');page.locator('#source').fill(latest);page.locator('#run').click();wait_state(page,'Completed real executions')
        page.wait_for_timeout(2300);assert rows(page,'result.csv')=={'"7"'}
        assert last_result(page)['source_sha256']==hashlib.sha256(latest.encode()).hexdigest()
    check('rapid edit cancels old work and old responses cannot replace current results',rapid)
    def missing():
        url=controlled(['-oracle','/missing/souffle','-timeout','2s'])
        other=browser.new_page();other.on('response',record);other.on('pageerror',lambda e:errors.append(str(e)));other.goto(url);wait_state(other,'Inspection ready');run(other)
        assert 'oracle: unavailable' in other.locator('#tools').inner_text() and other.locator('#oracle').is_disabled()
        assert other.locator('#native').is_enabled();other.locator('#native').click();other.wait_for_function('()=>document.querySelector("#external").textContent.includes("native: ok")',timeout=15000);other.close()
        url=controlled(['-cc','/missing/cc','-oracle','/missing/souffle'])
        other=browser.new_page();other.goto(url);wait_state(other,'Inspection ready');other.locator('#run').click();wait_state(other,'Completed real executions')
        assert 'cc: unavailable' in other.locator('#tools').inner_text() and other.locator('#native').is_disabled() and other.locator('#oracle').is_disabled();other.close()
    check('visible missing tools; native execution works without Souffle',missing)
    if a.stage=='final':
        def compact():
            page.locator('#example').select_option('reachability');page.locator('#load-example').click();wait_state(page,'Inspection ready');page.locator('#mode').select_option('guarded-compact');r=run(page)
            actions=r['inspection']['transform']['compaction'];assert any(x['action']=='remove' for x in actions) and any(x['action']=='skip' for x in actions)
            assert page.locator('#compact-details').is_visible() and len(rows(page,'out.csv'))==5
            native()
            page.locator('#compact-details').locator('summary').click()
            page.locator('#graph-stage').select_option('final')
            hit=page.locator('#dependencies .graph-hit[data-from="out"][data-to="missing_bf"]');hit.focus();page.keyboard.press('Enter')
            assert page.locator('#rule').input_value()=='rule_3', 'final graph lost its source-rule link'
            page.screenshot(path=str(a.output/'final-compact-live.png'),full_page=True)
            page.locator('#example').select_option('projection-barrier');page.locator('#load-example').click();wait_state(page,'Inspection ready');r=run(page)
            assert any(x['action']=='skip' and 'projection' in x['reason'] for x in r['inspection']['transform']['compaction'])
            assert len(rows(page,'out.csv'))==20
            page.locator('#source').fill(custom);r=run(page);assert all(x['action']!='remove' for x in r['inspection']['transform']['compaction'])
            assert rows(page,'result.csv')=={'"42"'}
        check('compact mode reports real removals/skips, matches native C and permits no-op',compact)
    assert not errors,errors
    browser.close()
except Exception as exc:
 checks.append({'check':'browser workflow','status':'FAIL','detail':str(exc)})
 print('FAIL',str(exc),flush=True)
 raise
finally:
 for proc,log in servers:
    proc.terminate()
    try:proc.wait(timeout=5)
    except subprocess.TimeoutExpired:proc.kill();proc.wait()
    log.close()
 (a.output/'checks.json').write_text(json.dumps({'stage':a.stage,'browser_version':browser_version,'checks':checks,'page_errors':errors},indent=2)+'\n')
 with gzip.open(a.output/'runs.json.gz','wt',encoding='utf-8') as f:json.dump(records,f,indent=2);f.write('\n')
