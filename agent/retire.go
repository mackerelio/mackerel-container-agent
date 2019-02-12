package agent

import (
	"os"
	"time"

	"github.com/Songmu/retry"

	"github.com/mackerelio/mackerel-container-agent/api"
)

func retire(client api.Client, hostResolver *hostResolver) error {
	hostID, err := hostResolver.getLocalHostID()
	if err != nil {
		if os.IsNotExist(err) { // ignore error when the host is not created yet
			return nil
		}
		return err
	}
	err = retry.Retry(3, 3*time.Second, func() error {
		return client.RetireHost(hostID)
	})
	if err != nil {
		return err
	}
	return hostResolver.removeHostID()
}
