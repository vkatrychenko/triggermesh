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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/triggermesh/triggermesh/pkg/apis/sources/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CloudEventsSourceLister helps list CloudEventsSources.
// All objects returned here must be treated as read-only.
type CloudEventsSourceLister interface {
	// List lists all CloudEventsSources in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.CloudEventsSource, err error)
	// CloudEventsSources returns an object that can list and get CloudEventsSources.
	CloudEventsSources(namespace string) CloudEventsSourceNamespaceLister
	CloudEventsSourceListerExpansion
}

// cloudEventsSourceLister implements the CloudEventsSourceLister interface.
type cloudEventsSourceLister struct {
	indexer cache.Indexer
}

// NewCloudEventsSourceLister returns a new CloudEventsSourceLister.
func NewCloudEventsSourceLister(indexer cache.Indexer) CloudEventsSourceLister {
	return &cloudEventsSourceLister{indexer: indexer}
}

// List lists all CloudEventsSources in the indexer.
func (s *cloudEventsSourceLister) List(selector labels.Selector) (ret []*v1alpha1.CloudEventsSource, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CloudEventsSource))
	})
	return ret, err
}

// CloudEventsSources returns an object that can list and get CloudEventsSources.
func (s *cloudEventsSourceLister) CloudEventsSources(namespace string) CloudEventsSourceNamespaceLister {
	return cloudEventsSourceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CloudEventsSourceNamespaceLister helps list and get CloudEventsSources.
// All objects returned here must be treated as read-only.
type CloudEventsSourceNamespaceLister interface {
	// List lists all CloudEventsSources in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.CloudEventsSource, err error)
	// Get retrieves the CloudEventsSource from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.CloudEventsSource, error)
	CloudEventsSourceNamespaceListerExpansion
}

// cloudEventsSourceNamespaceLister implements the CloudEventsSourceNamespaceLister
// interface.
type cloudEventsSourceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CloudEventsSources in the indexer for a given namespace.
func (s cloudEventsSourceNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.CloudEventsSource, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CloudEventsSource))
	})
	return ret, err
}

// Get retrieves the CloudEventsSource from the indexer for a given namespace and name.
func (s cloudEventsSourceNamespaceLister) Get(name string) (*v1alpha1.CloudEventsSource, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("cloudeventssource"), name)
	}
	return obj.(*v1alpha1.CloudEventsSource), nil
}
