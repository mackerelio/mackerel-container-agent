#!/bin/sh

set -eu

cd $(dirname $0)
d=$(mktemp -d)
trap 'rm -rf $d; exit 1' 1 2 3 15

git clone --depth=1 https://github.com/kubernetes/kubernetes $d
rm -rf stats
cp -R $d/staging/src/k8s.io/kubelet/pkg/apis/stats .
cp $d/LICENSE .
rm -rf $d
