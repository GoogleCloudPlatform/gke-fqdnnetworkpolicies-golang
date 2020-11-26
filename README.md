# FQDNNetworkPolicies

FQDNNetworkPolicies let you create Kubernetes Network Policies based on FQDNs
and not only IPs or CIDRs.

## How does it work?

TODO

## Limitations

TODO

## Development

This project is built with [kubebuilder](https://book.kubebuilder.io/introduction.html).
We recommend using [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) for your development environment.

### Prerequisites
You need the following tools installed on your development workstation.
* Docker
* kubectl
* Kind
* kustomize
* kubebuilder

### Getting up and running

1. Create your Kind cluster.

    ```
    make kind-cluster
    ```

1. Deploy cert-manager (necessary for the webhooks).

   ```
   make deploy-cert-manager
   ```

1. Build & deploy the controller. This will delete any previous controller pod running, even if it has the same tag.

   ```
   make force-deploy-manager
   ```

1. Observe the controller logs and apply valid and invalid resources.

   ```
   # In one terminal
   make follow-manager-logs
   # In another terminal
   kubectl apply -f config/samples/networking_v1alpha1_fqdnnetworkpolicy_invalid.yaml
   kubectl apply -f config/samples/networking_v1alpha1_fqdnnetworkpolicy_valid.yaml
   ```

1. Explore the Makefile for other available commands, and read the [kubebuilder book](https://book.kubebuilder.io/introduction.html).