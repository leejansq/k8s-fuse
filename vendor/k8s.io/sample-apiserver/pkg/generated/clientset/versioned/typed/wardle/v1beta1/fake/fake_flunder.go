/*
Copyright The Kubernetes Authors.

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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1beta1 "k8s.io/sample-apiserver/pkg/apis/wardle/v1beta1"
)

// FakeFlunders implements FlunderInterface
type FakeFlunders struct {
	Fake *FakeWardleV1beta1
	ns   string
}

var flundersResource = schema.GroupVersionResource{Group: "wardle.example.com", Version: "v1beta1", Resource: "flunders"}

var flundersKind = schema.GroupVersionKind{Group: "wardle.example.com", Version: "v1beta1", Kind: "Flunder"}

// Get takes name of the flunder, and returns the corresponding flunder object, and an error if there is any.
func (c *FakeFlunders) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.Flunder, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(flundersResource, c.ns, name), &v1beta1.Flunder{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Flunder), err
}

// List takes label and field selectors, and returns the list of Flunders that match those selectors.
func (c *FakeFlunders) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.FlunderList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(flundersResource, flundersKind, c.ns, opts), &v1beta1.FlunderList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.FlunderList{ListMeta: obj.(*v1beta1.FlunderList).ListMeta}
	for _, item := range obj.(*v1beta1.FlunderList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested flunders.
func (c *FakeFlunders) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(flundersResource, c.ns, opts))

}

// Create takes the representation of a flunder and creates it.  Returns the server's representation of the flunder, and an error, if there is any.
func (c *FakeFlunders) Create(ctx context.Context, flunder *v1beta1.Flunder, opts v1.CreateOptions) (result *v1beta1.Flunder, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(flundersResource, c.ns, flunder), &v1beta1.Flunder{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Flunder), err
}

// Update takes the representation of a flunder and updates it. Returns the server's representation of the flunder, and an error, if there is any.
func (c *FakeFlunders) Update(ctx context.Context, flunder *v1beta1.Flunder, opts v1.UpdateOptions) (result *v1beta1.Flunder, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(flundersResource, c.ns, flunder), &v1beta1.Flunder{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Flunder), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFlunders) UpdateStatus(ctx context.Context, flunder *v1beta1.Flunder, opts v1.UpdateOptions) (*v1beta1.Flunder, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(flundersResource, "status", c.ns, flunder), &v1beta1.Flunder{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Flunder), err
}

// Delete takes name of the flunder and deletes it. Returns an error if one occurs.
func (c *FakeFlunders) Delete(ctx context.Context, name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(flundersResource, c.ns, name), &v1beta1.Flunder{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFlunders) DeleteCollection(ctx context.Context, options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(flundersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.FlunderList{})
	return err
}

// Patch applies the patch and returns the patched flunder.
func (c *FakeFlunders) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Flunder, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(flundersResource, c.ns, name, pt, data, subresources...), &v1beta1.Flunder{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Flunder), err
}