"""Reject altered fold, abstention and selection evidence."""
import copy
import unittest
import check


class EvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.report = check.load(check.STUDY/'predictions.json.gz')
        cls.vectors = check.load(check.STUDY/'embeddings.json.gz')

    def test_retained_report(self):
        check.validate(self.report,self.vectors)

    def test_rejects_mutations(self):
        for mutation in ('selection','missing','fold','duplicate','unavailable'):
            with self.subTest(mutation=mutation):
                report = copy.deepcopy(self.report)
                row = report['models'][0]['predictions'][0]
                if mutation == 'selection':
                    row['selected'] = not row['selected']
                elif mutation == 'missing':
                    report['models'][0]['predictions'].pop()
                elif mutation == 'fold':
                    report['models'][0]['evaluation_fold'] = 1
                elif mutation == 'duplicate':
                    report['models'][0]['predictions'].append(copy.deepcopy(row))
                elif mutation == 'unavailable':
                    row['raw_score'] = None
                    row['selected'] = False
                with self.assertRaises(ValueError):
                    check.validate(report,self.vectors)


if __name__ == '__main__':
    unittest.main()
