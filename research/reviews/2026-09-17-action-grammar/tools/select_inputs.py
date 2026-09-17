#!/usr/bin/env python3
"""Freeze source identities from expanded pinned metadata before prose review."""
import argparse
import gzip
import hashlib
import io
import json
from pathlib import Path, PurePosixPath
import tarfile

ROOT = Path(__file__).resolve().parents[4]
SEED = 'action-grammar-confirmation-v1'


def digest(data):
    return hashlib.sha256(data).hexdigest()


def read(path):
    return json.loads(path.read_text())


def checked_file(root, name, expected):
    path = PurePosixPath(name)
    if path.is_absolute() or '..' in path.parts:
        raise ValueError('Invalid input path')
    file = root/name
    if not file.resolve().is_relative_to(root.resolve()):
        raise ValueError('Input escaped root')
    data = file.read_bytes()
    if digest(data) != expected:
        raise ValueError('Source or notice drift: '+name)
    return data


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--ptah-sources', type=Path, required=True)
    parser.add_argument('--historical-sources', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError('Output must be new')
    review = ROOT/'research/reviews'
    original = review/'2026-09-17-full-page-recall/manifest.json'
    base = read(original)
    previous = [original] + [review/('2026-09-17-'+name)/'confirmation/manifest.json'
                            for name in ('framing-recall', 'repetition-scope', 'local-repetition', 'context-recall', 'instruction-recall', 'construction-recall', 'purpose-recall', 'verb-scaffolding', 'rhetoric-relations', 'instruction-projection', 'rhetoric-scope', 'proposition-repetition', 'instruction-clauses')]
    used = [page for path in previous for page in read(path)['pages']]
    references = {p['reference'] for p in used}
    hashes = {p['sha256'] for p in used}
    old_history = {p['reference'] for p in base['frame'] if p['cohort'] == 'historical'}
    archive_path = ROOT/'research/generation/studies/long-prose-v3/selection-audit.tar.gz'
    archive_record = read(archive_path.with_name('selection-audit.json'))
    if digest(archive_path.read_bytes()) != archive_record['sha256']:
        raise ValueError('Selection archive drift')
    with tarfile.open(archive_path) as archive:
        raw = archive.extractfile('selection-audit/eligible.json').read()
        if digest(raw) != archive_record['files']['selection-audit/eligible.json']:
            raise ValueError('Eligible metadata drift')
        pools = json.loads(raw)
    frame = [p for p in base['frame'] if p['cohort'] == 'ptah']
    for pool in pools:
        for entry in pool['pool']:
            source = entry['source']
            if source['reference'] in old_history:
                continue
            frame.append(dict(cohort='historical', repository=source['repository'],
                original_path=source['path'], input_path=source['repository'].replace('/', '__')+'/'+source['path'],
                words=entry['words'], paragraphs=entry['paragraphs'],
                **{k:source[k] for k in ('sha256','bytes','format','snapshot','origin','rights','notices','reference')}))
    for page in frame:
        page['length_stratum'] = 'short' if page['words'] < 800 else 'medium' if page['words'] < 2000 else 'long'
    def rank(page):
        return digest((SEED+'\n'+page['reference']).encode())
    selected, cells, repos = [], [], set()
    for cohort in ('ptah','historical'):
        order = ('long','medium','short') if cohort == 'historical' else ('short','medium','long')
        for length in order:
            pool = [p for p in frame if p['cohort'] == cohort and p['length_stratum'] == length
                    and p['reference'] not in references and p['sha256'] not in hashes]
            eligible = sorted((p for p in pool if cohort != 'historical' or p['repository'] not in repos), key=rank)
            chosen = []
            for page in eligible:
                if page['sha256'] in hashes or (cohort == 'historical' and page['repository'] in repos):
                    continue
                chosen.append(page)
                repos.add(page['repository'])
                hashes.add(page['sha256'])
                if len(chosen) == 1:
                    break
            cells.append(dict(cohort=cohort,length=length,available=len(pool),eligible=len(eligible),
                              requested=1,selected=len(chosen)))
            if len(chosen) != 1:
                raise ValueError('Source cell shortage: '+str(cells[-1]))
            selected.extend(chosen)
    files, pages = {}, []
    for page in selected:
        pid = f'c{len(pages)+1:02}'
        row = dict(page,id=pid,path='sources/'+pid+Path(page['original_path']).suffix,confirmation_rank=rank(page))
        root = args.ptah_sources if page['cohort'] == 'ptah' else args.historical_sources
        data = checked_file(root,page['input_path'],page['sha256'])
        if len(data) != page['bytes']:
            raise ValueError('Source size drift')
        files[row['path']] = data
        if page['cohort'] == 'ptah':
            notices = {'notices/ptah/LICENSE': (ROOT/'e2e/rhetoricdata/LICENSE.ptah').read_bytes()}
        else:
            slug = page['repository'].replace('/','__')
            notices = {'notices/'+slug+'/'+n['path']: checked_file(root,slug+'/'+n['path'],n['sha256'])
                       for n in page['notices']}
        files.update(notices)
        row['retained_notices'] = [dict(path=p,sha256=digest(b)) for p,b in notices.items()]
        pages.append(row)
    output = io.BytesIO()
    with tarfile.open(fileobj=output,mode='w') as archive:
        for path,data in sorted(files.items()):
            info = tarfile.TarInfo(path)
            info.size,info.mode = len(data),0o644
            archive.addfile(info,io.BytesIO(data))
    args.output.mkdir(parents=True)
    packed = gzip.compress(output.getvalue(),mtime=0)
    (args.output/'inputs.tar.gz').write_bytes(packed)
    result = dict(version=1,selection=SEED,excluded_reviewed_references=sorted(references),
                  excluded_prior_historical_study_references=sorted(old_history),
                  previous_manifests={str(p.relative_to(ROOT)):digest(p.read_bytes()) for p in previous},
                  selection_archive_sha256=archive_record['sha256'],eligible_metadata_sha256=digest(raw),
                  inputs_sha256=digest(packed),frame=frame,cells=cells,pages=pages)
    (args.output/'manifest.json').write_text(json.dumps(result,indent=2)+'\n')
    for page in pages:
        print(page['id'],page['cohort'],page['length_stratum'],page['words'],page['reference'])


if __name__ == '__main__':
    main()
