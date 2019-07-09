package collector

import (
	"time"

	"github.com/giantswarm/microerror"
)

func convertToTime(datetime string) (time.Time, error) {
	layout := "2006-01-02T15:04:05.000Z"
	t, err := time.Parse(layout, datetime)

	if err != nil {
		return time.Time{}, microerror.Mask(err)
	}

	return t, nil
}
