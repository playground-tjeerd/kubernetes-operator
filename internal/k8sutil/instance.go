// SPDX-License-Identifier: BSD-3-Clause

package k8sutil

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InstanceLabel lets operators for different management servers share a cluster.
const InstanceLabel = "netbird.io/operator-instance"

const netbirdGroup = "netbird.io"

// Instance names an operator; the empty one owns unlabelled resources so existing installs keep working.
type Instance string

func (i Instance) Validate() error {
	if i == "" {
		return nil
	}
	if errs := validation.IsValidLabelValue(string(i)); len(errs) > 0 {
		return fmt.Errorf("invalid operator instance %q: %s", i, strings.Join(errs, ", "))
	}
	return nil
}

func (i Instance) Selector() labels.Selector {
	op, values := selection.DoesNotExist, []string(nil)
	if i != "" {
		op, values = selection.Equals, []string{string(i)}
	}
	req, err := labels.NewRequirement(InstanceLabel, op, values)
	if err != nil {
		panic(err)
	}
	return labels.NewSelector().Add(*req)
}

// Labels returns the labels to set on resources this instance creates.
func (i Instance) Labels() map[string]string {
	if i == "" {
		return nil
	}
	return map[string]string{InstanceLabel: string(i)}
}

func (i Instance) Owns(obj client.Object) bool {
	return i.Selector().Matches(labels.Set(obj.GetLabels()))
}

// CacheByObject scopes the cache of every netbird.io kind to this instance.
func (i Instance) CacheByObject(scheme *kruntime.Scheme) (map[client.Object]cache.ByObject, error) {
	byObject := map[client.Object]cache.ByObject{}
	for gvk := range scheme.AllKnownTypes() {
		if gvk.Group != netbirdGroup || strings.HasSuffix(gvk.Kind, "List") {
			continue
		}
		obj, err := scheme.New(gvk)
		if err != nil {
			return nil, err
		}
		cObj, ok := obj.(client.Object)
		if !ok {
			continue
		}
		byObject[cObj] = cache.ByObject{Label: i.Selector()}
	}
	return byObject, nil
}
