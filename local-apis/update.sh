#!/bin/sh

set -eu

cd $(dirname $0)
rm -rf kubernetes
git clone --depth=1 https://github.com/kubernetes/kubernetes
rm -rf stats
cp -R kubernetes/pkg/kubelet/apis/stats .
rm -rf kubernetes
