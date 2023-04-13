package trace

// GetTypeMarker returns the type marker indicating the type of request.
func GetTypeMarker(sc SpanContext) int {
	if sc == nil {
		return 0
	}

	return int((getSpecialFlag(sc) & typeMarkerMask) >> flagBitsNonTypeMarker)
}

// GetRequestType returns the request type as indicated by the span context.
func GetRequestType(sc SpanContext) string {
	typeMarker := GetTypeMarker(sc)
	switch typeMarker {
	case 0:
		return ReqTypeOldFormat
	case 1:
		return ReqTypeNormal
	case 2:
		return ReqTypeDebug
	case 3:
		return ReqTypeStressTest
	case 4:
		return ReqTypeShadow
	default:
		return ReqTypeUnknown
	}
}

// IsSpanContextDebug indicates whether the request is on debug mode.
func IsSpanContextDebug(sc SpanContext) bool {
	if sc == nil {
		return false
	}

	// For the old format, check debug bit.
	if isOldFormat(sc) {
		return (getSpecialFlag(sc) & traceFlagOldDebug) == traceFlagOldDebug
	}

	// new format, checking type marker
	return (getSpecialFlag(sc) & typeMarkerMask) == typeMarkerForDebug
}

// IsSpanContextFromStressTest indicates whether the request is part of a stress test.
func IsSpanContextFromStressTest(sc SpanContext) bool {
	if sc == nil {
		return false
	}

	return (getSpecialFlag(sc) & typeMarkerMask) == typeMarkerForStressTest
}

// IsSpanContextShadow indicates whether the request is shadow.
func IsSpanContextShadow(sc SpanContext) bool {
	if sc == nil {
		return false
	}

	return (getSpecialFlag(sc) & typeMarkerMask) == typeMarkerForShadow
}

// isOldFormat indicates whether the request uses the old span context ID format
func isOldFormat(sc SpanContext) bool {
	return (getSpecialFlag(sc) & typeMarkerMask) == typeMarkerForOldFormat
}

// IsSpanContextSampled indicates whether the request is to be sampled.
// It can be sampled when either sampling flag or debug flag is true.
func IsSpanContextSampled(sc SpanContext) bool {
	if sc == nil {
		return false
	}

	// SpanContextID generated without this lib could have random flags which leads to the high probability of sampling
	// This is a rough version check to reduce the probability of this kind of sampling
	return (getSpecialFlag(sc)&traceFlagSampled) == traceFlagSampled || IsSpanContextDebug(sc)
}

// IsSpanContextCritical indicates whether the request is critical.
func IsSpanContextCritical(sc SpanContext) bool {
	if sc == nil {
		return false
	}

	return (getSpecialFlag(sc) & traceFlagCritical) == traceFlagCritical
}

func getSpecialFlag(sc SpanContext) byte {
	return sc.TraceID()[traceIDSize-1]
}
