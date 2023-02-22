// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package predicates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func LabelSelectorPredicate(s metav1.LabelSelector) (predicate.Predicate, error) {
	selector, err := metav1.LabelSelectorAsSelector(&s)
	if err != nil {
		return predicate.Funcs{}, err
	}

	return NewPredicateFuncs(func(o client.Object) bool {
		return selector.Matches(labels.Set(o.GetLabels()))
	}), nil
}

func NewPredicateFuncs(filter func(object client.Object) bool) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return filter(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return filter(e.ObjectOld) || filter(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return filter(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return filter(e.Object)
		},
	}
}
