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

package container

import (
	"time"

	apicontainerstatus "github.com/mackerelio/mackerel-container-agent/internal/amazon-ecs-agent/agent/api/container/status"
)

// HealthStatus contains the health check result returned by docker
type HealthStatus struct {
	// Status is the container health status
	Status apicontainerstatus.ContainerHealthStatus `json:"status,omitempty"`
	// Since is the timestamp when container health status changed
	Since *time.Time `json:"statusSince,omitempty"`
	// ExitCode is the exitcode of health check if failed
	ExitCode int `json:"exitCode,omitempty"`
	// Output is the output of health check
	Output string `json:"output,omitempty"`
}
