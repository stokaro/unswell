"""Reject altered training, checkpoint and full-stream evidence."""
import copy
import unittest
import check


class EvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.report = check.load(check.STUDY/'predictions.json.gz')
        cls.prior,cls.metadata = check.source_metadata()

    def test_retained_report(self):
        check.validate(self.report,self.metadata,self.prior)

    def test_rejects_mutations(self):
        cases = ('selection','missing','fold','duplicate','unavailable','label',
                 'constant','loss','development','epoch','groups','parity','reload','threshold')
        for mutation in cases:
            with self.subTest(mutation=mutation):
                report = copy.deepcopy(self.report)
                model = next(m for m in report['models'] if m['kind'] == 'FT')
                row = model['predictions'][0]
                if mutation == 'selection':
                    row['selected'] = not row['selected']
                elif mutation == 'missing':
                    model['predictions'].pop()
                elif mutation == 'fold':
                    model['evaluation_fold'] = 1
                elif mutation == 'duplicate':
                    model['predictions'].append(copy.deepcopy(row))
                elif mutation == 'unavailable':
                    row['raw_score'] = None
                    row['selected'] = False
                elif mutation == 'label':
                    row['label'] = 1 if row['label'] != 1 else 0
                elif mutation == 'constant':
                    model['trainable_encoder_variables'][0] = '|Constant_11_output_0'
                elif mutation == 'loss':
                    model['epoch_mean_losses'].pop()
                elif mutation == 'development':
                    model['checkpoints'][0]['development_predictions'].pop()
                elif mutation == 'epoch':
                    model['selected_epoch'] = 2
                elif mutation == 'groups':
                    model['training_groups'].pop()
                elif mutation == 'parity':
                    model['initial_max_score_error'] = .1
                elif mutation == 'reload':
                    model['reload_max_score_error'] = float('nan')
                elif mutation == 'threshold':
                    model['checkpoints'][0]['operating_point']['development_true'] += 1
                with self.assertRaises(ValueError):
                    check.validate(report,self.metadata,self.prior)


if __name__ == '__main__':
    unittest.main()
