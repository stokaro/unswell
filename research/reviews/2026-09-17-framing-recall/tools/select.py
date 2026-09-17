"""Freeze confirmation pages by identity, excluding the development audit."""
import argparse
import gzip
import hashlib
import io
import json
from pathlib import Path
import tarfile


def digest(data):
    return hashlib.sha256(data).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--ptah-sources', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError('Output must be new')
    repo = Path(__file__).resolve().parents[4]
    development = repo/'research/reviews/2026-09-17-full-page-recall/manifest.json'
    manifest = json.loads(development.read_text())
    used = {p['reference'] for p in manifest['pages']}
    pages = []
    for cohort in ('ptah', 'historical'):
        for fmt in (('markdown', 'mdx') if cohort == 'ptah' else ('any',)):
            for length in ('short', 'medium', 'long'):
                candidates = [p for p in manifest['frame'] if p['cohort'] == cohort and p['reference'] not in used
                              and p['length_stratum'] == length and (fmt == 'any' or p['format'] == fmt)]
                def rank(p):
                    return digest(('issue299-confirmation-v1\n'+p['reference']).encode())
                row = min(candidates, key=rank)
                pid = f'c{len(pages)+1:02}'
                pages.append(dict(row, id=pid, path='sources/'+pid+Path(row['original_path']).suffix,
                                  confirmation_rank=rank(row), eligible=len(candidates)))
    files = {}
    history = repo/'research/generation/studies/long-prose-v3/inputs.tar.gz'
    if digest(history.read_bytes()) != manifest['historical_archive_sha256']:
        raise ValueError('Historical archive drift')
    with tarfile.open(history) as archive:
        for page in pages:
            notices = {}
            if page['cohort'] == 'ptah':
                data = (args.ptah_sources/page['input_path']).read_bytes()
                notices['notices/ptah/LICENSE'] = (repo/'e2e/rhetoricdata/LICENSE.ptah').read_bytes()
            else:
                data = archive.extractfile(page['input_path']).read()
                prefix = 'sources/historical/'+page['repository'].replace('/', '__')+'/'
                for notice in page['notices']:
                    content = archive.extractfile(prefix+notice['path']).read()
                    if digest(content) != notice['sha256']:
                        raise ValueError('Notice drift')
                    notices['notices/'+page['repository'].replace('/', '__')+'/'+notice['path']] = content
            if digest(data) != page['sha256'] or len(data) != page['bytes']:
                raise ValueError('Source drift')
            files[page['path']] = data
            files.update(notices)
            page['retained_notices'] = [dict(path=p, sha256=digest(b)) for p,b in notices.items()]
    output = io.BytesIO()
    with tarfile.open(fileobj=output, mode='w') as archive:
        for path,data in sorted(files.items()):
            info = tarfile.TarInfo(path)
            info.size, info.mode = len(data), 0o644
            archive.addfile(info, io.BytesIO(data))
    args.output.mkdir(parents=True)
    (args.output/'inputs.tar.gz').write_bytes(gzip.compress(output.getvalue(), mtime=0))
    result = dict(version=1, selection='issue299-confirmation-v1', development_manifest_sha256=digest(development.read_bytes()),
                  inputs_sha256=digest((args.output/'inputs.tar.gz').read_bytes()), pages=pages)
    (args.output/'manifest.json').write_text(json.dumps(result, indent=2)+'\n')


if __name__ == '__main__':
    main()
