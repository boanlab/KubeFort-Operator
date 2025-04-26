// SPDX-License-Identifier: Apache-2.0
// Copyright 2025 BoanLab @ DKU

package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1 "kubefort-operator/api/v1"
)

// KubeFortPolicyReconciler reconciles a KubeFortPolicy object
type KubeFortPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.boanlab.com,resources=kubefortpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.boanlab.com,resources=kubefortpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.boanlab.com,resources=kubefortpolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KubeFortPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile

// validateProcessRule validates the ProcessRule according to the specified rules
func validateProcessRule(rule securityv1.ProcessRule) error {
	// Check if Name, Path, or Dir is set (only one should be set)
	count := 0
	if rule.Name != "" {
		count++
	}
	if rule.Path != "" {
		count++
	}
	if rule.Dir != "" {
		count++
	}
	if count != 1 {
		return fmt.Errorf("ProcessRule must have exactly one of Name, Path, or Dir set")
	}

	// Check if Recursive is set only when Dir is set
	if rule.Recursive && rule.Dir == "" {
		return fmt.Errorf("ProcessRule Recursive can only be set when Dir is set")
	}

	return nil
}

// validateFileRule validates the FileRule according to the specified rules
func validateFileRule(rule securityv1.FileRule) error {
	// Check if Name, Path, or Dir is set (only one should be set)
	count := 0
	if rule.Name != "" {
		count++
	}
	if rule.Path != "" {
		count++
	}
	if rule.Dir != "" {
		count++
	}
	if count != 1 {
		return fmt.Errorf("FileRule must have exactly one of Name, Path, or Dir set")
	}

	// Check if Recursive is set only when Dir is set
	if rule.Recursive && rule.Dir == "" {
		return fmt.Errorf("FileRule Recursive can only be set when Dir is set")
	}

	return nil
}

// validateCIDR validates if the given string is a valid IPv4 CIDR
func validateCIDR(cidr string) error {
	// Split CIDR into IP and mask
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid CIDR format: %s", cidr)
	}

	// Validate IP address
	ip := parts[0]
	ipParts := strings.Split(ip, ".")
	if len(ipParts) != 4 {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}

	for _, part := range ipParts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return fmt.Errorf("invalid IP address: %s", ip)
		}
	}

	// Validate mask
	mask, err := strconv.Atoi(parts[1])
	if err != nil || mask < 0 || mask > 32 {
		return fmt.Errorf("invalid mask: %s", parts[1])
	}

	return nil
}

// validateNetworkRule validates the NetworkRule according to the specified rules
func validateNetworkRule(rule securityv1.NetworkRule) error {
	// Check if TargetSelector or IPBlock is set (only one should be set)
	if len(rule.TargetSelector) > 0 && rule.IPBlock.CIDR != "" {
		return fmt.Errorf("NetworkRule cannot have both TargetSelector and IPBlock set")
	}
	if len(rule.TargetSelector) == 0 && rule.IPBlock.CIDR == "" {
		return fmt.Errorf("NetworkRule must have either TargetSelector or IPBlock set")
	}

	// Validate CIDR if IPBlock is set
	if rule.IPBlock.CIDR != "" {
		if err := validateCIDR(rule.IPBlock.CIDR); err != nil {
			return fmt.Errorf("invalid CIDR in IPBlock: %v", err)
		}
	}

	// Validate Except CIDRs
	for _, except := range rule.IPBlock.Except {
		if err := validateCIDR(except); err != nil {
			return fmt.Errorf("invalid CIDR in Except: %v", err)
		}
	}

	return nil
}

// normalizeDir ensures Dir ends with '/'
func normalizeDir(dir string) string {
	if dir != "" && !strings.HasSuffix(dir, "/") {
		return dir + "/"
	}
	return dir
}

func (r *KubeFortPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Get the KubeFortPolicy
	var policy securityv1.KubeFortPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Validate ProcessRules
	for _, rule := range policy.Spec.Process {
		if err := validateProcessRule(rule); err != nil {
			log.Error(err, "Invalid ProcessRule")
			return ctrl.Result{}, err
		}
		// Normalize Dir
		if rule.Dir != "" {
			rule.Dir = normalizeDir(rule.Dir)
		}
	}

	// Validate FileRules
	for _, rule := range policy.Spec.File {
		if err := validateFileRule(rule); err != nil {
			log.Error(err, "Invalid FileRule")
			return ctrl.Result{}, err
		}
		// Normalize Dir
		if rule.Dir != "" {
			rule.Dir = normalizeDir(rule.Dir)
		}
	}

	// Validate NetworkRules
	for _, rule := range policy.Spec.Network {
		if err := validateNetworkRule(rule); err != nil {
			log.Error(err, "Invalid NetworkRule")
			return ctrl.Result{}, err
		}
	}

	// Update policy status
	policy.Status.PolicyStatus = "Active"
	if err := r.Status().Update(ctx, &policy); err != nil {
		log.Error(err, "Failed to update policy status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeFortPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1.KubeFortPolicy{}).
		Named("kubefortpolicy").
		Complete(r)
}
