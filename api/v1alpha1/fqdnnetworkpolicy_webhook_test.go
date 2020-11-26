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

package v1alpha1

import (
	"io/ioutil"
	"log"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

// loadResource loads a FQDNNetworkPolicy from a yaml file
// The path must be given relative to the root of the project
func (r *FQDNNetworkPolicy) loadResource(path string) *FQDNNetworkPolicy {
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

func (r *FQDNNetworkPolicy) getValidResource() *FQDNNetworkPolicy {
	return r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_valid.yaml")
}

func (r *FQDNNetworkPolicy) getValidNoPortResource() *FQDNNetworkPolicy {
	return r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_valid_noport.yaml")
}

func (r *FQDNNetworkPolicy) getValidNoProtocolResource() *FQDNNetworkPolicy {
	return r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_valid_noprotocol.yaml")
}

func (r *FQDNNetworkPolicy) getInvalidResource() *FQDNNetworkPolicy {
	return r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_invalid.yaml")
}

func TestValidateCreate(t *testing.T) {
	r := FQDNNetworkPolicy{}
	if r.getValidResource().ValidateCreate() != nil {
		t.Error("Valid resource marked as invalid during creation")
	}

	if r.getValidNoPortResource().ValidateCreate() != nil {
		t.Error("Valid resource with no port marked as invalid during creation")
	}

	if r.getValidNoProtocolResource().ValidateCreate() != nil {
		t.Error("Valid resource with no protocol marked as invalid during creation")
	}

	if r.getInvalidResource().ValidateCreate() == nil {
		t.Error("Invalid resource marked as valid during creation")
	}
}

func TestValidateUpdate(t *testing.T) {
	r := FQDNNetworkPolicy{}
	ro := FQDNNetworkPolicy{}
	r.getValidResource()
	ro.getInvalidResource()

	if r.getValidResource().ValidateUpdate(&ro) != nil {
		t.Error("Valid resource marked as invalid during update")
	}

	if r.getValidNoPortResource().ValidateUpdate(&ro) != nil {
		t.Error("Valid resource with no port marked as invalid during update")
	}

	if r.getValidNoProtocolResource().ValidateUpdate(&ro) != nil {
		t.Error("Valid resource with no protocol marked as invalid during update")
	}

	ro.getValidResource()
	if r.getInvalidResource().ValidateUpdate(&ro) == nil {
		t.Error("Invalid resource marked as valid during update")
	}
}

func TestValidateDelete(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if r.getValidResource().ValidateDelete() != nil {
		t.Error("Valid resource marked as invalid during deletion")
	}

	if r.getValidNoPortResource().ValidateDelete() != nil {
		t.Error("Valid resource with no port marked as invalid during deletion")
	}

	if r.getValidNoProtocolResource().ValidateDelete() != nil {
		t.Error("Valid resource with no protocol marked as invalid during deletion")
	}

	r.getInvalidResource()
	if r.ValidateDelete() != nil {
		t.Error("Impossible to delete invalid resource")
	}
}

func TestValidatePorts(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if r.getValidResource().ValidatePorts() != nil {
		t.Error("Valid resource marked as having invalid ports")
	}
	if r.getValidNoPortResource().ValidatePorts() != nil {
		t.Error("Valid resource with no port marked as having invalid ports")
	}
	if r.getValidNoProtocolResource().ValidatePorts() != nil {
		t.Error("Valid resource with no protocol marked as having invalid ports")
	}

	r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_invalid_port.yaml")
	if r.ValidatePorts() == nil {
		t.Error("Resource with invalid ports marked as valid")
	}
	r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_invalid_protocol.yaml")
	if r.ValidatePorts() == nil {
		t.Error("Resource with invalid protocol marked as valid")
	}
}

func TestValidateFQDNs(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if r.getValidResource().ValidateFQDNs() != nil {
		t.Error("Valid resource marked as having invalid FQDNs")
	}
	if r.getValidNoPortResource().ValidatePorts() != nil {
		t.Error("Valid resource with no port marked as having invalid ports")
	}
	if r.getValidNoProtocolResource().ValidatePorts() != nil {
		t.Error("Valid resource with no protocol marked as having invalid ports")
	}

	r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_invalid_wildcard.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with wildcard marked as valid")
	}
	r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_invalid_fqdntoolong.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with invalid FQDN (too long) marked as valid")
	}
	r.loadResource("./config/samples/networking_v1alpha1_fqdnnetworkpolicy_invalid_labeltoolong.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with invalid FQDN (label too long) marked as valid")
	}
}
