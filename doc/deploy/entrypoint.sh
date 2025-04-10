#!/usr/bin/env bash
set -e

NAME=${NAME:-"coredns"}
NAMESPACE=${NAMESPACE:-"dcloud"}
CHARTS=${CHARTS:-"./charts/coredns"}

helm upgrade -i ${NAME} ${CHARTS} -n ${NAMESPACE} --create-namespace -f values.yaml

# helm uninstall -n dcloud coredns