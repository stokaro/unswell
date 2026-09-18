# Complete scan costs

Recorded per fresh CLI process; peak RSS includes extraction, NLP and reporting.

| Set | Profile | Seconds baseline → candidate | Peak MiB baseline → candidate |
| --- | --- | --- | --- |
| development | technical | 1.63 → 2.48 | 207.5 → 214.7 |
| development | strict | 1.65 → 1.69 | 186.3 → 191.2 |
| exposed_repetition | technical | 0.58 → 0.63 | 116.5 → 118.9 |
| exposed_repetition | strict | 0.58 → 0.62 | 113.4 → 119.7 |
| exposed_framing | technical | 0.95 → 0.85 | 128.2 → 125.9 |
| exposed_framing | strict | 0.79 → 0.84 | 131.3 → 130.1 |
| exposed_local | technical | 0.67 → 0.72 | 120.7 → 118.1 |
| exposed_local | strict | 0.67 → 0.70 | 119.9 → 122.3 |
| exposed_context | technical | 0.89 → 0.75 | 125.3 → 132.3 |
| exposed_context | strict | 0.88 → 0.75 | 123.6 → 125.2 |
| exposed_instruction | technical | 0.92 → 1.00 | 166.9 → 166.1 |
| exposed_instruction | strict | 0.92 → 0.97 | 165.0 → 168.5 |
| exposed_construction | technical | 0.48 → 0.54 | 101.1 → 98.6 |
| exposed_construction | strict | 0.57 → 0.52 | 106.6 → 98.0 |
| exposed_purpose | technical | 0.48 → 0.53 | 106.9 → 104.3 |
| exposed_purpose | strict | 0.47 → 0.52 | 103.2 → 104.2 |
| exposed_verb | technical | 0.53 → 0.55 | 94.8 → 101.0 |
| exposed_verb | strict | 0.53 → 0.56 | 96.9 → 102.4 |
| exposed_relations | technical | 0.56 → 0.65 | 120.7 → 118.4 |
| exposed_relations | strict | 0.57 → 0.59 | 107.5 → 120.7 |
| exposed_projection | technical | 0.48 → 0.52 | 105.4 → 99.1 |
| exposed_projection | strict | 0.48 → 0.56 | 103.1 → 111.1 |
| exposed_scope | technical | 0.73 → 0.76 | 171.5 → 175.5 |
| exposed_scope | strict | 0.71 → 0.78 | 152.4 → 161.0 |
| exposed_proposition | technical | 0.43 → 0.48 | 84.5 → 85.9 |
| exposed_proposition | strict | 0.42 → 0.52 | 81.0 → 82.2 |
| exposed_clauses | technical | 0.48 → 0.50 | 92.6 → 87.4 |
| exposed_clauses | strict | 0.46 → 0.50 | 83.1 → 92.4 |
| exposed_action | technical | 0.45 → 0.48 | 92.8 → 93.0 |
| exposed_action | strict | 0.46 → 0.48 | 94.2 → 93.4 |
| exposed_roles | technical | 0.48 → 0.52 | 89.9 → 93.1 |
| exposed_roles | strict | 0.48 → 0.53 | 95.6 → 94.2 |
| exposed_boundaries | technical | 0.52 → 0.55 | 108.2 → 109.8 |
| exposed_boundaries | strict | 0.52 → 0.55 | 101.6 → 114.0 |
| exposed_grammar | technical | 0.46 → 0.52 | 101.5 → 98.4 |
| exposed_grammar | strict | 0.46 → 0.49 | 104.6 → 104.0 |
| exposed_discourse | technical | 0.80 → 0.73 | 107.6 → 121.5 |
| exposed_discourse | strict | 0.71 → 0.75 | 105.7 → 123.6 |
| confirmation | technical | 0.65 → 0.57 | 106.6 → 106.4 |
| confirmation | strict | 0.56 → 0.60 | 97.6 → 101.4 |
