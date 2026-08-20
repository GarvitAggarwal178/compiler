#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T2: corpus viability, untransformed only, over the 36
pre-registered programs (tests/corpus/PREREGISTERED.txt).

Reuses corpus_predicate.check_program (read-only import, not modified --
NIGHT-BATCH-01 hard prohibition #2) for the "seedable" and
"negated IDB literal" fields, since those are exactly conditions 2 and 1
of the same predicate. Adds: T_none (both copy conventions), fact input
row counts, and a fresh mechanical out-of-grammar feature scan (blueprint
v1.1/v1.2 §4 grammar has no functors, aggregates, components, records,
choice, or ADTs).

Global caps applied to every Souffle invocation: 300s timeout, 8GB
address-space limit. A run that exceeds either is recorded DNF with the
cap that fired, never retried.
"""
import json
import re
import resource
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from corpus_predicate import check_program, strip_comments  # noqa: E402
from tuple_report import analyze as tuple_analyze  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
PREREGISTERED = REPO / "tests" / "corpus" / "PREREGISTERED.txt"
MEASUREMENTS = REPO / "measurements"
OUT_DIR = MEASUREMENTS / "night01-t2"

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300

FUNCTOR_RE = re.compile(r"\.functor\b|@\w+\s*\(")
AGGREGATE_RE = re.compile(r"\b(count|sum|max|min|mean)\s*:")
COMPONENT_RE = re.compile(r"\.comp\b|\.init\b")
RECORD_RE = re.compile(r"\.type\s+\w+\s*=\s*\[")
CHOICE_RE = re.compile(r"choice-domain")
ADT_RE = re.compile(r"\.type\s+\w+\s*=.*\{")

INPUT_NAME_RE = re.compile(r"^\s*\.input\s+(\w+)", re.MULTILINE)
OUTPUT_NAME_RE = re.compile(r"^\s*\.output\s+(\w+)", re.MULTILINE)


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def out_of_grammar_features(text: str):
    feats = {}
    feats["functors"] = len(FUNCTOR_RE.findall(text))
    feats["aggregates"] = len(AGGREGATE_RE.findall(text))
    feats["components"] = len(COMPONENT_RE.findall(text))
    feats["records"] = len(RECORD_RE.findall(text))
    feats["choice"] = len(CHOICE_RE.findall(text))
    feats["adts"] = len(ADT_RE.findall(text))
    return feats


def count_fact_rows(facts_dir: Path, input_names):
    rows = {}
    for name in input_names:
        fp = facts_dir / f"{name}.facts"
        if fp.is_file():
            with open(fp, errors="replace") as f:
                rows[name] = sum(1 for _ in f)
        else:
            rows[name] = None  # not file-backed (e.g. facts given as literal clauses)
    return rows


def run_one(test_rel_path: str):
    test_dir = SOUFFLE_TESTS / test_rel_path
    dl_files = sorted(test_dir.glob("*.dl"))
    if not dl_files:
        return {"path": test_rel_path, "status": "error:no-dl-file"}
    dl_path = dl_files[0]

    facts_dir = test_dir / "facts" if (test_dir / "facts").is_dir() else test_dir

    safe_name = test_rel_path.replace("/", "__")
    workdir = OUT_DIR / safe_name
    workdir.mkdir(parents=True, exist_ok=True)
    log_name = "prof.log"

    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), "-p", log_name, str(dl_path)]
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")

    try:
        proc = subprocess.run(
            cmd, cwd=str(workdir), capture_output=True, text=True,
            timeout=TIMEOUT_S, preexec_fn=_limit_mem,
        )
    except subprocess.TimeoutExpired:
        (workdir / "meta.json").write_text(json.dumps({"status": "DNF:timeout-300s"}, indent=2))
        return {"path": test_rel_path, "status": "DNF:timeout-300s"}
    except MemoryError:
        (workdir / "meta.json").write_text(json.dumps({"status": "DNF:memcap-8gb"}, indent=2))
        return {"path": test_rel_path, "status": "DNF:memcap-8gb"}

    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)

    if proc.returncode != 0:
        result = {"path": test_rel_path, "status": f"error:returncode-{proc.returncode}",
                  "stderr_snippet": proc.stderr[:500]}
        (workdir / "meta.json").write_text(json.dumps(result, indent=2))
        return result

    raw = dl_path.read_text(errors="replace")
    text = strip_comments(raw)

    predicate_result = check_program(dl_path)
    input_names = sorted(set(INPUT_NAME_RE.findall(text)))
    output_names = sorted(set(OUTPUT_NAME_RE.findall(text)))
    fact_rows = count_fact_rows(facts_dir, input_names)
    feats = out_of_grammar_features(text)

    log_path = workdir / log_name
    if log_path.is_file():
        tr = tuple_analyze(log_path)
        t_none_excl, t_none_incl = tr["T_excl_copy"], tr["T_incl_copy"]
    else:
        t_none_excl = t_none_incl = None

    result = {
        "path": test_rel_path,
        "status": "ok",
        "dl_file": dl_path.name,
        "T_none_excl_copy": t_none_excl,
        "T_none_incl_copy": t_none_incl,
        "input_names": input_names,
        "output_names": output_names,
        "fact_rows": fact_rows,
        "seedable": predicate_result["output_with_literal"],
        "negated_idb_count": len(predicate_result["negated_idb_targets"]),
        "negated_idb_targets": predicate_result["negated_idb_targets"],
        "out_of_grammar": feats,
        "out_of_grammar_total": sum(feats.values()),
    }
    (workdir / "meta.json").write_text(json.dumps(result, indent=2))
    return result


def main():
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    paths = [p.strip() for p in PREREGISTERED.read_text().splitlines() if p.strip()]
    results = []
    for p in paths:
        r = run_one(p)
        print(f"{p}: {r['status']}", file=sys.stderr)
        results.append(r)

    (OUT_DIR / "summary.json").write_text(json.dumps(results, indent=2))
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
