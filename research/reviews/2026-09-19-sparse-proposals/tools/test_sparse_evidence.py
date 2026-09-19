"""Reject capacity, convergence and same-capacity comparison tampering."""
import copy
import unittest

from check_sparse import validate_sparse, validate_report
from results import compare


def sparse_fixture():
    pages = [dict(id=str(i),fold=i,group=str(i),cohort='fixture',units=[dict(label=1),dict(label=0)]) for i in range(5)]
    model = dict(evaluation_fold=0,features=['length'],training_groups=['2','3','4'],training_rows=6,
                 parameters=dict(Means=[0],Scales=[1],Weights=[1],Intercept=0),
                 operating_point=dict(available=True,threshold=1,development_positives=1,development_negatives=1,
                                      development_true=1,development_false=0),
                 predictions=[dict(page='0',unit=0,label=1,cohort='fixture',raw_score=2,selected=True),
                              dict(page='0',unit=1,label=0,cohort='fixture',raw_score=0,selected=False)])
    data = dict(pages=pages)
    model.update(kind='SW128',options=dict(L2=0.01,Tolerance=1e-8,MaxIterations=500,MaxOperations=3000000000),
                 iterations=20,operations=1000,gradient_norm=1e-9)
    return model,data


class SparseEvidenceTests(unittest.TestCase):
    def test_accepts_wider_contract(self):
        model,data = sparse_fixture()
        model['kind'] = 'SW1024'
        model['features'] = [str(i) for i in range(1024)]
        model['parameters'].update(Means=[0]*1024,Scales=[1]*1024,Weights=[0]*1024)
        validate_sparse(model,data)
        model['kind'] = 'SW128'
        with self.assertRaisesRegex(ValueError,'feature contract'):
            validate_sparse(model,data)

    def test_rejects_unconverged_or_changed_optimizer(self):
        for kind in ('unknown','gradient','nan','iterations','operations','options'):
            with self.subTest(kind=kind):
                model,data = sparse_fixture()
                if kind=='unknown': model['kind']='SW4096'
                if kind=='gradient': model['gradient_norm']=1e-6
                if kind=='nan': model['gradient_norm']=float('nan')
                if kind=='iterations': model['iterations']=500
                if kind=='operations': model['operations']=3000000001
                if kind=='options': model['options']['L2']=0
                with self.assertRaises(ValueError): validate_sparse(model,data)

    def test_rejects_missing_fit(self):
        model,data = sparse_fixture()
        report = dict(version=2,model_algorithm='unswell-research-sparse-logistic-lbfgs-v1',models=[model],rows=10)
        with self.assertRaisesRegex(ValueError,'Missing or duplicate'): validate_report(report,data)

    def test_comparison_counts_changed_decisions(self):
        model,_ = sparse_fixture()
        current,prior = [],[]
        for fold in range(5):
            new = copy.deepcopy(model)
            new['evaluation_fold'] = fold
            old = copy.deepcopy(new)
            old['kind'] = 'SW'
            current.append(new)
            prior.append(old)
        current[0]['predictions'][0]['selected']=False
        self.assertEqual(compare(dict(models=current),dict(models=prior))['changed_selections'],1)
        current[0]['features']=['changed']
        with self.assertRaisesRegex(ValueError,'feature identity'):
            compare(dict(models=current),dict(models=prior))


if __name__ == '__main__':
    unittest.main()
