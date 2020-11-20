#!/bin/bash

kubectl apply -f lua-configmap.yml
kubectl apply -f lua-service.yml
kubectl apply -f lua-deployment.yml