package acp

import "context"

// ResponseReceipt identifies a matched response within one connection's inbound
// frame order. A zero sequence means this call exposes no response receipt; it
// does not prove that a request was unsent or that no response reached the wire.
// It is not a successful-result or execution receipt, and cannot be compared across
// connections. Gaps in the sequence are permitted.
type ResponseReceipt struct {
	Sequence uint64
}

// NotificationSequence returns the inbound frame sequence on a notification
// handler context. It does not report request or outbound message order.
func NotificationSequence(ctx context.Context) (uint64, bool) {
	if ctx == nil {
		return 0, false
	}
	scope, _ := ctx.Value(notificationContextKey{}).(*notificationCallScope)
	if scope == nil || scope.sequence == 0 {
		return 0, false
	}
	return scope.sequence, true
}
