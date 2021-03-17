// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/miekg/dns"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkingv1alpha1 "github.com/GoogleCloudPlatform/gke-fqdnnetworkpolicies-golang/api/v1alpha1"

	networking "k8s.io/api/networking/v1"
)

// FQDNNetworkPolicyReconciler reconciles a FQDNNetworkPolicy object
type FQDNNetworkPolicyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	ownerAnnotation        = "fqdnnetworkpolicies.networking.gke.io/owned-by"
	deletePolicyAnnotation = "fqdnnetworkpolicies.networking.gke.io/delete-policy"
	finalizerName          = "finalizer.fqdnnetworkpolicies.networking.gke.io"
	// TODO make retry configurable
	retry = time.Second * time.Duration(10)
)

// +kubebuilder:rbac:groups=networking.gke.io,resources=fqdnnetworkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.gke.io,resources=fqdnnetworkpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies/status,verbs=get;update;patch

// Reconcile is reconciling a FQDNNetworkPolicy. It's run when there is a notification
// for a FQDNNetworkPolicy or after a given requeue time.
func (r *FQDNNetworkPolicyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("fqdnnetworkpolicy", req.NamespacedName)

	// retrieving the FQDNNetworkPolicy on which we are working
	fqdnNetworkPolicy := &networkingv1alpha1.FQDNNetworkPolicy{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, fqdnNetworkPolicy); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch FQDNNetworkPolicy")
		return ctrl.Result{}, err
	}

	if fqdnNetworkPolicy.ObjectMeta.DeletionTimestamp.IsZero() {
		// Our FQDNNetworkPolicy is not being deleted
		// If it doesn't already have our finalizer set, we set it
		if !containsString(fqdnNetworkPolicy.GetFinalizers(), finalizerName) {
			fqdnNetworkPolicy.SetFinalizers(append(fqdnNetworkPolicy.GetFinalizers(), finalizerName))
			if err := r.Update(context.Background(), fqdnNetworkPolicy); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Our FQDNNetworkPolicy is being deleted
		fqdnNetworkPolicy.Status.State = networkingv1alpha1.DestroyingState
		fqdnNetworkPolicy.Status.Reason = "Deleting NetworkPolicy"
		if e := r.Update(ctx, fqdnNetworkPolicy); e != nil {
			log.Error(e, "unable to update FQDNNetworkPolicy status")
			return ctrl.Result{}, e
		}

		if containsString(fqdnNetworkPolicy.GetFinalizers(), finalizerName) {
			// Our finalizer is set, so we need to delete the associated NetworkPolicy
			if err := r.deleteNetworkPolicy(ctx, fqdnNetworkPolicy); err != nil {
				return ctrl.Result{}, err
			}

			// deletion of the NetworkPolicy went well, we remove the finalizer from the list
			fqdnNetworkPolicy.SetFinalizers(removeString(fqdnNetworkPolicy.GetFinalizers(), finalizerName))
			fqdnNetworkPolicy.Status.Reason = "NetworkPolicy deleted or abandonned"
			if err := r.Update(context.Background(), fqdnNetworkPolicy); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Updating the NetworkPolicy associated with our FQDNNetworkPolicy
	// nextSyncIn represents when we should check in again on that FQDNNetworkPolicy.
	// It's probably related to the TTL of the DNS records.
	nextSyncIn, err := r.updateNetworkPolicy(ctx, fqdnNetworkPolicy)
	if err != nil {
		log.Error(err, "unable to update NetworkPolicy")
		fqdnNetworkPolicy.Status.State = networkingv1alpha1.PendingState
		fqdnNetworkPolicy.Status.Reason = err.Error()
		n := metav1.NewTime(time.Now().Add(retry))
		fqdnNetworkPolicy.Status.NextSyncTime = &n
		if e := r.Update(ctx, fqdnNetworkPolicy); e != nil {
			log.Error(e, "unable to update FQDNNetworkPolicy status")
			return ctrl.Result{}, e
		}
		return ctrl.Result{RequeueAfter: retry}, nil
	}
	log.Info("NetworkPolicy updated, next sync in " + fmt.Sprint(nextSyncIn))

	// Need to fetch the object again before updating it
	// as its status may have changed since the first time
	// we fetched it.
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, fqdnNetworkPolicy); err != nil {
		log.Error(err, "unable to fetch FQDNNetworkPolicy")
		return ctrl.Result{}, err
	}

	fqdnNetworkPolicy.Status.State = networkingv1alpha1.ActiveState
	nextSyncTime := metav1.NewTime(time.Now().Add(*nextSyncIn))
	fqdnNetworkPolicy.Status.NextSyncTime = &nextSyncTime

	// Updating the status of our FQDNNetworkPolicy
	if err := r.Update(ctx, fqdnNetworkPolicy); err != nil {
		log.Error(err, "unable to update FQDNNetworkPolicy status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: *nextSyncIn}, nil
}

func (r *FQDNNetworkPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mgr.GetFieldIndexer()
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.FQDNNetworkPolicy{}).
		Complete(r)
}

func (r *FQDNNetworkPolicyReconciler) updateNetworkPolicy(ctx context.Context,
	fqdnNetworkPolicy *networkingv1alpha1.FQDNNetworkPolicy) (*time.Duration, error) {
	log := r.Log.WithValues("fqdnnetworkpolicy", fqdnNetworkPolicy.Namespace+"/"+fqdnNetworkPolicy.Name)
	toCreate := false

	// Trying to fetch an existing NetworkPolicy of the same name as our FQDNNetworkPolicy
	networkPolicy := &networking.NetworkPolicy{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: fqdnNetworkPolicy.Namespace,
		Name:      fqdnNetworkPolicy.Name,
	}, networkPolicy); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// If there is none, that's OK, it means that we just haven't created it yet
			log.V(1).Info("associated NetworkPolicy doesn't exist, creating it")
			toCreate = true
		} else {
			return nil, err
		}
	}
	if !toCreate {
		log.V(2).Info("Found NetworkPolicy")
	}

	// If we have found a NetworkPolicy, but it doesn't have the right annotation
	// it means that it was created manually beforehand, and we don't want to touch it.
	// This also means that you can have a FQDNNetworkPolicy "adopt" a NetworkPolicy of the
	// same name by adding the correct annotation.
	if !toCreate && networkPolicy.Annotations[ownerAnnotation] != fqdnNetworkPolicy.Name {
		return nil, errors.New("NetworkPolicy missing owned-by annotation or owned by a different resource")
	}

	// Updating NetworkPolicy
	networkPolicy.Name = fqdnNetworkPolicy.Name
	networkPolicy.Namespace = fqdnNetworkPolicy.Namespace
	if networkPolicy.Annotations == nil {
		networkPolicy.Annotations = make(map[string]string)
	}
	networkPolicy.Annotations[ownerAnnotation] = fqdnNetworkPolicy.Name
	networkPolicy.Spec.PodSelector = fqdnNetworkPolicy.Spec.PodSelector
	rules, nextSync, err := r.getNetworkPolicyEgressRules(ctx, fqdnNetworkPolicy)
	if err != nil {
		return nil, err
	}
	networkPolicy.Spec.Egress = rules
	networkPolicy.Spec.PolicyTypes = []networking.PolicyType{networking.PolicyTypeEgress}

	// creating NetworkPolicy if needed
	if toCreate {
		if err := r.Create(ctx, networkPolicy); err != nil {
			log.Error(err, "unable to create NetworkPolicy")
			return nil, err
		}
	}
	// Updating the NetworkPolicy
	if err := r.Update(ctx, networkPolicy); err != nil {
		log.Error(err, "unable to update NetworkPolicy")
		return nil, err
	}

	return nextSync, nil
}

