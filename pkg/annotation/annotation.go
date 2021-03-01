// Package annotation contains common Kubernetes metadata. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package annotation

const (
	// ChartOperatorPaused annotation when present prevents chart-operator from
	// reconciling the resource.
	ChartOperatorPaused = "chart-operator.giantswarm.io/paused"

	// CordonReason is the name of the annotation that indicates
	// the reason of why chart-operator should not apply any update on this chart CR.
	CordonReason = "chart-operator.giantswarm.io/cordon-reason"

	// CordonUntilDate is the name of the annotation that indicates
	// the expiration date of rule of this cordon.
	CordonUntilDate = "chart-operator.giantswarm.io/cordon-until"

	// ForceHelmUpgrade is the name of the annotation that controls whether
	// force is used when upgrading the Helm release.
	ForceHelmUpgrade = "chart-operator.giantswarm.io/force-helm-upgrade"

	// RollbackCount is the name of the annotation storing the number of
	// rollbacks performed from the previous pending status.
	RollbackCount = "chart-operator.giantswarm.io/rollback-count"

	// ValuesMD5Checksum is the name of the annotation storing an MD5 checksum
	// of the Helm release values.
	ValuesMD5Checksum = "chart-operator.giantswarm.io/values-md5-checksum"

	Webhook = "chart-operator.giantswarm.io/webhook-url"
)
