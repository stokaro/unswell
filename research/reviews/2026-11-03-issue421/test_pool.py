"""Negative tests for candidate-pool and separated-stream joins."""

import copy
import unittest

import pool


class PoolContractTest(unittest.TestCase):
    def test_oracle_grade_cannot_be_exported_as_policy(self):
        output = {'kind': 'retrospective_event_loss_ledger',
                  'inference_selector_trained_or_evaluated': False,
                  'retrospective_grades_are_not_inference_inputs': True}
        pool.assert_research_only(output)
        output['selector'] = {'features': ['frozen_grade', 'event_id']}
        with self.assertRaisesRegex(ValueError, 'leaked'):
            pool.assert_research_only(output)

    def test_incompatible_source_join_and_unsupported_full_rejected(self):
        pages = {f'page-{number:03}': {
            'path': f'whole/p{number:03}.md',
            'source_sha256': f'hash-{number:03}'} for number in range(1, 37)}
        manifest = {'pages': [{'path': item['path'],
                               'source_sha256': item['source_sha256']}
                              for item in pages.values()]}
        original = {'reference': {'events': {}}}
        witnesses = {}
        for part, count in pool.DENOMINATORS.items():
            for number in range(count):
                event_id = f'{part}/event-{number:03}'
                original['reference']['events'][event_id] = {
                    'path': 'whole/p001.md'}
                witnesses[event_id] = {'status': 'missed', 'full_groups': []}
        first = 'whole/event-000'
        witnesses[first] = {'status': 'full',
                            'full_groups': [['generator/page-002/f01']]}
        selected = ['generator/page-002/f01'] + [f'noise-{number}'
                                                for number in range(2846)]
        private = {'original_scope': {'pages': 36, 'units': 4549,
                                      'judgments': 804, 'events': 123},
                   'all_selected_actual_aliases': selected,
                   'complete_witnesses': witnesses,
                   'cohorts': {}}
        public = {'original_scope': private['original_scope'], 'cohorts': {}}
        for part, count in pool.DENOMINATORS.items():
            delivered = 2847 if part == 'whole' else 0
            full = 1 if part == 'whole' else 0
            private['cohorts'][part] = {'events': count, 'full': full,
                                        'delivered': delivered,
                                        'accepted': delivered,
                                        'states': {'accepted': delivered}}
            public['cohorts'][part] = {'events': count, 'full': full,
                                       'delivered': delivered,
                                       'accepted': delivered,
                                       'uncertain': 0, 'rejected': 0}
        with self.assertRaisesRegex(ValueError, 'another source'):
            pool.check_separated(private, public, manifest, original, pages, {})
        unsupported = copy.deepcopy(private)
        unsupported['complete_witnesses'][first]['full_groups'] = []
        with self.assertRaisesRegex(ValueError, 'Unsupported'):
            pool.check_separated(unsupported, public, manifest, original,
                                 pages, {})


if __name__ == '__main__':
    unittest.main()
