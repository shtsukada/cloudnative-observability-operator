package conditions

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ConditionReady       = "Ready"
	ConditionProgressing = "Progressing"
	ConditionDegraded    = "Degraded"
)

const (
	ReasonReconciling           = "Reconciling"
	ReasonApplySucceeded        = "ApplySucceeded"
	ReasonApplyFailed           = "ApplyFailed"
	ReasonWaitingForDeployment  = "WaitingForDeployment"
	ReasonDeploymentAvailable   = "DeploymentAvailable"
	ReasonDeploymentUnavailable = "DeploymentUnavailable"
	ReasonImagePullBackOff      = "ImagePullBackOff"
	ReasonErrForbidden          = "Forbidden"
	ReasonErrInvalid            = "Invalid"
	ReasonErrNotFound           = "NotFound"
	ReasonErrConflict           = "Conflict"
	ReasonErrUnknown            = "Unknown"
)

const (
	EventTypeNormal  = "Normal"
	EventTypeWarning = "Warning"
)

var (
	CondTrue  = metav1.ConditionTrue
	CondFalse = metav1.ConditionFalse
)
