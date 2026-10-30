// SPDX-License-Identifier: BSD-3-Clause

package k8sutil

import (
	"testing"

	"github.com/go-openapi/testify/v2/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"

	nbv1 "github.com/netbirdio/kubernetes-operator/api/v1"
	nbv1alpha1 "github.com/netbirdio/kubernetes-operator/api/v1alpha1"
)

func TestInstanceOwns(t *testing.T) {
	t.Parallel()

	unlabelled := &corev1.Service{}
	cp2 := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{InstanceLabel: "cp2"}}}
	cp3 := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{InstanceLabel: "cp3"}}}

	require.True(t, Instance("").Owns(unlabelled))
	require.False(t, Instance("").Owns(cp2))
	require.False(t, Instance("cp2").Owns(unlabelled))
	require.True(t, Instance("cp2").Owns(cp2))
	require.False(t, Instance("cp2").Owns(cp3))
}

func TestInstanceLabels(t *testing.T) {
	t.Parallel()

	require.Empty(t, Instance("").Labels())
	require.Equal(t, map[string]string{InstanceLabel: "cp2"}, Instance("cp2").Labels())
}

func TestInstanceValidate(t *testing.T) {
	t.Parallel()

	require.NoError(t, Instance("").Validate())
	require.NoError(t, Instance("cp2").Validate())
	require.Error(t, Instance("not/valid").Validate())
}

func TestInstanceCacheByObject(t *testing.T) {
	t.Parallel()

	scheme := kruntime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, nbv1.AddToScheme(scheme))
	require.NoError(t, nbv1alpha1.AddToScheme(scheme))

	byObject, err := Instance("cp2").CacheByObject(scheme)
	require.NoError(t, err)

	kinds := map[string]bool{}
	for obj, opts := range byObject {
		gvks, _, err := scheme.ObjectKinds(obj)
		require.NoError(t, err)
		require.EqualT(t, "netbird.io", gvks[0].Group)
		require.EqualT(t, InstanceLabel+"=cp2", opts.Label.String())
		kinds[gvks[0].Kind] = true
	}
	require.True(t, kinds["NetworkRouter"])
	require.True(t, kinds["SetupKey"])
	require.True(t, kinds["NBRoutingPeer"])
	require.False(t, kinds["NetworkRouterList"])
	require.False(t, kinds["Service"])
}
