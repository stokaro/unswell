#!/usr/bin/env python3
"""Answer the questions of `unswell compare` without a model, so that the parts that need no model can be read."""
import sys
prompt = sys.stdin.read()
if "more_like_a_model" in prompt:
    print('{"more_like_a_model": "same", "reason": "not asked"}')
elif "List the statements" in prompt:
    print('{"facts": ["Nothing was asked."]}')
elif "numbered pairs" in prompt:
    import re
    pairs = len(re.findall(r"^===== PAIR \d+ =====$", prompt, flags=re.M))
    print('{"sentences": [' + ", ".join('{"sentence": %d, "kept": "missing", "quote": ""}' % (n + 1) for n in range(pairs)) + ']}')
else:
    print('{"facts": [{"fact": 1, "kept": "missing", "quote": ""}]}')
