/*
Copyright 2022 TriggerMesh Inc.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/triggermesh/triggermesh/pkg/apis/sources/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSolaceSources implements SolaceSourceInterface
type FakeSolaceSources struct {
	Fake *FakeSourcesV1alpha1
	ns   string
}

var solacesourcesResource = schema.GroupVersionResource{Group: "sources.triggermesh.io", Version: "v1alpha1", Resource: "solacesources"}

var solacesourcesKind = schema.GroupVersionKind{Group: "sources.triggermesh.io", Version: "v1alpha1", Kind: "SolaceSource"}

// Get takes name of the solaceSource, and returns the corresponding solaceSource object, and an error if there is any.
func (c *FakeSolaceSources) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.SolaceSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(solacesourcesResource, c.ns, name), &v1alpha1.SolaceSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolaceSource), err
}

// List takes label and field selectors, and returns the list of SolaceSources that match those selectors.
func (c *FakeSolaceSources) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SolaceSourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(solacesourcesResource, solacesourcesKind, c.ns, opts), &v1alpha1.SolaceSourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SolaceSourceList{ListMeta: obj.(*v1alpha1.SolaceSourceList).ListMeta}
	for _, item := range obj.(*v1alpha1.SolaceSourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested solaceSources.
func (c *FakeSolaceSources) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(solacesourcesResource, c.ns, opts))

}

// Create takes the representation of a solaceSource and creates it.  Returns the server's representation of the solaceSource, and an error, if there is any.
func (c *FakeSolaceSources) Create(ctx context.Context, solaceSource *v1alpha1.SolaceSource, opts v1.CreateOptions) (result *v1alpha1.SolaceSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(solacesourcesResource, c.ns, solaceSource), &v1alpha1.SolaceSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolaceSource), err
}

// Update takes the representation of a solaceSource and updates it. Returns the server's representation of the solaceSource, and an error, if there is any.
func (c *FakeSolaceSources) Update(ctx context.Context, solaceSource *v1alpha1.SolaceSource, opts v1.UpdateOptions) (result *v1alpha1.SolaceSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(solacesourcesResource, c.ns, solaceSource), &v1alpha1.SolaceSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolaceSource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSolaceSources) UpdateStatus(ctx context.Context, solaceSource *v1alpha1.SolaceSource, opts v1.UpdateOptions) (*v1alpha1.SolaceSource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(solacesourcesResource, "status", c.ns, solaceSource), &v1alpha1.SolaceSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolaceSource), err
}

// Delete takes name of the solaceSource and deletes it. Returns an error if one occurs.
func (c *FakeSolaceSources) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(solacesourcesResource, c.ns, name, opts), &v1alpha1.SolaceSource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSolaceSources) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(solacesourcesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SolaceSourceList{})
	return err
}

// Patch applies the patch and returns the patched solaceSource.
func (c *FakeSolaceSources) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SolaceSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(solacesourcesResource, c.ns, name, pt, data, subresources...), &v1alpha1.SolaceSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SolaceSource), err
}
