#!/usr/bin/env python3
"""
Canonical stub definitions for dlc's (Lane A, does not exist yet)
parse/print interface, as understood by M1's Lane B harness. Every
function here is a deliberate stub returning status="not_implemented"
until Lane A lands; the shape (what goes in, what comes back) is fixed
now so Lane A has an explicit contract and this harness code does not
change shape when the parser lands -- only these two function bodies do.

Separate from harness/differential.py's run_dlc(), which is about full
compile+run (comparing output relations against Soufflé). These two
functions are about parsing and pretty-printing specifically -- narrower
operations that don't need a facts directory or execution at all.
"""
from dataclasses import dataclass


@dataclass
class ParseResult:
    status: str  # "not_implemented" | "parsed" | "error"
    ast: object = None
    diagnostic: str = ""


@dataclass
class PrintResult:
    status: str  # "not_implemented" | "printed" | "error"
    text: str = ""
    diagnostic: str = ""


def run_dlc_parse(source_text: str) -> ParseResult:
    """STUB. M1 (Lane A) replaces this body with a real call into dlc's
    parser, returning the real AST in `.ast`. Everything that consumes
    this function (parse_coverage.py, round_trip_scaffold.py,
    run_rejection_tests.py) is already correct for that day."""
    return ParseResult(status="not_implemented",
                        diagnostic="dlc parse entry point not implemented yet (M1, Lane A)")


def run_dlc_pretty_print(ast) -> PrintResult:
    """STUB. M1 (Lane A) replaces this body with dlc's pretty-printer,
    once it exists, returning the printed source text in `.text`."""
    return PrintResult(status="not_implemented",
                        diagnostic="dlc pretty-printer not implemented yet (M1, Lane A)")
