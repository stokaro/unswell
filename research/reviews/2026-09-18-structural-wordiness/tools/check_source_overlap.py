"""Audit exact complete prose blocks against previously exposed source bytes."""
import argparse,hashlib,json,pathlib,tarfile
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--study',type=pathlib.Path,required=True)
parser.add_argument('--prior-reviews',type=pathlib.Path,required=True)
parser.add_argument('--packets',type=pathlib.Path,required=True)
args=parser.parse_args()
def norm(text):return ' '.join(text.split())
manifest=json.loads((args.study/'confirmation/manifest.json').read_text())
prior={}
for m in sorted(list(args.prior_reviews.glob('*/confirmation/manifest.json'))+[args.prior_reviews/'2026-09-17-full-page-recall/manifest.json']):
 archive=m.parent/'inputs.tar.gz'
 if not archive.exists():continue
 with tarfile.open(archive) as t:files={f.name:t.extractfile(f).read() for f in t if f.isfile()}
 for p in json.loads(m.read_text())['pages']:
  if p['reference'] in manifest['excluded_reviewed_references']:
   prior[p['reference']]=(norm(files[p['path']].decode()),str(m.relative_to(args.prior_reviews)))
matched=[]
for p in manifest['pages']:
 raw=(args.packets/pathlib.Path(p['path']).name).read_bytes()
 assert hashlib.sha256(raw).hexdigest()==p['sha256']
 packet=json.loads((args.packets/(p['id']+'.json')).read_text())
 for block in packet['blocks']:
  if block.get('excluded'):continue
  lo,hi=block['span']['start'],block['span']['end'];text=raw[lo:hi].decode();n=norm(text)
  if len(n.split())<12:continue
  matches=[dict(reference=ref,manifest=mf) for ref,(source,mf) in prior.items() if n in source]
  if matches:matched.append(dict(page=p['id'],block=block['id'],start=lo,end=hi,quote=text,matches=matches))
print(json.dumps(dict(version=1,phase='before_runtime_changes_and_confirmation_diagnostics',
 method='Exact whole extracted block substring after whitespace normalization; minimum 12 whitespace tokens.',
 prior_references_scanned=len(prior),known_partially_exposed_pages=sorted({m['page'] for m in matched}),matches=matched,
 limitation='This does not detect paraphrased, shorter, or partial overlap and does not establish independent authorship.'),indent=2))
