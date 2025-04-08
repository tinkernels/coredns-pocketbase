#!/bin/bash
set -e

helm pull coredns/coredns --version=1.39.2 -d charts/ --untar
