# FQDNNetworkPolicies

FQDNNetworkPolicies let you create Kubernetes Network Policies based on Fully
Qualified Domain Names(FQDNs) in addition to the standard functionality that
only allows IP address ranges (CIDR ranges). This implementation uses a custom
resource definition (CRD) and a controller inside the Kubernetes cluster that
periodically polls the DNS service and emits a NetworkPolicy object for every FQDNNetworkPolicy
object. 

## How does it work?

A FQDNNetworkPolicy looks a lot like a NetworkPolicy, but you can configure hostnames
in the "to" field:

```
apiVersion: networking.gke.io/v1alpha3
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

When you create this FQDNNetworkPolicy, the controller will in turn create a corresponding NetworkPolicy with
the same name, in the same namespace, that has the same `podSelector`, the same ports, but replacing
the hostnames with corresponding IP addresss it received by polling.

We recommend the use of [NodeLocal DNSCache](https://kubernetes.io/docs/tasks/administer-cluster/nodelocaldns/) to improve stability of records and reduce the number of DNS requests sent outside of the cluster.

**Note**: Just like with normal network policies, once specific pods are selected,
   all not explicitly allowed traffic is denied. Since FQDNNetworkPolicies are
   egress policies, we recommend to explicitly allow DNS traffic to allow name
   resolution. See
   [Kubernetes Network Policy Recipes](https://github.com/ahmetb/kubernetes-network-policy-recipes/blob/master/11-deny-egress-traffic-from-an-application.md#allowing-dns-traffic)

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

There might be scenarios where IPv6 AAAA lookups are not desired or supported in the resulting NetworkPolicy. 
To skip AAAA loopups set the `fqdnnetworkpolicies.networking.gke.io/aaaa-lookups` annotation to `skip`

## Limitations

There are a few functional limitations to FQDNNetworkPolicies:

* Only *hostnames* are supported. In particular, you can't configure a FQDNNetworkPolicy with:
  * IP addresses or CIDR blocks. Use NetworkPolicies directly for that.
  * wildcard hostnames like `*.example.com`.
* Only A, AAAA, and CNAME records are supported.
  * Google Cloud VPCs and GKE do not currently support IPv6, so AAAA records are not relevant in their context.
* Records defined in the `/etc/hosts` file are not supported. Those records are probably static, so we recommend you use
  a normal `NetworkPolicy` for them.
* When using an [IDN](https://en.wikipedia.org/wiki/Internationalized_domain_name),
  use the punycode equivalent as the locale used inside the controller might not
  be compatible with your locale.

### Use case limitations

The current controller implementation polls all domains from a single controller
instance in the Kubernetes cluster and repolls records after the TTL of the
first record expires. This leads to the following use case limitations in the
current implementation:

-  Since traffic blocking is implemented using existing Network Policies,
   FQDNNetworkPolicies is not a Layer 7 firewall but only blocks traffic based
   on IP addresses. If you need actual traffic filtering based on specific
   domains, use a proxy-based solution or consider
   [Egress gateways](https://istio.io/latest/docs/tasks/traffic-management/egress/egress-gateway/)
   in Istio for traffic filtering. An example where Layer 7 firewall behaviour
   differs from NetworkPolicies is when multiple hosts are using the same IP
   address - this implementation allows all of them as soon as one host is allowed.
-  The current implementation does not intercept actual pod DNS requests
   unlike CNI based solutions such as
   [CiliumNetworkPolicy](https://docs.cilium.io/en/v1.9/concepts/kubernetes/policy/#ciliumnetworkpolicy).
   It relies on the controller to update the NetworkPolicy based on results of
   polling and repolling after TTL expires. Since there might be conditions where the new IP address is not yet allowed by NetworkPolicy, the use of  [exponential backoff](https://en.wikipedia.org/wiki/Exponential_backoff) when connecting is recommended.
-  Since polling is currently only done on a single host, don't use the
   current implementation for allowing access to hosts that dynamically return
   different A records on subsequent requests, as different hosts might get
   different results and results might not be cached. Examples of such dynamic
   hosts are www.google.com, www.googleapis.com www.facebook.com and services
   behind AWS Route53 or Elastic Load Balancing. 
-  The use of
   [NodeLocal DNSCache](https://kubernetes.io/docs/tasks/administer-cluster/nodelocaldns/)
   is recommend to improve stability of records and reduce the number of DNS
   requests sent outside of the cluster 
-  For allowing traffic to Google APIs on Google Cloud, use a
   [Private Google Access](https://cloud.google.com/vpc/docs/configure-private-google-access)
   option instead and
   [configure DNS accordingly](https://cloud.google.com/vpc/docs/configure-private-google-access#config-domain).
   Then allow the respective IP addresses using a Standard Network Policy

An upcoming controller implementation will optionally poll all domains on each
node in the cluster, which together with
[NodeLocal DNSCache](https://kubernetes.io/docs/tasks/administer-cluster/nodelocaldns/)
can improve stability for dynamic hosts.

## Installation

Follow these instructions to install the FQDNNetworkPolicies controller in your GKE cluster.

1. Install [cert-manager](https://cert-manager.io/docs/installation/kubernetes/).

   ```
   kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.8.0/cert-manager.yaml
   ```

1. Install the FQDNNetworkPolicy controller.

   ```
   export VERSION=$(curl https://storage.googleapis.com/fqdnnetworkpolicies-manifests/latest)
   kubectl apply -f https://storage.googleapis.com/fqdnnetworkpolicies-manifests/${VERSION}.yaml
   ```

## Upgrades

Upgrading in place from the `v1alpha1` API (used in the 0.1 release) to the
`v1alpha2` (introduced in the 0.2 release) is not supported. You'll need to
uninstall the controller, reinstall it, update your FQDNNetworkPolicies to the
`v1alpha2` API and recreate them.

In the same manner, upgrading in place from the `v1alpha2` API (used in the 0.2 release) to the
`v1alpha3` (introduced in the 0.3 release) is not supported. You'll need to
uninstall the controller, reinstall it, update your FQDNNetworkPolicies to the
`v1alpha3` API and recreate them.

## Uninstall

To uninstall the FQDNNetworkPolicies controller from your GKE cluster, delete the FQDNNetworkPolicies first,
and then remove the resources.
Replace `YOUR_VERSION` by the version you are using.

```
export VERSION=YOUR_VERSION
kubectl delete fqdnnetworkpolicies.networking.gke.io -A --all
kubectl delete -f https://storage.googleapis.com/fqdnnetworkpolicies-manifests/${VERSION}.yaml
```

## Development

This project is built with [kubebuilder](https://book.kubebuilder.io/introduction.html).
We recommend using [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) for your development environment.

### Prerequisites
You need the following tools installed on your development workstation.
* Docker
* kubectl
* Kind
* kustomize
* kubebuilder (3.5.0, you may need to export the [KUBEBUILDER_ASSET variable](https://book.kubebuilder.io/quick-start.html))

### Building and running locally

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
   kubectl apply -f config/samples/networking_v1alpha3_fqdnnetworkpolicy_invalid.yaml
   kubectl apply -f config/samples/networking_v1alpha3_fqdnnetworkpolicy_valid.yaml
   ```

1. Explore the Makefile for other available commands, and read the [kubebuilder book](https://book.kubebuilder.io/introduction.html).

### Creating a release

1. Tag the commit you want to mark as a release. We follow semantic versioning.
1. Push the tag to GitHub.
1. Create a release in GitHub.
1. If you want that release to be the new default one, run:
   ```
   VERSION=$YOUR_TAG make latest
   ```
