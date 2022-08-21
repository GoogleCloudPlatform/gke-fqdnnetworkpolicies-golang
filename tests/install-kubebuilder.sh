#!/bin/sh
# Copyright 2022 Google LLC.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
set -x

# Set some variables
target=${KUBEBUILDER_ASSETS:-/kubebuilder}
k8sversion=${K8S_VERSION:-1.24.1}
os=$(go env GOOS)
arch=$(go env GOARCH)

mkdir -p ${target}

# Download and install test tools
curl -sSLo envtest-bins.tar.gz "https://go.kubebuilder.io/test-tools/${k8sversion}/${os}/${arch}"
tar -C ${target} --strip-components=1 -zvxf envtest-bins.tar.gz
mv ${target}/bin/* ${target}/
