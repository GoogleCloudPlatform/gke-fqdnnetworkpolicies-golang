# FQDNNetworkPolicies

FQDNNetworkPolicies let you create Kubernetes Network Policies based on FQDNs
and not only IPs or CIDRs.

## How does it work?

A FQDNNetworkPolicy looks a lot like a NetworkPolicy, but you can configure hostnames
in the "to" field:

```
apiVersion: networking.gke.io/v1alpha1
kind: FQDNNetworkPolicy
metadata:
  name: example
spec:
  podSelector:
    matchLabels:
      role: example
  egress:
    - to:
      - fqdns:
        - example.com
      ports:
      - port: 443
        protocol: TCP
```

When you create this FQDNNetworkPolicy, it will in turn create a corresponding NetworkPolicy with
the same name, in the same namespace, that has the same `podSelector`, the same ports, but replacing
the hostnames with corresponding IPs.

### Annotations

There are 2 annotations to know when working with FQDNNetworkPolicies.

If a NetworkPolicy has been created by a FQDNNetworkPolicy, it has the `fqdnnetworkpolicies.networking.gke.io/owned-by`
set to the name of the FQDNNetworkPolicy. If, when you create a FQDNNetworkPolicy, a NetworkPolicy with the same name
already exists, then the FQDNNetworkPolicy will not do anything. You can have the FQDNNetworkPolicy "adopt" the
NetworkPolicy by manually setting the `fqdnnetworkpolicies.networking.gke.io/owned-by` to the right value on the
NetworkPolicy.

By default, the NetworkPolicy associated with a FQDNNetworkPolicy gets deleted when you delete the FQDNNetworkPolicy.
To prevent this behavior, set the `fqdnnetworkpolicies.networking.gke.io/delete-policy` annotation to `abandon` on the
NetworkPolicy.

## Limitations

There are a few limitations to FQDNNetworkPolicies:

* Only *egress* rules are supported.
* Only *hostnames* are supported. In particular, you can't configure a FQDNNetworkPolicy with:
  * IPs or CIDRs. Use NetworkPolicies directly for that.
  * wildcard hostnames like `*.example.com`.
* Only A and CNAME records are supported. In particular, AAAA records for IPv6 are not supported.
* Records defined in the `/etc/hosts` file are not supported. Those records are probably static, so we recommend you use
  a normal `NetworkPolicy` for them.

## Development

This project is built with [kubebuilder](https://book.kubebuilder.io/introduction.html).
We recommend using [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) for your development environment.

### Prerequisites
You need the following tools installed on your development workstation.
* Docker
* kubectl
* Kind
* kustomize
* kubebuilder (2.3.1, you may need to export the [KUBEBUILDER_ASSET variable](https://book.kubebuilder.io/quick-start.html))

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