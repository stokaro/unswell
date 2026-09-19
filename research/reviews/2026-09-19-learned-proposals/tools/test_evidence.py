"""Evidence validation failures and conservative proposal accounting."""
import copy
import unittest

from check_evidence import validate_model
from summarize import metrics, retrieval
from rule_reference import reference


def fixture():
    pages = [dict(id=str(i),fold=i,group=str(i),cohort='fixture',units=[dict(label=1),dict(label=0)]) for i in range(5)]
    model = dict(evaluation_fold=0,features=['length'],training_groups=['2','3','4'],training_rows=6,
                 parameters=dict(Means=[0],Scales=[1],Weights=[1],Intercept=0),
                 operating_point=dict(available=True,threshold=1,development_positives=1,development_negatives=1,
                                      development_true=1,development_false=0),
                 predictions=[dict(page='0',unit=0,label=1,cohort='fixture',raw_score=2,selected=True),
                              dict(page='0',unit=1,label=0,cohort='fixture',raw_score=0,selected=False)])
    return model,dict(pages=pages)


class EvidenceTests(unittest.TestCase):
    def test_valid_fixture(self):
        validate_model(*fixture())

    def test_rejects_tampering(self):
        for kind in ['group','label','missing','duplicate','threshold','nonfinite','constraint']:
            with self.subTest(kind=kind):
                model,data = fixture()
                if kind == 'group': model['training_groups'].append('0')
                if kind == 'label': model['predictions'][0]['label'] = 0
                if kind == 'missing': model['predictions'].pop()
                if kind == 'duplicate': model['predictions'].append(copy.deepcopy(model['predictions'][0]))
                if kind == 'threshold': model['operating_point']['threshold'] = 3
                if kind == 'nonfinite': model['predictions'][0]['raw_score'] = float('nan')
                if kind == 'constraint': model['operating_point']['development_false'] = 1
                with self.assertRaises(ValueError): validate_model(model,data)

    def test_unlabeled_selection_is_not_a_false_positive(self):
        result = metrics([dict(label=None,selected=True)])
        self.assertEqual(result['false_positive'],0)
        self.assertEqual(result['unlabeled_selected'],1)
        self.assertIsNone(result['precision'])
        self.assertIsNone(result['recall'])

    def test_every_target_piece_must_be_selected(self):
        events = [dict(kind='defects',page='page',id='event',category='wordiness',cohort='fixture',prepared_pieces=[0,1])]
        self.assertFalse(retrieval(events,{('page',0)})[0]['retrieved'])
        self.assertTrue(retrieval(events,{('page',0),('page',1)})[0]['retrieved'])
        events[0]['prepared_pieces'] = []
        self.assertFalse(retrieval(events,set())[0]['retrieved'])

    def test_rule_reference_requires_bound_event_denominators(self):
        with self.assertRaises(ValueError):
            reference({'events':[]},{})


if __name__ == '__main__':
    unittest.main()
