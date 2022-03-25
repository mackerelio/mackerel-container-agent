// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//  http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package status

import (
	"errors"
	"strings"
)

const (
	// ContainerHealthUnknown is the initial status of container health
	ContainerHealthUnknown ContainerHealthStatus = iota
	// ContainerHealthy represents the status of container health check when returned healthy
	ContainerHealthy
	// ContainerUnhealthy represents the status of container health check when returned unhealthy
	ContainerUnhealthy
)

// ContainerHealthStatus is an enumeration of container health check status
type ContainerHealthStatus int32

// BackendStatus returns the container health status recognized by backend
func (healthStatus ContainerHealthStatus) BackendStatus() string {
	switch healthStatus {
	case ContainerHealthy:
		return "HEALTHY"
	case ContainerUnhealthy:
		return "UNHEALTHY"
	default:
		return "UNKNOWN"
	}
}

// String returns the readable description of the container health status
func (healthStatus ContainerHealthStatus) String() string {
	return healthStatus.BackendStatus()
}

// UnmarshalJSON overrides the logic for parsing the JSON-encoded container health data
func (healthStatus *ContainerHealthStatus) UnmarshalJSON(b []byte) error {
	*healthStatus = ContainerHealthUnknown

	if strings.ToLower(string(b)) == "null" {
		return nil
	}
	if b[0] != '"' || b[len(b)-1] != '"' {
		return errors.New("container health status unmarshal: status must be a string or null; Got " + string(b))
	}

	strStatus := string(b[1 : len(b)-1])
	switch strStatus {
	case "UNKNOWN":
	// The health status is already set to ContainerHealthUnknown initially
	case "HEALTHY":
		*healthStatus = ContainerHealthy
	case "UNHEALTHY":
		*healthStatus = ContainerUnhealthy
	default:
		return errors.New("container health status unmarshal: unrecognized status: " + string(b))
	}
	return nil
}

// MarshalJSON overrides the logic for JSON-encoding the ContainerHealthStatus type
func (healthStatus *ContainerHealthStatus) MarshalJSON() ([]byte, error) {
	if healthStatus == nil {
		return nil, nil
	}
	return []byte(`"` + healthStatus.String() + `"`), nil
}
