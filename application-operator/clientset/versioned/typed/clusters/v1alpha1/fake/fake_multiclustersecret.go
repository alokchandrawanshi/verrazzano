// Copyright (c) 2021, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/verrazzano/verrazzano/application-operator/apis/clusters/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMultiClusterSecrets implements MultiClusterSecretInterface
type FakeMultiClusterSecrets struct {
	Fake *FakeClustersV1alpha1
	ns   string
}

var multiclustersecretsResource = schema.GroupVersionResource{Group: "clusters.verrazzano.io", Version: "v1alpha1", Resource: "multiclustersecrets"}

var multiclustersecretsKind = schema.GroupVersionKind{Group: "clusters.verrazzano.io", Version: "v1alpha1", Kind: "MultiClusterSecret"}

// Get takes name of the multiClusterSecret, and returns the corresponding multiClusterSecret object, and an error if there is any.
func (c *FakeMultiClusterSecrets) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MultiClusterSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(multiclustersecretsResource, c.ns, name), &v1alpha1.MultiClusterSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MultiClusterSecret), err
}

// List takes label and field selectors, and returns the list of MultiClusterSecrets that match those selectors.
func (c *FakeMultiClusterSecrets) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MultiClusterSecretList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(multiclustersecretsResource, multiclustersecretsKind, c.ns, opts), &v1alpha1.MultiClusterSecretList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MultiClusterSecretList{ListMeta: obj.(*v1alpha1.MultiClusterSecretList).ListMeta}
	for _, item := range obj.(*v1alpha1.MultiClusterSecretList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested multiClusterSecrets.
func (c *FakeMultiClusterSecrets) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(multiclustersecretsResource, c.ns, opts))

}

// Create takes the representation of a multiClusterSecret and creates it.  Returns the server's representation of the multiClusterSecret, and an error, if there is any.
func (c *FakeMultiClusterSecrets) Create(ctx context.Context, multiClusterSecret *v1alpha1.MultiClusterSecret, opts v1.CreateOptions) (result *v1alpha1.MultiClusterSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(multiclustersecretsResource, c.ns, multiClusterSecret), &v1alpha1.MultiClusterSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MultiClusterSecret), err
}

// Update takes the representation of a multiClusterSecret and updates it. Returns the server's representation of the multiClusterSecret, and an error, if there is any.
func (c *FakeMultiClusterSecrets) Update(ctx context.Context, multiClusterSecret *v1alpha1.MultiClusterSecret, opts v1.UpdateOptions) (result *v1alpha1.MultiClusterSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(multiclustersecretsResource, c.ns, multiClusterSecret), &v1alpha1.MultiClusterSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MultiClusterSecret), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMultiClusterSecrets) UpdateStatus(ctx context.Context, multiClusterSecret *v1alpha1.MultiClusterSecret, opts v1.UpdateOptions) (*v1alpha1.MultiClusterSecret, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(multiclustersecretsResource, "status", c.ns, multiClusterSecret), &v1alpha1.MultiClusterSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MultiClusterSecret), err
}

// Delete takes name of the multiClusterSecret and deletes it. Returns an error if one occurs.
func (c *FakeMultiClusterSecrets) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(multiclustersecretsResource, c.ns, name, opts), &v1alpha1.MultiClusterSecret{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMultiClusterSecrets) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(multiclustersecretsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MultiClusterSecretList{})
	return err
}

// Patch applies the patch and returns the patched multiClusterSecret.
func (c *FakeMultiClusterSecrets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MultiClusterSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(multiclustersecretsResource, c.ns, name, pt, data, subresources...), &v1alpha1.MultiClusterSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MultiClusterSecret), err
}
