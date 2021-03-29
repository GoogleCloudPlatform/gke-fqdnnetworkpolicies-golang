// Copyright 2020 Google LLC
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
	"testing"
)

func TestValidateCreate(t *testing.T) {
	r := FQDNNetworkPolicy{}
	if r.GetValidResource().ValidateCreate() != nil {
		t.Error("Valid resource marked as invalid during creation")
	}

	if r.GetValidNoPortResource().ValidateCreate() != nil {
		t.Error("Valid resource with no port marked as invalid during creation")
	}

	if r.GetValidNoProtocolResource().ValidateCreate() != nil {
		t.Error("Valid resource with no protocol marked as invalid during creation")
	}

	if r.GetInvalidResource().ValidateCreate() == nil {
		t.Error("Invalid resource marked as valid during creation")
	}
}

func TestValidateUpdate(t *testing.T) {
	r := FQDNNetworkPolicy{}
	ro := FQDNNetworkPolicy{}
	r.GetValidResource()
	ro.GetInvalidResource()

	if r.GetValidResource().ValidateUpdate(&ro) != nil {
		t.Error("Valid resource marked as invalid during update")
	}

	if r.GetValidNoPortResource().ValidateUpdate(&ro) != nil {
		t.Error("Valid resource with no port marked as invalid during update")
	}

	if r.GetValidNoProtocolResource().ValidateUpdate(&ro) != nil {
		t.Error("Valid resource with no protocol marked as invalid during update")
	}

	ro.GetValidResource()
	if r.GetInvalidResource().ValidateUpdate(&ro) == nil {
		t.Error("Invalid resource marked as valid during update")
	}
}

func TestValidateDelete(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if r.GetValidResource().ValidateDelete() != nil {
		t.Error("Valid resource marked as invalid during deletion")
	}

	if r.GetValidNoPortResource().ValidateDelete() != nil {
		t.Error("Valid resource with no port marked as invalid during deletion")
	}

	if r.GetValidNoProtocolResource().ValidateDelete() != nil {
		t.Error("Valid resource with no protocol marked as invalid during deletion")
	}

	r.GetInvalidResource()
	if r.ValidateDelete() != nil {
		t.Error("Impossible to delete invalid resource")
	}
}

func TestValidatePorts(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if r.GetValidResource().ValidatePorts() != nil {
		t.Error("Valid resource marked as having invalid ports")
	}
	if r.GetValidNoPortResource().ValidatePorts() != nil {
		t.Error("Valid resource with no port marked as having invalid ports")
	}
	if r.GetValidNoProtocolResource().ValidatePorts() != nil {
		t.Error("Valid resource with no protocol marked as having invalid ports")
	}

	r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_invalid_port.yaml")
	if r.ValidatePorts() == nil {
		t.Error("Resource with invalid ports marked as valid")
	}
	r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_invalid_protocol.yaml")
	if r.ValidatePorts() == nil {
		t.Error("Resource with invalid protocol marked as valid")
	}
}

func TestValidateFQDNs(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if r.GetValidResource().ValidateFQDNs() != nil {
		t.Error("Valid resource marked as having invalid FQDNs")
	}
	if r.GetValidNoPortResource().ValidatePorts() != nil {
		t.Error("Valid resource with no port marked as having invalid ports")
	}
	if r.GetValidNoProtocolResource().ValidatePorts() != nil {
		t.Error("Valid resource with no protocol marked as having invalid ports")
	}

	r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_invalid_wildcard.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with wildcard marked as valid")
	}
	r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_invalid_fqdntoolong.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with invalid FQDN (too long) marked as valid")
	}
	r.LoadResource("./config/samples/networking_v1alpha2_fqdnnetworkpolicy_invalid_labeltoolong.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with invalid FQDN (label too long) marked as valid")
	}
}
