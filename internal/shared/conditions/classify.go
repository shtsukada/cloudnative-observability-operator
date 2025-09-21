package conditions

import apierrors "k8s.io/apimachinery/pkg/api/errors"

func ClassifyApplyError(err error) string {
	if err == nil {
		return ReasonApplySucceeded
	}
	switch {
	case apierrors.IsForbidden(err):
		return ReasonErrForbidden
	case apierrors.IsInvalid(err):
		return ReasonErrInvalid
	case apierrors.IsNotFound(err):
		return ReasonErrNotFound
	case apierrors.IsConflict(err):
		return ReasonErrConflict
	default:
		return ReasonErrUnknown
	}
}