// deleteNetworkPolicy deletes the NetworkPolicy associated with the fqdnNetworkPolicy FQDNNetworkPolicy
func (r *FQDNNetworkPolicyReconciler) deleteNetworkPolicy(ctx context.Context,
	fqdnNetworkPolicy *networkingv1alpha1.FQDNNetworkPolicy) error {
	log := r.Log.WithValues("fqdnnetworkpolicy", fqdnNetworkPolicy.Namespace+"/"+fqdnNetworkPolicy.Name)

	// Trying to fetch an existing NetworkPolicy of the same name as our FQDNNetworkPolicy
	networkPolicy := &networking.NetworkPolicy{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: fqdnNetworkPolicy.Namespace,
		Name:      fqdnNetworkPolicy.Name,
	}, networkPolicy); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// If there is none, that's weird, but that's what we want
			log.Info("associated NetworkPolicy doesn't exist")
			return nil
		}
		return err
	}
	if networkPolicy.Annotations[deletePolicyAnnotation] == "abandon" {
		log.Info("NetworkPolicy has delete policy set to abandon, not deleting")
		return nil
	}
	if networkPolicy.Annotations[ownerAnnotation] != fqdnNetworkPolicy.Name {
		log.Info("NetworkPolicy is not owned by FQDNNetworkPolicy, not deleting")
		return nil
	}
	if err := r.Delete(ctx, networkPolicy); err != nil {
		log.Error(err, "unable to delete the NetworkPolicy")
		return err
	}
	log.Info("NetworkPolicy deleted")
	return nil
}

