"""Offline source and response-contract tests; no model calls."""
import copy
import unittest

import diagnosis as d


def packet(parts):
    raw = b''; units = []
    for i, value in enumerate(parts):
        start = len(raw); raw += value.encode('utf-8')
        units.append({'id': f'u{i:04}', 'span': {'start': start, 'end': len(raw)}})
        raw += b'\n\n'
    return {'page': 'page-test', 'source_sha256': d.digest(raw),
            'source': raw.decode('utf-8'), 'format': 'markdown',
            'units': units, 'protected': []}


def response(request):
    return {'request': request['request'], 'page': request['page'],
            'assessments': [{'unit': unit, 'technical_payload': 'States the condition.',
                             'burden_decision': 'no_burden',
                             'burden_reason': 'The condition is necessary.',
                             'finding_ids': []} for unit in request['focus']],
            'findings': []}


def finding(unit, quote, occurrence=1):
    return {'id': 'f1', 'category': 'empty_framing',
            'targets': [{'unit': unit, 'quote': quote, 'occurrence': occurrence}],
            'diagnostic': 'The clause announces the section.',
            'why_revision_needed': 'The heading already supplies the same navigation.',
            'meaning_to_preserve': 'Keep the condition and named operation.'}


def diagnosed(request, item):
    result = response(request); result['findings'] = [item]
    order = {u['id']: i for i, u in enumerate(request['units'])}
    owner = min((t['unit'] for t in item['targets']), key=order.__getitem__)
    for row in result['assessments']:
        if row['unit'] == owner:
            row.update(burden_decision='diagnose', finding_ids=[item['id']])
    return result


class DiagnosisTests(unittest.TestCase):
    def test_utf8_source_and_raw_slices_are_unchanged(self):
        source = packet(['\ufeffCafé uses 3 ms.\r\nKeep `x`.', '**Label**: &amp; value.'])
        before = copy.deepcopy(source)
        request = d.OLD.requests(source)[0]
        self.assertEqual(source, before)
        self.assertEqual(request['units'][0]['raw'], '\ufeffCafé uses 3 ms.\r\nKeep `x`.')
        self.assertEqual(d.OLD.verify_request(request), source)

    def test_tampering_or_invented_metadata_is_rejected(self):
        request = d.OLD.requests(packet(['**Label** value.']))[0]
        request['units'][0]['raw'] = 'Label value.'
        with self.assertRaisesRegex(ValueError, 'Raw unit slice changed'):
            d.validate(request, response(request))
        source = packet(['Necessary condition.']); source['labels'] = ['private']
        with self.assertRaisesRegex(ValueError, 'Unexpected object fields'):
            d.OLD.requests(source)

    def test_every_focus_unit_is_assessed_once(self):
        requests = d.OLD.requests(packet([f'Condition {i}.' for i in range(129)]))
        self.assertEqual([len(r['focus']) for r in requests], [64, 64, 1])
        self.assertEqual([u for r in requests for u in r['focus']],
                         [u['id'] for u in requests[0]['units']])
        assembled = d.assemble(requests, [response(r) for r in requests])
        self.assertEqual(len(assembled['coverage']), 129)
        for changed in [[], [response(requests[0])['assessments'][0]] * 2]:
            result = response(requests[0]); result['assessments'] = changed
            with self.assertRaises(ValueError): d.validate(requests[0], result)

    def test_partial_page_does_not_assemble(self):
        requests = d.OLD.requests(packet(['Necessary condition.'] * 65))
        with self.assertRaises(ValueError): d.assemble(requests[:1], [response(requests[0])])
        with self.assertRaisesRegex(ValueError, 'Incomplete page results'):
            d.assemble(requests, [response(requests[0])])

    def test_burden_decision_and_diagnosis_must_agree(self):
        request = d.OLD.requests(packet(['An announcement.']))[0]
        for field, value in [('technical_payload', ''), ('burden_reason', ' '),
                             ('burden_decision', 'retain'), ('finding_ids', ['invented']),
                             ('burden_decision', 'diagnose')]:
            with self.subTest(field=field):
                result = response(request); result['assessments'][0][field] = value
                with self.assertRaises(ValueError): d.validate(request, result)
        result = diagnosed(request, finding('u0000', 'An announcement.'))
        self.assertEqual(len(d.validate(request, result)['findings']), 1)
        result['assessments'][0]['burden_decision'] = 'uncertain'
        with self.assertRaises(ValueError): d.validate(request, result)

    def test_repair_fields_are_not_allowed(self):
        request = d.OLD.requests(packet(['An announcement.']))[0]
        result = diagnosed(request, finding('u0000', 'An announcement.'))
        result['findings'][0]['suggestion'] = 'Delete it.'
        with self.assertRaisesRegex(ValueError, 'Unexpected object fields'):
            d.validate(request, result)

    def test_exact_occurrence_and_protected_prose(self):
        request = d.OLD.requests(packet(['écho echo echo.']))[0]
        result = d.validate(request, diagnosed(request, finding('u0000', 'echo', 2)))
        self.assertEqual(result['findings'][0]['targets'][0]['span'], {'start': 11, 'end': 15})
        source = packet(['code prose']); source['protected'] = [{'span': {'start': 0, 'end': 4}}]
        request = d.OLD.requests(source)[0]
        with self.assertRaisesRegex(ValueError, 'eligible prose'):
            d.validate(request, diagnosed(request, finding('u0000', 'code')))
        with self.assertRaisesRegex(ValueError, 'Quotation is not'):
            d.validate(request, diagnosed(request, finding('u0000', 'invented')))

    def test_cross_window_finding_has_one_owner(self):
        requests = d.OLD.requests(packet([f'Paragraph {i}.' for i in range(65)]))
        item = finding('u0000', 'Paragraph 0.')
        item['targets'].append({'unit': 'u0064', 'quote': 'Paragraph 64.', 'occurrence': 1})
        first = diagnosed(requests[0], item)
        self.assertEqual(len(d.validate(requests[0], first)['findings']), 1)
        with self.assertRaisesRegex(ValueError, 'another focus window'):
            d.validate(requests[1], diagnosed(requests[1], item))
        assembled = d.assemble(requests, [first, response(requests[1])])
        self.assertEqual(assembled['findings'][0]['id'], 'page-test-part-001/f1')


if __name__ == '__main__':
    unittest.main()
