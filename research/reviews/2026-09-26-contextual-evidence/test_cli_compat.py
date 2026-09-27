import json
import unittest
from cli_compat import NOTICE, trace_usage

class StartupNoticeTests(unittest.TestCase):
    def trace(self, middle):
        return '\n'.join(json.dumps(e) for e in [
            {'type':'thread.started'}, *middle, {'type':'turn.started'},
            {'type':'item.completed','item':{'type':'agent_message','text':'{}'}},
            {'type':'turn.completed','usage':{'output_tokens':2}}])

    def notice(self, message=NOTICE):
        return {'type':'item.completed','item':{'id':'item_0','type':'error','message':message}}

    def test_exact_startup_notice_retains_usage(self):
        self.assertEqual(trace_usage(self.trace([self.notice()])), {'output_tokens':2})

    def test_no_notice_is_also_valid(self):
        self.assertEqual(trace_usage(self.trace([])), {'output_tokens':2})

    def test_other_errors_tools_or_repeated_notice_are_rejected(self):
        for events in [[self.notice('Authentication failed')], [self.notice(),self.notice()],
                       [{'type':'item.completed','item':{'type':'command_execution'}}],
                       [{'type':'turn.started'},self.notice()]]:
            with self.subTest(events=events), self.assertRaises(ValueError): trace_usage(self.trace(events))
