"""Exact output sets within dlc's LF-delimited plain TSV contract."""
from pathlib import Path


def read_outputs(directory, expected_names):
    directory = Path(directory)
    expected = set(expected_names)
    actual = {p.name for p in directory.glob('*.csv')}
    if actual != expected:
        raise ValueError(f'output files differ: missing={sorted(expected-actual)!r}, extra={sorted(actual-expected)!r}')
    result = {}
    for name in sorted(expected):
        data = (directory / name).read_bytes()
        if data and not data.endswith(b'\n'):
            raise ValueError(f'unterminated LF row in {name}')
        result[name] = set(data.split(b'\n')[:-1]) if data else set()
    return result


def differences(actual, expected):
    return {name: {'extra': [repr(v) for v in sorted(actual.get(name, set())-expected.get(name, set()))],
                   'missing': [repr(v) for v in sorted(expected.get(name, set())-actual.get(name, set()))]}
            for name in sorted(actual.keys() | expected.keys())
            if name not in actual or name not in expected or actual[name] != expected[name]}
