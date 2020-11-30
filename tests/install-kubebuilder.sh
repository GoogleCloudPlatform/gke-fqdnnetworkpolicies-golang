#!/bin/sh

set -e
set -x

version=${KUBEBUILDER_VERSION:-2.3.1}
target=${KUBEBUILDER_ASSETS:-/kubebuilder/bin}
os=$(go env GOOS)
arch=$(go env GOARCH)

# download kubebuilder and extract it to tmp
wget --quiet -O - https://go.kubebuilder.io/dl/${version}/${os}/${arch} | tar -xz -C /tmp

mv /tmp/kubebuilder_${version}_${os}_${arch}/bin ${target}