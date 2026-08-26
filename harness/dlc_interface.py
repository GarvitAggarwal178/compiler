#!/usr/bin/env python3
"""
dlc's lex/parse/roundtrip interface, as understood by M1's Lane B harness.
Per docs/M1-BUILD.md §1 (supersedes CLAUDE.md §2), the lexer and parser
are Lane B now -- not "Lane A, does not exist yet" as this module's
docstring originally said. `run_dlc_lex` (§3.1) and `run_dlc_parse`/
`run_dlc_roundtrip` (§3.3) are real, calling the built `dlc` binary.

`run_dlc_roundtrip` replaces the originally-planned
run_dlc_parse+run_dlc_pretty_print pair for the round-trip gate: the
parse->print->reparse->compare sequence is done entirely inside the `dlc
roundtrip` subcommand (Go, using ast.Equal directly -- src/parser/
DESIGN.md explains why the comparison belongs there, not reimplemented
against a JSON AST dump in Python). `run_dlc_pretty_print`/`PrintResult`
are gone; nothing outside harness/round_trip_scaffold.py ever called them,
and that file is rewritten to call run_dlc_roundtrip instead.

Separate from harness/differential.py's run_dlc(), which is about full
compile+run (comparing output relations against Soufflé). These functions
are about lexing/parsing specifically -- narrower operations that don't
need a facts directory or execution at all.
"""
import json
import subprocess
import sys
import tempfile
from dataclasses import dataclass, field
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
DLC_BINARY = REPO / "bin" / "dlc"  # gitignored build artifact, not provenance -- see .gitignore


@dataclass
class ParseResult:
    status: str  # "parsed" | "error" | "panic" | "read_error" | "build_missing"
    decl_count: int = 0
    clause_count: int = 0
    error_count: int = 0
    diagnostics: list = field(default_factory=list)
    diagnostic: str = ""  # top-level panic/read-error message, if any


@dataclass
class RoundtripResult:
    status: str  # "match" | "mismatch" | "parse_error" | "reparse_error" | "panic" | "build_missing"
    diagnostics: list = field(default_factory=list)
    printed: str = ""
    diagnostic: str = ""


@dataclass
class LexResult:
    status: str  # "lexed" | "panic" | "read_error" | "build_missing"
    token_count: int = 0
    error_count: int = 0
    tokens: list = field(default_factory=list)
    diagnostic: str = ""


def build_dlc():
    """Builds the dlc binary to a fixed path, once, so per-file harness
    runs invoke a binary instead of paying `go run` compile cost per
    file. Raises if the build fails -- a build failure is a build/CI
    plumbing problem (CLAUDE.md §5 CONTINUE-and-log bucket), not silently
    swallowed here."""
    DLC_BINARY.parent.mkdir(parents=True, exist_ok=True)
    proc = subprocess.run(
        ["go", "build", "-o", str(DLC_BINARY), "./src/cmd/dlc"],
        cwd=str(REPO), capture_output=True, encoding="utf-8", errors="replace",
    )
    if proc.returncode != 0:
        raise RuntimeError(f"go build failed:\n{proc.stdout}\n{proc.stderr}")


