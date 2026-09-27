#!/usr/bin/env python3
"""Replay frozen full-page inputs with two explicit offline CLI binaries."""
import argparse,gzip,hashlib,json,platform,re,subprocess,tarfile,tempfile,time
from pathlib import Path
parser=argparse.ArgumentParser()
parser.add_argument('--root',type=Path,required=True)
parser.add_argument('--before',type=Path,required=True)
parser.add_argument('--after',type=Path,required=True)
parser.add_argument('--output',type=Path,required=True)
a=parser.parse_args();root=a.root.resolve();out=a.output.resolve();out.mkdir(parents=True,exist_ok=False)
def sha(data):return hashlib.sha256(data).hexdigest()
for n,h in json.loads((root/'input-freeze.json').read_text()).items():assert sha((root/n).read_bytes())==h,n
review=json.loads((root/'source-review.json').read_text());sources={};meta={}
with tarfile.open(root/'inputs.tar.gz') as arc:
 for row in review['confirmation']+[review['development']]:
  path=row.get('archive_path',row.get('path'));data=arc.extractfile(path).read();assert sha(data)==row['sha256'];sources[path]=data;meta[path]=row
reports={};runs={}
with tempfile.TemporaryDirectory(prefix='unswell-hooks-') as d:
 work=Path(d)
 for n,data in sources.items():
  p=work/n;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(data)
 for name,binary in [('before',a.before.resolve()),('after',a.after.resolve())]:
  for profile in ['technical','strict']:
   report=work/'report.json';cmd=[str(binary),'check','--profile',profile,'--include-source','--report','json:report.json',*sorted(sources)]
   cmd=['/usr/bin/time','-l' if platform.system()=='Darwin' else '-v',*cmd]
   start=time.monotonic();r=subprocess.run(cmd,cwd=work,capture_output=True,timeout=180);wall=time.monotonic()-start
   assert r.returncode in (0,1),r.stderr.decode()
   v=json.loads(report.read_bytes());assert v['status']=='complete' and v['manifest']['complete'] and not v['errors'] and not v.get('abstentions')
   assert [x['name'] for x in v['documents']]==sorted(sources)
   for x in v['documents']:assert x['source'].encode()==sources[x['name']]
   payload=gzip.compress(report.read_bytes(),mtime=0);file=f'{name}-{profile}.json.gz';(out/file).write_bytes(payload);reports[name,profile]=v
   stderr=r.stderr.decode()
   if platform.system()=='Darwin':rss=int(re.search(r'(\d+)\s+maximum resident set size',stderr)[1])
   else:rss=int(re.search(r'Maximum resident set size \(kbytes\):\s*(\d+)',stderr)[1])*1024
   runs[file]={'sha256':sha(payload),'tool_commit':v['manifest']['tool_commit'],'binary_sha256':sha(binary.read_bytes()),'wall_seconds':wall,'peak_rss_bytes':rss,'exit':r.returncode,'findings':len(v['findings'])}
(out/'runs.json').write_text(json.dumps(runs,indent=2)+'\n')
def key(f):return f['rule_id'],f['primary']['path'],f['primary']['span']['start'],f['primary']['span']['end']
result={'reviewer':review['reviewer'],'host':platform.platform(),'runs':runs,'profiles':{}}
for profile in ['technical','strict']:
 old,new=reports['before',profile],reports['after',profile];left={key(f):f for f in old['findings']};right={key(f):f for f in new['findings']}
 removed=[f for k,f in left.items() if k not in right];added=[f for k,f in right.items() if k not in left]
 assert not removed
 for k in left:assert left[k]==right[k],('Existing finding changed',k)
 assert len(added)==1,added
 f=added[0];assert f['rule_id']=='filler.unnamed-numerical-choice';assert f['primary']['path']==review['development']['archive_path']
 event=review['development']['events'][0];assert f['primary']['span']=={'start':event['start'],'end':event['end']-1};assert f['primary']['snippet']==event['quote'][:-1]
 for previous,current in zip(old['documents'],new['documents'],strict=True):
  assert {k:v for k,v in previous.items() if k not in ('config_hash','paragraph_maximum')}=={k:v for k,v in current.items() if k not in ('config_hash','paragraph_maximum')}
  if current['name']==review['development']['archive_path']:
   assert previous['paragraph_maximum']==0 and current['paragraph_maximum']==12
  else:assert previous['paragraph_maximum']==current['paragraph_maximum']
 assert old['gate']==new['gate']
 for row in review['confirmation']:
  for control in row['controls']:assert sources[row['path']][control['start']:control['end']].decode()==control['quote']
 result['profiles'][profile]={'before':len(left),'after':len(right),'unchanged':len(left),'removed':removed,'added':added,'dispositions':[{'key':key(f),'accepted':True,'reason':event['reason']}],'exposed_development_recall':{'before':0,'after':1,'denominator':1},'confirmation':{'pages':2,'positive_events':0,'added_findings':0,'controls':6,'recall':None,'reason':'No positive events; sensitivity is unmeasured.'},'gate':new['gate']}
(out/'summary.json').write_text(json.dumps(result,indent=2)+'\n')
for p,x in result['profiles'].items():print(p,x['before'],x['after'],x['confirmation'])
