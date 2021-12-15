#!/bin/bash

apptestctl bootstrap --kubeconfig="$(kind get kubeconfig)" --install-operators=false
