/*
Copyright 2022 Google LLC.

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

package v1alpha3

import (
	"testing"
)

func TestValidateCreate(t *testing.T) {
	r := FQDNNetworkPolicy{}
	if _, err := r.GetValidResource().ValidateCreate(); err != nil {
		t.Error("Valid resource marked as invalid during creation")
	}

	if _, err := r.GetValidIngressResource().ValidateCreate(); err != nil {
		t.Error("Valid resource with Ingress policy marked as invalid during creation")
	}

	if _, err := r.GetValidNoPortResource().ValidateCreate(); err != nil {
		t.Error("Valid resource with no port marked as invalid during creation")
	}

	if _, err := r.GetValidNoProtocolResource().ValidateCreate(); err != nil {
		t.Error("Valid resource with no protocol marked as invalid during creation")
	}

	if _, err := r.GetInvalidResource().ValidateCreate(); err == nil {
		t.Error("Invalid resource marked as valid during creation")
	}
}

func TestValidateUpdate(t *testing.T) {
	r := FQDNNetworkPolicy{}
	ro := FQDNNetworkPolicy{}
	r.GetValidResource()
	ro.GetInvalidResource()

	if _, err := r.GetValidResource().ValidateUpdate(&ro); err != nil {
		t.Error("Valid resource marked as invalid during update")
	}

	if _, err := r.GetValidIngressResource().ValidateUpdate(&ro); err != nil {
		t.Error("Valid resource with Ingress policy marked as invalid during update")
	}

	if _, err := r.GetValidNoPortResource().ValidateUpdate(&ro); err != nil {
		t.Error("Valid resource with no port marked as invalid during update")
	}

	if _, err := r.GetValidNoProtocolResource().ValidateUpdate(&ro); err != nil {
		t.Error("Valid resource with no protocol marked as invalid during update")
	}

	ro.GetValidResource()
	if _, err := r.GetInvalidResource().ValidateUpdate(&ro); err == nil {
		t.Error("Invalid resource marked as valid during update")
	}
}

func TestValidateDelete(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if _, err := r.GetValidResource().ValidateDelete(); err != nil {
		t.Error("Valid resource marked as invalid during deletion")
	}

	if _, err := r.GetValidIngressResource().ValidateDelete(); err != nil {
		t.Error("Valid resource with Ingress policy marked as invalid during deletion")
	}

	if _, err := r.GetValidNoPortResource().ValidateDelete(); err != nil {
		t.Error("Valid resource with no port marked as invalid during deletion")
	}

	if _, err := r.GetValidNoProtocolResource().ValidateDelete(); err != nil {
		t.Error("Valid resource with no protocol marked as invalid during deletion")
	}

	r.GetInvalidResource()
	if _, err := r.ValidateDelete(); err != nil {
		t.Error("Impossible to delete invalid resource")
	}
}

func TestValidatePorts(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if r.GetValidResource().ValidatePorts() != nil {
		t.Error("Valid resource marked as having invalid ports")
	}
	if r.GetValidIngressResource().ValidatePorts() != nil {
		t.Error("Valid resource with Ingress policy marked as having invalid ports")
	}
	if r.GetValidNoPortResource().ValidatePorts() != nil {
		t.Error("Valid resource with no port marked as having invalid ports")
	}
	if r.GetValidNoProtocolResource().ValidatePorts() != nil {
		t.Error("Valid resource with no protocol marked as having invalid ports")
	}

	r.LoadResource("./config/samples/networking_v1alpha3_fqdnnetworkpolicy_invalid_port.yaml")
	if r.ValidatePorts() == nil {
		t.Error("Resource with invalid ports marked as valid")
	}
	r.LoadResource("./config/samples/networking_v1alpha3_fqdnnetworkpolicy_invalid_protocol.yaml")
	if r.ValidatePorts() == nil {
		t.Error("Resource with invalid protocol marked as valid")
	}
}

func TestValidateFQDNs(t *testing.T) {
	r := FQDNNetworkPolicy{}

	if r.GetValidResource().ValidateFQDNs() != nil {
		t.Error("Valid resource marked as having invalid FQDNs")
	}
	if r.GetValidIngressResource().ValidateFQDNs() != nil {
		t.Error("Valid resource with Ingress policy marked as having invalid FQDNs")
	}
	if r.GetValidNoPortResource().ValidatePorts() != nil {
		t.Error("Valid resource with no port marked as having invalid ports")
	}
	if r.GetValidNoProtocolResource().ValidatePorts() != nil {
		t.Error("Valid resource with no protocol marked as having invalid ports")
	}

	r.LoadResource("./config/samples/networking_v1alpha3_fqdnnetworkpolicy_invalid_wildcard.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with wildcard marked as valid")
	}
	r.LoadResource("./config/samples/networking_v1alpha3_fqdnnetworkpolicy_invalid_fqdntoolong.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with invalid FQDN (too long) marked as valid")
	}
	r.LoadResource("./config/samples/networking_v1alpha3_fqdnnetworkpolicy_invalid_labeltoolong.yaml")
	if r.ValidateFQDNs() == nil {
		t.Error("Resource with invalid FQDN (label too long) marked as valid")
	}
}