def run_dlc_lex(file_path) -> LexResult:
    """Runs the real `dlc lex <file>` binary (§3.1) and parses its JSON
    output. Never raises on a lexer-level problem -- a Go panic inside
    dlc is reported back as status="panic" (see cmd/dlc's top-level
    recover, src/cmd/dlc/DESIGN.md), not as a Python exception, so a
    caller iterating many files can just record the status and continue."""
    if not DLC_BINARY.is_file():
        return LexResult(status="build_missing", diagnostic=f"{DLC_BINARY} does not exist; call build_dlc() first")
    proc = subprocess.run(
        [str(DLC_BINARY), "lex", str(file_path)],
        capture_output=True, encoding="utf-8", errors="replace",
    )
    if proc.returncode != 0 and not proc.stdout.strip():
        return LexResult(status="panic", diagnostic=f"nonzero exit {proc.returncode}, no output; stderr: {proc.stderr[:500]}")
    try:
        doc = json.loads(proc.stdout.strip().splitlines()[-1]) if proc.stdout.strip() else {}
    except (json.JSONDecodeError, IndexError) as e:
        return LexResult(status="panic", diagnostic=f"non-JSON output (possible uncontrolled crash): {e}; stdout: {proc.stdout[:500]}; stderr: {proc.stderr[:500]}")
    return LexResult(
        status=doc.get("status", "panic"),
        token_count=doc.get("token_count", 0),
        error_count=doc.get("error_count", 0),
        tokens=doc.get("tokens", []),
        diagnostic=doc.get("panic", ""),
    )


def _run_dlc_on_text(subcommand: str, source_text: str):
    """Shared plumbing for run_dlc_parse/run_dlc_roundtrip: write
    source_text to a temp file (the CLI reads files, not stdin, to keep
    its interface identical for every subcommand), invoke `dlc
    <subcommand> <tmpfile>`, and return the parsed JSON document or None
    plus a diagnostic string on any panic/non-JSON-output condition."""
    if not DLC_BINARY.is_file():
        return None, f"{DLC_BINARY} does not exist; call build_dlc() first"
    with tempfile.NamedTemporaryFile(mode="w", suffix=".dl", delete=False) as f:
        f.write(source_text)
        tmp_path = f.name
    try:
        proc = subprocess.run(
            [str(DLC_BINARY), subcommand, tmp_path],
            capture_output=True, encoding="utf-8", errors="replace",
        )
    finally:
        Path(tmp_path).unlink(missing_ok=True)
    if proc.returncode != 0 and not proc.stdout.strip():
        return None, f"nonzero exit {proc.returncode}, no output; stderr: {proc.stderr[:500]}"
    try:
        doc = json.loads(proc.stdout.strip().splitlines()[-1]) if proc.stdout.strip() else {}
    except (json.JSONDecodeError, IndexError) as e:
        return None, f"non-JSON output (possible uncontrolled crash): {e}; stdout: {proc.stdout[:500]}; stderr: {proc.stderr[:500]}"
    return doc, None


def run_dlc_parse(source_text: str) -> ParseResult:
    """Runs the real `dlc parse` subcommand (§3.3 gates one and three)
    and parses its JSON output. Never raises on a compiler-level problem
    -- reported back as status="panic", same convention as run_dlc_lex."""
    doc, err = _run_dlc_on_text("parse", source_text)
    if doc is None:
        return ParseResult(status="panic", diagnostic=err)
    if "panic" in doc and doc.get("status") not in ("parsed", "error"):
        return ParseResult(status=doc.get("status", "panic"), diagnostic=doc.get("panic", ""))
    return ParseResult(
        status=doc.get("status", "panic"),
        decl_count=doc.get("decl_count", 0),
        clause_count=doc.get("clause_count", 0),
        error_count=doc.get("error_count", 0),
        diagnostics=doc.get("diagnostics", []),
    )


def run_dlc_roundtrip(source_text: str) -> RoundtripResult:
    """Runs the real `dlc roundtrip` subcommand (§3.3 gate two): parse ->
    print -> reparse -> ast.Equal, entirely inside dlc. Never raises."""
    doc, err = _run_dlc_on_text("roundtrip", source_text)
    if doc is None:
        return RoundtripResult(status="panic", diagnostic=err)
    if "panic" in doc and doc.get("status") not in ("match", "mismatch", "parse_error", "reparse_error"):
        return RoundtripResult(status=doc.get("status", "panic"), diagnostic=doc.get("panic", ""))
    return RoundtripResult(
        status=doc.get("status", "panic"),
        diagnostics=doc.get("diagnostics", []),
        printed=doc.get("printed", ""),
    )
