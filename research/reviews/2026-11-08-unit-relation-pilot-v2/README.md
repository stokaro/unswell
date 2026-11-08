# Terminal result: mandatory source-relation verdicts, V2

The [signed V2 packet](PROTOCOL.md) reused the first V1 Fastfetch inventory
response byte-for-byte. It then made two new sequential GPT-6-Astra medium
calls. The second Fastfetch inventory chunk was valid. The Fastfetch verdict
response failed the frozen validator because five decisions used `diagnosis`
where the contract required `diagnose`. The prompt and JSON schema did not
specify that exact token. The run stopped at the first error; the remaining
six new calls were not made. No page has a valid delivered verdict, so the
registered editorial quality metrics are unscored.

The raw, invalid verdict is still useful for deciding whether to repeat this
method. It contained one diagnosis overlapping a frozen Fastfetch event and
abstained on the other two. Even perfect recovery of the two events on the
remaining pages would yield at most 3/5 frozen events, below the preregistered
4/5 advance threshold. This is a selected, exposed development observation,
not a scored quality gain or a prospective result. A format-only V3 repetition
of this same arm is not warranted. The model did identify additional source
claims, but none changes that upper bound.

The source pages, five frozen assistant-reviewed events, protected
counterexamples, inventories, signed packets, exact requests, raw tool-free
traces, journal, and stop result are preserved here. `python3 -B
verify_terminal.py --check` independently replays the packet, source-bound
inventory, response hashes and traces, first-error stop, and terminal result.
`python3 -B -m unittest test_packet.py test_verify_terminal.py` includes
negative tamper checks. No product detector, output policy, or #349 acceptance
claim changed.

The next #349 experiment needs a different semantic-recall hypothesis and
fresh frozen pages, followed by complete 123-event development accounting and
separate prospective confirmation before any product quality claim.
