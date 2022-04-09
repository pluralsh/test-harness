#!/bin/bash

sleep 10
kubectl wait --for=condition=ready --timeout=30m -n $1 applications.app.k8s.io/$1