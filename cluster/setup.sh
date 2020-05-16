#!/bin/bash

set -e

docker ps

# script creates a Kubernetes cluser from scratch
# configuration file is cluser-config.yaml

# if set to 1 create admin account
create_admin=1

cluster_name="kind"
echo $cluster_name > CLUSTERNAME

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

exec $@