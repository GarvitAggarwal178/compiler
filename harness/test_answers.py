import tempfile
import unittest
from pathlib import Path
from answers import read_outputs


class AnswerTests(unittest.TestCase):
    def test_exact_rows_and_manifest(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            p = root/'out.csv'
            for data, expected in [(b'', set()), (b'\n', {b''}), (b'()\n', {b'()'}),
                                   ('a\u2028b\n'.encode(), {'a\u2028b'.encode()}),
                                   (b'a\nb\na\n', {b'a',b'b'}), (b'a\r\n', {b'a\r'})]:
                p.write_bytes(data)
                self.assertEqual(read_outputs(root, ['out.csv']), {'out.csv': expected})
            p.write_bytes(b'a')
            with self.assertRaises(ValueError): read_outputs(root, ['out.csv'])
            p.write_bytes(b'')
            with self.assertRaises(ValueError): read_outputs(root, ['missing.csv'])
            (root/'extra.csv').write_bytes(b'')
            with self.assertRaises(ValueError): read_outputs(root, ['out.csv'])

    def test_equal_cardinality_does_not_imply_equality(self):
        with tempfile.TemporaryDirectory() as td:
            root=Path(td); (root/'out.csv').write_bytes(b'a\nb\n')
            left=read_outputs(root,['out.csv'])
            (root/'out.csv').write_bytes('a\u2028b\n'.encode())
            self.assertNotEqual(left,read_outputs(root,['out.csv']))
            (root/'out.csv').write_bytes(b'c\nd\n')
            self.assertNotEqual(left,read_outputs(root,['out.csv']))
