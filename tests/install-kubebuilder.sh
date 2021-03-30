#!/bin/sh
# Copyright 2021 Google LLC
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

version=${KUBEBUILDER_VERSION:-2.3.1}
target=${KUBEBUILDER_ASSETS:-/kubebuilder/bin}
os=$(go env GOOS)
arch=$(go env GOARCH)

# download kubebuilder and extract it to tmp
wget --quiet -O - https://go.kubebuilder.io/dl/${version}/${os}/${arch} | tar -xz -C /tmp

mv /tmp/kubebuilder_${version}_${os}_${arch}/bin ${target}