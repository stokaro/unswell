"""Create deterministic archives of completed local study artifacts."""
import argparse
import gzip
import hashlib
import json
import pathlib
import tarfile

parser = argparse.ArgumentParser()
parser.add_argument('--study', type=pathlib.Path, required=True)
parser.add_argument('--output', type=pathlib.Path, required=True)
parser.add_argument('--kind', choices=['outputs', 'measurements', 'features'], required=True)
args = parser.parse_args()
study = args.study.resolve()
output = args.output.resolve()
output.mkdir(parents=True, exist_ok=True)
if args.kind == 'outputs':
    files = [p for p in (study / 'runs').rglob('*') if p.is_file() and p.name != 'requests.json']
    files.append(study / 'generation-start.json')
    for name in ['generation-completion.json', 'completion-driver.py']:
        if (study / name).exists():
            files.append(study / name)
    for family in ['luna', 'haiku', 'qwen']:
        assert len(list((study / 'runs' / family / 'raw').glob('*/result.json'))) == 152
        assert (study / 'runs' / family / 'records.json').exists()
elif args.kind == 'features':
    files = [p for p in (study / 'analysis/features').glob('*') if p.is_file()]
    historical = study / 'analysis/features-historical.json'
    if historical.exists():
        files.append(historical)
else:
    files = [p for p in (study / 'analysis').rglob('*') if p.is_file() and
             'features' not in p.relative_to(study / 'analysis').parts and not p.name.startswith('features-') and
             p.suffix not in ['.log', '.pending']]
files.sort(key=lambda p: str(p.relative_to(study)))
archive = output / (args.kind + '.tar.gz')
manifest = {}
with archive.open('wb') as raw, gzip.GzipFile(fileobj=raw, mode='wb', filename='', mtime=0) as compressed, tarfile.open(fileobj=compressed, mode='w') as tar:
    for path in files:
        assert not path.is_symlink()
        name = str(path.relative_to(study))
        info = tarfile.TarInfo(name)
        info.size = path.stat().st_size
        info.mode = 0o644
        with path.open('rb') as source:
            manifest[name] = hashlib.file_digest(source, 'sha256').hexdigest()
            source.seek(0)
            tar.addfile(info, source)
with archive.open('rb') as source:
    digest = hashlib.file_digest(source, 'sha256').hexdigest()
record = dict(version='unswell-long-prose-result-archive-v1', kind=args.kind, file=archive.name,
              bytes=archive.stat().st_size, sha256=digest, files=manifest)
(output / (args.kind + '.json')).write_text(json.dumps(record, indent=2) + '\n')
print(archive.name, len(files), 'files', archive.stat().st_size, 'bytes', digest)
