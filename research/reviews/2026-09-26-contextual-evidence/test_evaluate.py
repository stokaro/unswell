import copy
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

from evaluate import collect, score, verify_evaluation_freeze


class ScoringTests(unittest.TestCase):
    def setUp(self):
        self.events = {}; self.baseline = {}; self.predictions = {}
        self.review = {'reviewer': 'test-only synthetic reviewer', 'findings': {}, 'events': {}}
        self.pages = set()
        for cohort in ['whole', 'context', 'confirmation']:
            event = cohort + '/event'; finding = cohort + '/finding'; page = cohort + '/page'
            self.pages.add(page)
            self.events[event] = {'path': page, 'targets': [{'start': 0, 'end': 20}]}
            self.baseline[event] = {'after': {'status': 'missed'}}
            self.predictions[finding] = {'path': page, 'targets': [{'span': {'start': 0, 'end': 20}}]}
            self.review['findings'][finding] = {
                'status': 'accepted', 'rationale': 'Test fixture only.',
                'suggestion_safety': {'status': 'safe', 'rationale': 'Synthetic fixture only.'}}
            self.review['events'][event] = {'status': 'full', 'findings': [finding], 'rationale': 'Test fixture only.'}

    def result(self):
        return score(self.events, self.baseline, self.predictions, self.pages, self.review, self.pages)

    def test_pass_is_development_only(self):
        result = self.result()
        self.assertTrue(result['development_screen_passed'])
        self.assertFalse(result['product_qualified'])

    def test_wrong_model_or_effort_is_rejected_before_collection(self):
        for model, effort in [('different-model', 'medium'), ('gpt-6-astra', 'low')]:
            with self.subTest(model=model, effort=effort):
                with patch('evaluate.verify_freeze'), patch('evaluate.read', side_effect=[
                    {'pages': []}, {'requested_model': model, 'effort': effort},
                    {'model': 'gpt-6-astra', 'effort': 'medium'},
                ]):
                    with self.assertRaisesRegex(ValueError, 'Model or effort'):
                        collect(Path('/packet'), Path('/run'))

    def test_evaluation_freeze_rejects_machine_specific_or_escaping_paths(self):
        with tempfile.TemporaryDirectory() as directory:
            packet = Path(directory); (packet / 'freeze.json').write_text('{}')
            for name in ['/old/worktree/annotations.json', '../annotations.json']:
                frozen = {'packet_freeze_sha256': 'test', 'files': {name: 'test'}}
                with self.subTest(name=name), patch('evaluate.read', return_value=frozen), \
                        patch('evaluate.digest', return_value='test'):
                    with self.assertRaisesRegex(ValueError, 'repository-relative'):
                        verify_evaluation_freeze(packet)

    def test_missing_event_or_finding_review_is_rejected(self):
        for key in ['events', 'findings']:
            review = copy.deepcopy(self.review); review[key].pop(next(iter(review[key])))
            with self.assertRaises(ValueError):
                score(self.events, self.baseline, self.predictions, self.pages, review, self.pages)

    def test_uncertain_is_not_removed_from_precision(self):
        self.review['findings']['whole/finding']['status'] = 'uncertain'
        self.review['events']['whole/event'].update(status='missed', findings=[])
        result = self.result()
        self.assertEqual(result['cohorts']['whole']['findings'], 1)
        self.assertEqual(result['cohorts']['whole']['accepted_fraction'], 0)
        self.assertFalse(result['development_screen_passed'])

    def test_partial_and_baseline_union_are_distinct(self):
        self.review['events']['whole/event']['status'] = 'partial'
        self.baseline['whole/event']['after']['status'] = 'full'
        result = self.result()['cohorts']['whole']
        self.assertEqual(result['full'], 0)
        self.assertEqual(result['baseline_union_full'], 1)

    def test_correct_diagnosis_does_not_hide_unsafe_suggestion(self):
        self.review['findings']['whole/finding']['suggestion_safety'].update(
            status='unsafe', rationale='Test repair removes a necessary condition.')
        result = self.result()['cohorts']['whole']
        self.assertEqual(result['accepted_fraction'], 1)
        self.assertEqual(result['suggestion_safety']['unsafe'], 1)
        self.assertEqual(result['accepted_diagnostics_with_unsafe_suggestions'], 1)

    def test_missing_or_unreviewed_safety_is_rejected(self):
        for value in [{}, {'status': 'unreviewed', 'rationale': 'Not examined.'},
                      {'status': 'safe', 'rationale': ''}]:
            self.review['findings']['whole/finding']['suggestion_safety'] = value
            with self.subTest(value=value), self.assertRaisesRegex(ValueError, 'safety judgment'):
                self.result()

    def test_unavailable_page_stays_in_denominator(self):
        result = score(self.events, self.baseline, {}, set(),
                       {'reviewer': 'synthetic test', 'findings': {}, 'events': {
                           k: {'status': 'missed', 'findings': [], 'rationale': 'Unavailable test input.'}
                           for k in self.events}}, self.pages)
        self.assertEqual(result['cohorts']['whole']['events'], 1)
        self.assertEqual(result['cohorts']['whole']['valid_pages'], 0)
        self.assertFalse(result['development_screen_passed'])

    def test_rejected_wrong_page_or_nonoverlapping_support_fails(self):
        for kind in ['rejected', 'wrong_page', 'wrong_span']:
            predictions = copy.deepcopy(self.predictions); review = copy.deepcopy(self.review)
            if kind == 'rejected': review['findings']['whole/finding']['status'] = 'rejected'
            if kind == 'wrong_page': predictions['whole/finding']['path'] = 'elsewhere'
            if kind == 'wrong_span': predictions['whole/finding']['targets'][0]['span'] = {'start': 50, 'end': 60}
            with self.subTest(kind=kind):
                with self.assertRaises(ValueError):
                    score(self.events, self.baseline, predictions, self.pages, review, self.pages)


if __name__ == '__main__':
    unittest.main()