// getNetworkPolicyEgressRules returns a slice of NetworkPolicyEgressRules based on the
// provided slice of FQDNNetworkPolicyEgressRules, also returns when the next sync should happen
// based on the TTL of records
func (r *FQDNNetworkPolicyReconciler) getNetworkPolicyEgressRules(ctx context.Context, fqdnNetworkPolicy *networkingv1alpha1.FQDNNetworkPolicy) ([]networking.NetworkPolicyEgressRule, *time.Duration, error) {
	log := r.Log.WithValues("fqdnnetworkpolicy", fqdnNetworkPolicy.Namespace+"/"+fqdnNetworkPolicy.Name)
	fer := fqdnNetworkPolicy.Spec.Egress
	rules := []networking.NetworkPolicyEgressRule{}

	// getting the nameservers from the local /etc/resolv.conf
	ns, err := getNameservers()
	if err != nil {
		log.Error(err, "unable to get nameservers")
		return nil, nil, err
	}
	var nextSync uint32
	// Highest value possible for the resync time on the FQDNNetworkPolicy
	// TODO what should this be?
	nextSync = 30

	// TODO what do we do if nothing resolves, or if the list is empty?
	// What's the behavior of NetworkPolicies in that case?
	for _, frule := range fer {
		peers := []networking.NetworkPolicyPeer{}
		for _, to := range frule.To {
			for _, fqdn := range to.FQDNs {
				f := fqdn
				// The FQDN in the DNS request needs to end by a dot
				if l := fqdn[len(fqdn)-1]; l != '.' {
					f = fqdn + "."
				}
				c := new(dns.Client)
				c.SingleInflight = true

				// A records
				m := new(dns.Msg)
				m.SetQuestion(f, dns.TypeA)

				// TODO: We're always using the first nameserver. Should we do
				// something different? Note from Jens:
				// by default only if options rotate is set in resolv.conf
				// they are rotated. Otherwise the first is used, after a (5s)
				// timeout the next etc. So this is not too bad for now.
				r, _, err := c.Exchange(m, ns[0]+":53")
				if err != nil {
					log.Error(err, "unable to resolve "+f)
					continue
				}
				if len(r.Answer) == 0 {
					log.V(1).Info("could not find A record for " + f)
				}
				for _, ans := range r.Answer {
					if t, ok := ans.(*dns.A); ok {
						// Adding a peer per answer
						peers = append(peers, networking.NetworkPolicyPeer{
							IPBlock: &networking.IPBlock{CIDR: t.A.String() + "/32"}})
						// We want the next sync for the FQDNNetworkPolicy to happen
						// just after the TTL of the DNS record has expired.
						// Because a single FQDNNetworkPolicy may have different DNS
						// records with different TTLs, we pick the lowest one
						// and resynchronise after that.
						if ans.Header().Ttl < nextSync {
							nextSync = ans.Header().Ttl
						}
					}
				}

				// AAAA records
				m6 := new(dns.Msg)
				m6.SetQuestion(f, dns.TypeAAAA)

				// TODO: We're always using the first nameserver. Should we do
				// something different? Note from Jens:
				// by default only if options rotate is set in resolv.conf
				// they are rotated. Otherwise the first is used, after a (5s)
				// timeout the next etc. So this is not too bad for now.
				r6, _, err := c.Exchange(m6, ns[0]+":53")
				if err != nil {
					log.Error(err, "unable to resolve "+f)
					continue
				}
				if len(r6.Answer) == 0 {
					log.V(1).Info("could not find AAAA record for " + f)
				}
				for _, ans := range r6.Answer {
					if t, ok := ans.(*dns.AAAA); ok {
						// Adding a peer per answer
						peers = append(peers, networking.NetworkPolicyPeer{
							IPBlock: &networking.IPBlock{CIDR: t.AAAA.String() + "/128"}})
						// We want the next sync for the FQDNNetworkPolicy to happen
						// just after the TTL of the DNS record has expired.
						// Because a single FQDNNetworkPolicy may have different DNS
						// records with different TTLs, we pick the lowest one
						// and resynchronise after that.
						if ans.Header().Ttl < nextSync {
							nextSync = ans.Header().Ttl
						}
					}
				}
			}
		}

		if len(peers) == 0 {
			// If no peers have been found (most likely because the provided
			// FQDNs don't resolve to anything), then we don't create an egress
			// rule at all to fail close. If we create one with only a "ports"
			// section, but no "to" section, we're failing open.
			log.V(1).Info("No peers found, skipping egress rule.")
			continue
		}

		rules = append(rules, networking.NetworkPolicyEgressRule{
			Ports: frule.Ports,
			To:    peers,
		})
	}

	n := time.Second * time.Duration(nextSync)

	return rules, &n, nil
}
