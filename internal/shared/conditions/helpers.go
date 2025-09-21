package conditions

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCondition(condType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	now := metav1.NewTime(time.Now())
	return metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: 0,
		LastTransitionTime: now,
	}
}

func Upsert(conds *[]metav1.Condition, c metav1.Condition) {
	list := *conds
	for i := range list {
		if list[i].Type == c.Type {
			if list[i].Status != c.Status {
				c.LastTransitionTime = metav1.Now()
			} else {
				c.LastTransitionTime = list[i].LastTransitionTime
			}
			list[i] = c
			*conds = list
			return
		}
	}
	*conds = append(*conds, c)
}

func ReadyTrue(msg string) metav1.Condition {
	return NewCondition(ConditionReady, CondTrue, ReasonDeploymentAvailable, msg)
}
func ReadyFalseProgressing(msg string) metav1.Condition {
	return NewCondition(ConditionReady, CondFalse, ReasonWaitingForDeployment, msg)
}
func ReadyFalseDegraded(reason, msg string) metav1.Condition {
	return NewCondition(ConditionReady, CondFalse, reason, msg)
}
func Progressing(msg string) metav1.Condition {
	return NewCondition(ConditionProgressing, CondTrue, ReasonReconciling, msg)
}
func Degraded(reason, msg string) metav1.Condition {
	return NewCondition(ConditionDegraded, CondTrue, reason, msg)
}
