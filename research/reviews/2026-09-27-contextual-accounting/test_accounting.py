"""Contract tests, not evidence of improved editorial detection."""
import copy
import json
from pathlib import Path
import tarfile
import unittest

import accounting as a


def packet(parts):
    raw = b''; units = []
    for i, text in enumerate(parts):
        start = len(raw); raw += text.encode()
        units.append({'id': f'u{i:04}', 'span': {'start': start, 'end': len(raw)}})
        raw += b'\n\n'
    return {'page': 'page-test', 'source_sha256': a.digest(raw), 'source': raw.decode(),
            'format': 'markdown', 'units': units, 'protected': []}


def response(request):
    return {'request': request['request'], 'page': request['page'],
            'assessments': [{'unit': unit, 'decision': 'retain',
                             'contribution': 'States the configured timeout.',
                             'reason': 'The timeout is needed for this operation.',
                             'finding_ids': []} for unit in request['focus']], 'findings': []}


def finding(unit, quote, occurrence=1):
    return {'id': 'f1', 'category': 'empty_framing',
            'targets': [{'unit': unit, 'quote': quote, 'occurrence': occurrence}],
            'diagnostic': 'The framing repeats the section heading.',
            'why_revision_needed': 'No extra scope or instruction is supplied.',
            'suggestion': 'Remove only the repeated framing.',
            'meaning_preservation': 'Keep the operation and its conditions.'}


def revision(request, item):
    result = response(request); result['findings'] = [item]
    ids = [u['id'] for u in request['units']]
    owner = min((t['unit'] for t in item['targets']), key=ids.index)
    for row in result['assessments']:
        if row['unit'] == owner:
            row.update(decision='revise', finding_ids=[item['id']])
    return result


