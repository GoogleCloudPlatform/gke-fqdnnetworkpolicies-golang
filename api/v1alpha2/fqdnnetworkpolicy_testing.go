// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*


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

package v1alpha2

import (
	"io/ioutil"
	"log"
	"runtime"
	"strings"

	"sigs.k8s.io/yaml"
)

// LoadResource unmarshalls a given yaml file in a FQDNNetworkPolicy
func (r *FQDNNetworkPolicy) LoadResource(path string) *FQDNNetworkPolicy {
	// Path to this file
	_, filename, _, _ := runtime.Caller(0)
	pSlice := strings.Split(filename, "/")
	projectRoot := strings.Join(pSlice[:len(pSlice)-3], "/")
	yamlFile, err := ioutil.ReadFile(projectRoot + "/" + path)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, r)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return r
}

// GetValidResource returns loads a valid FQDNNetworkPolicy for testing
func (r *FQDNNetworkPolicy) GetValidResource() *FQDNNetworkPolicy {
	return r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_valid.yaml")
}

// GetValidIngressResource returns loads a valid FQDNNetworkPolicy with an Ingress policy for testing
func (r *FQDNNetworkPolicy) GetValidIngressResource() *FQDNNetworkPolicy {
	return r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_valid_ingress.yaml")
}

func (r *FQDNNetworkPolicy) GetValidNoPortResource() *FQDNNetworkPolicy {
	return r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_valid_noport.yaml")
}

func (r *FQDNNetworkPolicy) GetValidNoProtocolResource() *FQDNNetworkPolicy {
	return r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_valid_noprotocol.yaml")
}

func (r *FQDNNetworkPolicy) GetValidNonExistentFQDNResource() *FQDNNetworkPolicy {
	return r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_valid_nonexistentfqdn.yaml")
}

func (r *FQDNNetworkPolicy) GetInvalidResource() *FQDNNetworkPolicy {
	return r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_invalid.yaml")
}
