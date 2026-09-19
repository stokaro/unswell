"""Independent saved-model logits; never used by the product or Go training."""
import sys,json,hashlib
from pathlib import Path
import numpy as np
import onnxruntime as ort
import tokenizers
root=Path(sys.argv[1]);data=json.loads(Path(sys.argv[2]).read_text())
report=json.loads((root/'report.json').read_text());head=json.loads((root/'head.json').read_text())
index={(r['page'],r['unit']):r['text'] for r in data['rows']}
rows=[p for p in report['predictions'] if p['raw_score'] is not None]
# Predetermined first 32 hash-ordered available held-out targets.
rows=sorted(rows,key=lambda p:hashlib.sha256((p['page']+':'+str(p['unit'])).encode()).hexdigest())[:32]
t=tokenizers.Tokenizer.from_file(str(root/'tokenizer.json'));t.enable_padding()
enc=t.encode_batch([index[(r['page'],r['unit'])] for r in rows])
inputs={k:np.array([getattr(e,a) for e in enc],dtype=np.int64) for k,a in [('input_ids','ids'),('attention_mask','attention_mask'),('token_type_ids','type_ids')]}
o=ort.SessionOptions();o.intra_op_num_threads=2;o.inter_op_num_threads=1
s=ort.InferenceSession(str(root/'selected.onnx'),o,providers=['CPUExecutionProvider'])
h=s.run(None,inputs)[0];mask=inputs['attention_mask'][...,None].astype(np.float32)
v=(h*mask).sum(axis=1)/mask.sum(axis=1);v/=np.linalg.norm(v,axis=1,keepdims=True)
logits=(v@np.array(head['weights'],dtype=np.float32)+np.array(head['bias'],dtype=np.float32)).reshape(-1)
error=float(np.abs(logits-np.array([r['raw_score'] for r in rows])).max())
out={'max_absolute_error':error,'tolerance':1e-4,'passed':error<=1e-4,'onnxruntime':ort.__version__,'numpy':np.__version__,'tokenizers':tokenizers.__version__,'snapshot_sha256':hashlib.sha256((root/'selected.onnx').read_bytes()).hexdigest(),'rows':[{'page':r['page'],'unit':r['unit'],'go_logit':r['raw_score'],'reference_logit':float(x)} for r,x in zip(rows,logits)]}
print(json.dumps(out));assert out['passed'],'Saved Go weights do not reproduce in ONNX Runtime'
