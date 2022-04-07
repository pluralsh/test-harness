#!/bin/bash

sleep 10
$app = shift
kubectl wait --for=condition=ready --timeout=30m -n $app applications.app.k8s.io/$app