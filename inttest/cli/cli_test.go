/*
Copyright 2021 k0s authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package ctr

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/k0sproject/k0s/inttest/common"
	"github.com/stretchr/testify/suite"
)

type CliSuite struct {
	common.FootlooseSuite
}

func (s *CliSuite) TestK0sCliCommandNegative() {
	ssh, err := s.SSH(s.ControllerNode(0))
	s.Require().NoError(err)
	defer ssh.Disconnect()

	// k0s controller command should fail if non existent path to config is passed
	_, err = ssh.ExecWithOutput("k0s controller --config /some/fake/path")
	s.Require().Error(err)

	// k0s install command should fail if non existent path to config is passed
	_, err = ssh.ExecWithOutput("k0s install controller --config /some/fake/path")
	s.Require().Error(err)

	// k0s start should fail if service is not installed
	_, err = ssh.ExecWithOutput("k0s start")
	s.Require().Error(err)

	// k0s stop should fail if service is not installed
	_, err = ssh.ExecWithOutput("k0s stop")
	s.Require().Error(err)

}

func (s *CliSuite) TestK0sCliKubectlCommand() {
	ssh, err := s.SSH(s.ControllerNode(0))
	s.Require().NoError(err)
	defer ssh.Disconnect()

	_, err = ssh.ExecWithOutput("k0s install controller --enable-worker")
	s.Require().NoError(err)

	_, err = ssh.ExecWithOutput("k0s start")
	s.Require().NoError(err)

	err = s.WaitForKubeAPI(s.ControllerNode(0))
	s.Require().NoError(err)

	output, err := ssh.ExecWithOutput("k0s kubectl get namespaces -o json")
	s.Require().NoError(err)

	namespaces := &K8sNamespaces{}

	err = json.Unmarshal([]byte(output), namespaces)
	s.NoError(err)

	s.Len(namespaces.Items, 4)
	s.Equal("default", namespaces.Items[0].Metadata.Name)
	s.Equal("kube-node-lease", namespaces.Items[1].Metadata.Name)
	s.Equal("kube-public", namespaces.Items[2].Metadata.Name)
	s.Equal("kube-system", namespaces.Items[3].Metadata.Name)

	_, err = ssh.ExecWithOutput("k0s stop")
	s.Require().NoError(err)
}

func TestCliCommandSuite(t *testing.T) {
	s := CliSuite{
		common.FootlooseSuite{
			ControllerCount: 1,
		},
	}
	suite.Run(t, &s)
}

type K8sNamespaces struct {
	APIVersion string `json:"apiVersion"`
	Items      []struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Labels            struct {
				KubernetesIoMetadataName string `json:"kubernetes.io/metadata.name"`
			} `json:"labels"`
			ManagedFields []struct {
				APIVersion string `json:"apiVersion"`
				FieldsType string `json:"fieldsType"`
				FieldsV1   struct {
					FMetadata struct {
						FLabels struct {
							FKubernetesIoMetadataName struct {
							} `json:"f:kubernetes.io/metadata.name"`
						} `json:"f:labels"`
					} `json:"f:metadata"`
				} `json:"fieldsV1"`
				Manager   string    `json:"manager"`
				Operation string    `json:"operation"`
				Time      time.Time `json:"time"`
			} `json:"managedFields"`
			Name            string `json:"name"`
			ResourceVersion string `json:"resourceVersion"`
			UID             string `json:"uid"`
		} `json:"metadata"`
		Spec struct {
			Finalizers []string `json:"finalizers"`
		} `json:"spec"`
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	} `json:"items"`
	Kind     string `json:"kind"`
	Metadata struct {
		ResourceVersion string `json:"resourceVersion"`
		SelfLink        string `json:"selfLink"`
	} `json:"metadata"`
}
