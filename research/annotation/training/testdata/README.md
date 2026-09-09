# Scripted training fixtures

Every source and response in this directory is a teaching fixture created for
Unswell tests. The round declares two simulated raters and requires an explicit
simulation option for fitting. No human annotation or statistical qualification
is represented by these files.

The six sources are pinned to training (two), calibration (two), development (one),
and final test (one). These declarations exercise partition mechanics; they do not
establish independent real-world samples. Only paragraph targets have responses.
The corresponding sentence targets test that parent labels are not inherited.
Development and final-test sources deliberately lack a training permission.

The two training paragraphs have 3 and 9 words. Their training-only population mean
is 6 and standard deviation is 3. Calibration uses different paragraphs and does
not alter these parameters. Tests replace reserved text with much longer prose
and change a final-test label; the fitted classifier and calibration stay equal.

These texts and scripted labels must not be used as accepted human data, a quality
benchmark, evidence of independent annotation, or a basis for product probabilities.
