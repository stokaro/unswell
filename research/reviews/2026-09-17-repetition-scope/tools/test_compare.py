"""Reject partial repetition credit and concealed diagnostic changes."""
import copy
from pathlib import Path
import sys
import unittest
from unittest.mock import patch

sys.dont_write_bytecode=True
sys.path.insert(0,str(Path(__file__).resolve().parent))
import compare as c

class EvidenceTests(unittest.TestCase):
 @classmethod
 def setUpClass(cls):
  _,cls.pages,_,_,cls.events=c.audit.load(c.DEVELOPMENT)
  cls.report=c.audit.read(c.ROOT/'reports/development/after/technical.json.gz')
  cls.review=c.audit.read(c.ROOT/'dispositions.json')
  cls.row=cls.review['development']['technical']['added'][0]
  cls.finding=cls.report['findings'][cls.row['index']]

 def test_frozen_results_include_gain_and_loss(self):
  r=c.evaluate()
  pages={p['page']:p for p in r['development']['technical']['pages']}
  self.assertEqual((pages['p004']['before'],pages['p004']['after']),(1,0))
  self.assertEqual((pages['p014']['before'],pages['p014']['after']),(0,1))
  self.assertEqual(sum(p['after'] for p in r['confirmation']['technical']['pages']),0)

 def test_reject_missing_repeated_occurrence(self):
  f=copy.deepcopy(self.finding);f['related']=[]
  row=dict(self.row,finding_sha256=c.audit.sha(c.audit.encode(f)))
  with self.assertRaisesRegex(ValueError,'Unsupported credit'):c.validate_claim(f,row,self.pages,self.events)

 def test_reject_shared_subject_paraphrase_credit(self):
  row=dict(self.row,events=['p012-d01'])
  with self.assertRaisesRegex(ValueError,'Unsupported credit'):c.validate_claim(self.finding,row,self.pages,self.events)

 def test_reject_uncertain_credit(self):
  eid=next(k for k,e in self.events.items() if e['kind']=='uncertain')
  row=dict(self.row,events=[eid])
  with self.assertRaisesRegex(ValueError,'Uncertain/control credit'):c.validate_claim(self.finding,row,self.pages,self.events)

 def test_stable_location_cannot_hide_gate_change(self):
  before={'findings':[self.finding]};after=copy.deepcopy(before);after['findings'][0]['gate']='forbid'
  with self.assertRaisesRegex(ValueError,'Retained diagnostic changed'):c.delta(before,after)

 def test_reject_stale_hash(self):
  row=dict(self.row,finding_sha256='0'*64)
  with self.assertRaisesRegex(ValueError,'Finding review drift'):c.validate_claim(self.finding,row,self.pages,self.events)

 def test_removed_actionable_finding_cannot_be_omitted(self):
  review=copy.deepcopy(self.review)
  review['development']['technical']['removed']=[r for r in review['development']['technical']['removed'] if not r['events']]
  original=c.audit.read
  def read(path):return review if path==c.ROOT/'dispositions.json' else original(path)
  with patch.object(c.audit,'read',side_effect=read):
   with self.assertRaisesRegex(ValueError,'Incomplete delta review'):c.evaluate()

if __name__=='__main__':unittest.main()
