#!/bin/bash

set -e

# script creates a Kubernetes cluser from scratch
# configuration file is cluser-config.yaml

# if set to 1 create admin account
create_admin=1

# set path to kind
export PATH=$PATH:$(go env GOPATH)/bin

if [ "$1" != "" ]; then
    cluster_name=$1
else
    cluster_name="kind"
fi

echo $cluster_name > CLUSTERNAME

if [ "$2" != "" ]; then
    echo "Passed to many arguments"
    echo -e "usage:\n./setup.sh 'cluster name' "
    exit 1
fi

kind delete cluster --name $cluster_name
kind create cluster --config cluster-config.yaml --name $cluster_name

kind get kubeconfig --name $cluster_name > ../core/kubeconfig

if [ $create_admin == 1 ]; then
    echo "Admin account created"
    
    kubectl create -n kube-system serviceaccount admin
    kubectl create clusterrolebinding permissive-binding \
     --clusterrole=cluster-admin \
     --user=admin \
     --user=kubelet \
     --group=system:serviceaccounts

     admin_token_name=$(kubectl -n kube-system get serviceaccount admin -o yaml | tail -1 | cut -d":" -f 2 | cut -d" " -f2)
     admin_token=$(kubectl -n kube-system get secret $admin_token_name -o yaml | grep token | head -1 | sed -e 's/.* \(.*\)$/\1/')
     echo "Admin token: "
     echo $admin_token | base64 --decode
     echo ""
else
    echo "Admin account not created"
fi

echo "------------------------"

context="kind-"$cluster_name
kubectl cluster-info --context $context