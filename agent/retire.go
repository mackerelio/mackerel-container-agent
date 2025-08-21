package agent

import (
	"time"

	"github.com/Songmu/retry"

	"github.com/mackerelio/mackerel-container-agent/api"
)

func retire(client api.Client, hostResolver *hostResolver) error {
	hostID, notExist, err := hostResolver.getLocalHostID()
	if err != nil {
		if notExist { // ignore error when the host is not created yet
			return nil
		}
		return err
	}
	logger.Infof("retire: host id = %s", hostID)
	err = retry.Retry(3, 3*time.Second, func() error {
		return client.RetireHost(hostID)
	})
	if err != nil {
		return err
	}
	return hostResolver.removeHostID()
}
