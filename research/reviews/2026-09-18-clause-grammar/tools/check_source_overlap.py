import argparse,hashlib,json,pathlib,tarfile
root=pathlib.Path(__file__).resolve().parent.parent
parser=argparse.ArgumentParser(description='Audit extracted source-only packets for exact prior prose. No diagnostic input is used.')
parser.add_argument('--packets',type=pathlib.Path,required=True,help='Directory containing cNN.json extractor packets and archived source basenames')
scratch=parser.parse_args().packets
def norm(t):return ' '.join(t.split())
manifest=json.loads((root/'confirmation/manifest.json').read_text())
prior={}
for m in sorted(list(root.parent.glob('*/confirmation/manifest.json'))+[root.parent/'2026-09-17-full-page-recall/manifest.json']):
    if m.parent.parent==root:continue
    arc=m.parent/'inputs.tar.gz'
    if not arc.exists():continue
    with tarfile.open(arc) as t:
        files={f.name:t.extractfile(f).read() for f in t if f.isfile()}
    for p in json.loads(m.read_text())['pages']:
        if p['reference'] not in manifest['excluded_reviewed_references']:continue
        prior[p['reference']]=(norm(files[p['path']].decode()),str(m.relative_to(root.parent)))
matched=[]
for p in manifest['pages']:
    raw=(scratch/pathlib.Path(p['path']).name).read_bytes()
    assert hashlib.sha256(raw).hexdigest()==p['sha256'], 'Source hash differs: '+p['id']
    packet=json.loads((scratch/(p['id']+'.json')).read_text())
    for b in packet['blocks']:
        if b.get('excluded'):continue
        lo,hi=b['span']['start'],b['span']['end'];text=raw[lo:hi].decode();n=norm(text)
        if len(n.split())<12:continue
        matches=[dict(reference=ref,manifest=mf) for ref,(s,mf) in prior.items() if n in s]
        if matches:matched.append(dict(page=p['id'],block=b['id'],start=lo,end=hi,quote=text,matches=matches))
result=dict(version=1,phase='before_runtime_changes_and_confirmation_diagnostics',method='Exact whole extracted block substring after whitespace normalization, minimum 12 whitespace tokens; source bytes only. This is a narrow overlap audit, not semantic deduplication.',prior_references_scanned=len(prior),known_partially_exposed_pages=sorted({m['page'] for m in matched}),matches=matched,limitation='No matching block does not prove independent provenance or absence of paraphrased, shorter, or partial overlap.')

print(json.dumps(result,indent=2))
