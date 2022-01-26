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

package v1

// VolumeResponse is the schema for the volume response JSON object
type VolumeResponse struct {
	DockerName  string `json:"DockerName,omitempty"`
	Source      string `json:"Source,omitempty"`
	Destination string `json:"Destination,omitempty"`
}

// PortResponse defines the schema for portmapping response JSON
// object.
type PortResponse struct {
	ContainerPort uint16 `json:"ContainerPort,omitempty"`
	Protocol      string `json:"Protocol,omitempty"`
	HostPort      uint16 `json:"HostPort,omitempty"`
}