class AccountingTests(unittest.TestCase):
    def test_raw_utf8_crlf_bom_and_no_mutation(self):
        p = packet(['\ufeffCafé uses 3 ms.\r\nKeep `x`.', '**Label**: &amp; value.'])
        before = copy.deepcopy(p); request = a.requests(p)[0]
        self.assertEqual(p, before)
        self.assertEqual(request['units'][0]['raw'], '\ufeffCafé uses 3 ms.\r\nKeep `x`.')
        self.assertEqual(a.verify_request(request), p)

    def test_raw_slice_tampering_rejected(self):
        request = a.requests(packet(['**Label** value.']))[0]
        request['units'][0]['raw'] = 'Label value.'
        with self.assertRaisesRegex(ValueError, 'Raw unit slice changed'):
            a.validate(request, response(request))

    def test_split_unicode_and_overlapping_ranges_rejected(self):
        p = packet(['é', 'word']); p['units'][0]['span']['end'] = 1
        with self.assertRaises(UnicodeDecodeError): a.requests(p)
        p = packet(['first', 'second']); p['units'][1]['span']['start'] = 0
        with self.assertRaisesRegex(ValueError, 'unit range'): a.requests(p)

    def test_source_hash_and_unexpected_metadata_rejected(self):
        p = packet(['Keep this.']); p['source'] += 'x'
        with self.assertRaisesRegex(ValueError, 'Source identity'): a.requests(p)
        p = packet(['Keep this.']); p['labels'] = ['bad']
        with self.assertRaisesRegex(ValueError, 'Unexpected object fields'): a.requests(p)

    def test_all_windows_cover_once_without_dropping_tail(self):
        p = packet([f'Keep condition {i}.' for i in range(129)])
        requests = a.requests(p)
        self.assertEqual([len(r['focus']) for r in requests], [64, 64, 1])
        self.assertEqual([u for r in requests for u in r['focus']], [u['id'] for u in p['units']])
        result = a.assemble(requests, [response(r) for r in requests])
        self.assertEqual(len(result['coverage']), 129)
        self.assertEqual(len(a.validate(requests[0], response(requests[0]))['assessments']), 64)

    def test_partial_or_duplicate_window_is_not_complete_page(self):
        requests = a.requests(packet(['Keep condition.'] * 65))
        for changed in [requests[:1], [requests[0], requests[0]], requests[::-1]]:
            with self.subTest(changed=changed[0]['request']):
                with self.assertRaises(ValueError): a.assemble(changed, [response(r) for r in changed])
        with self.assertRaisesRegex(ValueError, 'Incomplete page results'):
            a.assemble(requests, [response(requests[0])])

    def test_decision_requires_explanation_and_linked_diagnostic(self):
        request = a.requests(packet(['An announcement.']))[0]
        for field, value in [('contribution', ''), ('reason', ' '), ('decision', 'reviewed'),
                             ('finding_ids', ['invented']), ('decision', 'revise')]:
            with self.subTest(field=field, value=value):
                result = response(request); result['assessments'][0][field] = value
                with self.assertRaises(ValueError): a.validate(request, result)
        result = revision(request, finding('u0000', 'An announcement.'))
        self.assertEqual(len(a.validate(request, result)['findings']), 1)

    def test_no_fake_inventory_completion_or_uncertainty_with_finding(self):
        request = a.requests(packet(['First.', 'Second.']))[0]
        for rows in [[], [response(request)['assessments'][0]] * 2]:
            result = response(request); result['assessments'] = rows
            with self.assertRaises(ValueError): a.validate(request, result)
        result = revision(request, finding('u0000', 'First.'))
        result['assessments'][0]['decision'] = 'uncertain'
        with self.assertRaises(ValueError): a.validate(request, result)

    def test_repeat_occurrence_binds_original_bytes(self):
        request = a.requests(packet(['écho echo echo.']))[0]
        result = a.validate(request, revision(request, finding('u0000', 'echo', 2)))
        target = result['findings'][0]['targets'][0]
        self.assertEqual(target['span'], {'start': 11, 'end': 15})

    def test_protected_only_target_rejected(self):
        p = packet(['code prose']); p['protected'] = [{'span': {'start': 0, 'end': 4}}]
        request = a.requests(p)[0]
        with self.assertRaisesRegex(ValueError, 'eligible prose'):
            a.validate(request, revision(request, finding('u0000', 'code')))
        self.assertEqual(len(a.validate(request, revision(request, finding('u0000', 'prose')))['findings']), 1)

    def test_cross_window_finding_has_single_owner(self):
        requests = a.requests(packet([f'Paragraph {i}.' for i in range(65)]))
        item = finding('u0000', 'Paragraph 0.')
        item['targets'].append({'unit': 'u0064', 'quote': 'Paragraph 64.', 'occurrence': 1})
        first = revision(requests[0], item)
        self.assertEqual(len(a.validate(requests[0], first)['findings']), 1)
        with self.assertRaisesRegex(ValueError, 'another focus window'):
            a.validate(requests[1], revision(requests[1], item))
        assembled = a.assemble(requests, [first, response(requests[1])])
        self.assertEqual(assembled['findings'][0]['id'], 'page-test-part-001/f1')
        self.assertEqual(assembled['assessments'][0]['finding_ids'], ['page-test-part-001/f1'])

    def test_request_identity_and_shifted_focus_rejected(self):
        request = a.requests(packet(['Keep condition.'] * 65))[0]
        request['focus'] = request['focus'][1:]
        with self.assertRaisesRegex(ValueError, 'focus partition'): a.verify_request(request)
        request = a.requests(packet(['Keep condition.']))[0]
        result = response(request); result['request'] = 'another'
        with self.assertRaisesRegex(ValueError, 'another request'): a.validate(request, result)

    def test_retained_failure_stays_invalid_with_visible_raw_slice(self):
        with tarfile.open(a.PRIOR / 'evidence.tar.gz', 'r:gz') as archive:
            names = archive.getnames()
            request_name = next(n for n in names if n.endswith('packet/requests/page-019.json'))
            p = json.load(archive.extractfile(request_name))
        request = next(r for r in a.requests(p) if 'u0038' in r['focus'])
        unit = next(u for u in request['units'] if u['id'] == 'u0038')
        self.assertTrue(unit['raw'].startswith('Deployment reports**'))
        with self.assertRaisesRegex(ValueError, 'Quotation is not'):
            a.validate(request, revision(request, finding('u0038', '**Deployment reports**')))
        result = a.validate(request, revision(request, finding('u0038', 'Deployment reports**')))
        self.assertEqual(result['findings'][0]['targets'][0]['span']['start'], unit['span']['start'])


if __name__ == '__main__':
    unittest.main()
