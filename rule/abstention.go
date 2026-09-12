package rule

// ReasonBudgetExhausted names the abstention of a rule whose per-document
// candidate budget ran out before it finished the document.
const ReasonBudgetExhausted = "budget_exhausted"

// Abstention is the error a rule returns from Evaluate when it declines one
// document for a declared reason. The engine records it, keeps the other
// rules' findings, and treats the run as complete. The abstaining rule adds no
// findings and no activation values for that document. Reason must satisfy
// feature.ValidApplicabilityReason. Any other error from Evaluate still makes
// the analysis incomplete.
type Abstention struct {
	Reason string
	Err    error
}

// Abstain wraps cause as an abstention with the given reason. A nil cause is
// permitted; the reason alone then describes the abstention.
func Abstain(reason string, cause error) error {
	return &Abstention{Reason: reason, Err: cause}
}

// Error names the reason and the cause, when one exists.
func (a *Abstention) Error() string {
	if a.Err == nil {
		return "rule abstained: " + a.Reason
	}
	return "rule abstained: " + a.Reason + ": " + a.Err.Error()
}

// Unwrap exposes the cause to errors.Is and errors.As.
func (a *Abstention) Unwrap() error { return a.Err }

// Detail returns the cause message, or an empty string without a cause.
func (a *Abstention) Detail() string {
	if a.Err == nil {
		return ""
	}
	return a.Err.Error()
}
