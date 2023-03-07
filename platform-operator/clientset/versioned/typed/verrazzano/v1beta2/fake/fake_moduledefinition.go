// Copyright (c) 2020, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1beta2 "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeModuleDefinitions implements ModuleDefinitionInterface
type FakeModuleDefinitions struct {
	Fake *FakeVerrazzanoV1beta2
	ns   string
}

var moduledefinitionsResource = schema.GroupVersionResource{Group: "verrazzano", Version: "v1beta2", Resource: "moduledefinitions"}

var moduledefinitionsKind = schema.GroupVersionKind{Group: "verrazzano", Version: "v1beta2", Kind: "ModuleDefinition"}

// Get takes name of the moduleDefinition, and returns the corresponding moduleDefinition object, and an error if there is any.
func (c *FakeModuleDefinitions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta2.ModuleDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(moduledefinitionsResource, c.ns, name), &v1beta2.ModuleDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.ModuleDefinition), err
}

// List takes label and field selectors, and returns the list of ModuleDefinitions that match those selectors.
func (c *FakeModuleDefinitions) List(ctx context.Context, opts v1.ListOptions) (result *v1beta2.ModuleDefinitionList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(moduledefinitionsResource, moduledefinitionsKind, c.ns, opts), &v1beta2.ModuleDefinitionList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta2.ModuleDefinitionList{ListMeta: obj.(*v1beta2.ModuleDefinitionList).ListMeta}
	for _, item := range obj.(*v1beta2.ModuleDefinitionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested moduleDefinitions.
func (c *FakeModuleDefinitions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(moduledefinitionsResource, c.ns, opts))

}

// Create takes the representation of a moduleDefinition and creates it.  Returns the server's representation of the moduleDefinition, and an error, if there is any.
func (c *FakeModuleDefinitions) Create(ctx context.Context, moduleDefinition *v1beta2.ModuleDefinition, opts v1.CreateOptions) (result *v1beta2.ModuleDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(moduledefinitionsResource, c.ns, moduleDefinition), &v1beta2.ModuleDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.ModuleDefinition), err
}

// Update takes the representation of a moduleDefinition and updates it. Returns the server's representation of the moduleDefinition, and an error, if there is any.
func (c *FakeModuleDefinitions) Update(ctx context.Context, moduleDefinition *v1beta2.ModuleDefinition, opts v1.UpdateOptions) (result *v1beta2.ModuleDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(moduledefinitionsResource, c.ns, moduleDefinition), &v1beta2.ModuleDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.ModuleDefinition), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeModuleDefinitions) UpdateStatus(ctx context.Context, moduleDefinition *v1beta2.ModuleDefinition, opts v1.UpdateOptions) (*v1beta2.ModuleDefinition, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(moduledefinitionsResource, "status", c.ns, moduleDefinition), &v1beta2.ModuleDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.ModuleDefinition), err
}

// Delete takes name of the moduleDefinition and deletes it. Returns an error if one occurs.
func (c *FakeModuleDefinitions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(moduledefinitionsResource, c.ns, name, opts), &v1beta2.ModuleDefinition{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeModuleDefinitions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(moduledefinitionsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta2.ModuleDefinitionList{})
	return err
}

// Patch applies the patch and returns the patched moduleDefinition.
func (c *FakeModuleDefinitions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta2.ModuleDefinition, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(moduledefinitionsResource, c.ns, name, pt, data, subresources...), &v1beta2.ModuleDefinition{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta2.ModuleDefinition), err
}
