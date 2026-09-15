"""Verify every archive member, optionally extracting into a new directory."""
import argparse
import contextlib
import hashlib
import json
import pathlib
import tarfile

parser = argparse.ArgumentParser()
parser.add_argument('--archives', type=pathlib.Path, required=True)
parser.add_argument('--extract-to', type=pathlib.Path)
args = parser.parse_args()
archives = args.archives.resolve()
if args.extract_to:
    args.extract_to.mkdir(parents=True, exist_ok=False)
for name in ['inputs', 'selection-audit', 'outputs', 'measurements', 'features']:
    record = json.loads((archives / (name + '.json')).read_text())
    archive = archives / (name + '.tar.gz')
    assert record['file'] == archive.name
    assert archive.stat().st_size == record['bytes']
    with archive.open('rb') as source:
        assert hashlib.file_digest(source, 'sha256').hexdigest() == record['sha256']
    expected = record.get('files')
    if name == 'inputs':
        with tarfile.open(archive, 'r:gz') as tar:
            frozen = tar.extractfile('freeze.json').read()
        assert hashlib.sha256(frozen).hexdigest() == record['freeze_sha256']
        expected = dict(json.loads(frozen)['files'], **{'freeze.json': record['freeze_sha256']})
    assert expected
    with tarfile.open(archive, 'r|gz') as tar:
        seen = set()
        for member in tar:
            path = pathlib.PurePosixPath(member.name)
            assert member.isfile() and not path.is_absolute() and '..' not in path.parts
            assert '\\' not in member.name and ':' not in member.name
            assert str(path) == member.name and member.name in expected and member.name not in seen
            seen.add(member.name)
            digest = hashlib.sha256()
            with contextlib.ExitStack() as stack:
                source = stack.enter_context(tar.extractfile(member))
                destination = None
                if args.extract_to:
                    target = args.extract_to / pathlib.Path(*path.parts)
                    target.parent.mkdir(parents=True, exist_ok=True)
                    destination = stack.enter_context(target.open('xb'))
                while chunk := source.read(1 << 20):
                    digest.update(chunk)
                    if destination:
                        destination.write(chunk)
            assert digest.hexdigest() == expected[member.name]
        assert seen == set(expected)
    print(name, len(seen), 'members verified')
