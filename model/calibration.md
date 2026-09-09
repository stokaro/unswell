# Separate isotonic calibration

`FitIsotonic` fits a monotonic mapping from scalar classifier scores to binary
labels. It implements the `unswell-isotonic-pav-linear-v1` numerical candidate in
[ADR 0021](../docs/adr/0021-isotonic-calibration.md). The mapping is separate from
logistic normalization, weights, and the classifier's original sigmoid response.

Freeze the classifier before scoring a separate calibration partition. Supply
those scores and corresponding 0/1 targets as `CalibrationSample` rows. Choose
the method and comparison protocol on development data; final-test labels never
enter this fit. The numerical API receives no group records and cannot verify
partition independence, annotation validity, feature compatibility, or permission
to train. The experiment/model pack must establish and retain those contracts.

The fitter copies up to 100,000 finite scores and binary labels, sorts by score,
combines ties with their full sample counts, and pools adjacent means until they
increase. Equal levels are pooled too. The response at every distinct score is
its pool's fraction of positive labels. Integer cross-products compare pool means
without floating-point tie ambiguity. Zero and negative zero share one knot.

`Isotonic.Parameters` returns a detached snapshot. `NewIsotonic` rejects invalid
dimensions, nonfinite knots, duplicate or unordered scores, and responses that
decrease or leave [0,1]. The restored mapping owns its data. `Evaluate` returns
the fitted value at a knot and linear interpolation between knots. It accepts
concurrent calls without shared output buffers. It rejects nonfinite inputs and
returns `ErrCalibrationRange` outside the fitted score interval. On any error,
the returned zero is not an available estimate. Single-score fits are defined
only at that score; no extrapolation is implied.

`CalibrationFit` records algorithm and input identities, sample/knot/pool counts,
and mean squared error on its own calibration inputs. Its SHA-256 layout is
specified in the ADR. Reordering rows preserves fitted parameters but changes
the ordered-input identity. Fitting one observed class is mathematically defined;
it does not establish that the calibration data are adequate.

The fit error is not a held-out Brier score. A mapping that returns values between
zero and one has not thereby demonstrated probability accuracy. Compare methods,
coverage, reliability, and subgroup errors on the declared protocol before
selection. Qualified editorial probabilities, model-pack loading, compatibility,
and gate integration remain in #23–#25. Ordinary scans continue to use the
existing rules and index when `calibration.model: none` is selected.

The [external consumer](../examples/consumer/calibration_test.go) executes the
separate fitting stages with disjoint scripted numeric inputs. It tests the API;
its labels are not human annotations or measurements of prose quality.
