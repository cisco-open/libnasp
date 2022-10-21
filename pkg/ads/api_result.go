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

package ads

import (
	"context"
	"reflect"
	"sync"
)

// responseType lists the type constraint for all the API responses
type responseType interface {
	*httpClientPropertiesResponse | *listenerPropertiesResponse | *clientPropertiesResponse
}

type computeResponseFunc[I comparable, R responseType] func(I) R

// apiResultObservable emits API result changes to clients interested in API result changes for a given API input
type apiResultObservable[I comparable, R responseType] struct {

	// resultsByInput stores input value of API calls
	// and their corresponding result
	resultsByInput map[I]R

	// observerChansByInput channels through which observers interested to get updates on changes to api result
	// identified by api input
	observerChansByInput map[I][]chan<- struct{}

	// computeResponse is reference to the API call function for computing the
	// API result
	computeResponse         computeResponseFunc[I, R]
	recomputeResultFuncName string

	mu sync.RWMutex // sync access to fields being read and updated by multiple go routines
}

// registerForUpdates registers a client to get notified whenever the API result for the given input
// got updated. The notifications are sent through the given resultUpdated channel.
func (r *apiResultObservable[I, R]) registerForUpdates(input I) (<-chan struct{}, func()) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.observerChansByInput == nil {
		r.observerChansByInput = make(map[I][]chan<- struct{})
	}

	resultUpdatedChan := make(chan struct{}, 1)
	r.observerChansByInput[input] = append(r.observerChansByInput[input], resultUpdatedChan)

	response := r.computeResponse(input)

	cancel := func() {
		r.cancel(input, resultUpdatedChan)
	}
	if response == nil {
		return resultUpdatedChan, cancel
	}

	if r.resultsByInput == nil {
		r.resultsByInput = make(map[I]R)
	}

	r.resultsByInput[input] = response
	resultUpdatedChan <- struct{}{}

	return resultUpdatedChan, cancel
}

func (r *apiResultObservable[I, R]) cancel(input I, ch chan struct{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if observers, ok := r.observerChansByInput[input]; ok {
		found := -1
		for i := range observers {
			if observers[i] == ch {
				found = i
				break
			}
		}
		if found > -1 {
			observers = append(observers[:found], observers[found+1:]...)
			r.observerChansByInput[input] = observers
		}
	}

	close(ch)
}

func (r *apiResultObservable[I, R]) get(input I) (R, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res, ok := r.resultsByInput[input]
	return res, ok
}

func (r *apiResultObservable[I, R]) Refresh(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for input, observers := range r.observerChansByInput {
		if r.computeResponse != nil {
			newResponse := r.computeResponse(input)
			lastResponse := r.resultsByInput[input]
			if reflect.DeepEqual(lastResponse, newResponse) {
				continue
			}

			// api call result has changed since last invocation
			if r.resultsByInput == nil {
				r.resultsByInput = make(map[I]R)
			}
			r.resultsByInput[input] = newResponse

			// notify observers that there is an updated version of the API result for them
			for _, resultsUpdated := range observers {
				select {
				case <-ctx.Done():
					return
				case resultsUpdated <- struct{}{}:
				}
			}
		}
	}
}

func sendLatest[R any](ctx context.Context, data R, ch chan R) {
	// The channel may still hold data as consumer might have not read all items yet. Since consumer is
	// only interested in the latest version of the data we drain old data from the channel first before pushing
	// the new data
	for len(ch) > 0 {
		select {
		case <-ctx.Done(): // context has been terminated
			return
		case <-ch:
		}
	}
	select {
	case <-ctx.Done(): // context has been terminated
		return
	case ch <- data:
	}
}
