// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shoots_test

import (
	"context"
	"time"

	. "github.com/gardener/gardener/test/integration/shoots"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("#CIt", func() {

	var (
		defaultTimeout = time.Second * 1
		assertFoo      = func(inputOne string) func(ctx context.Context) {
			return func(ctx context.Context) {
				Expect(inputOne).To(Equal("foo"))
			}
		}
	)

	It("should panic when the first argument is not a function", func() {
		executable := func() {
			CIt("foo", "dummy", defaultTimeout)
		}
		Expect(executable).Should(Panic())
	})

	It("should panic when invoking without context", func() {
		executable := func() {
			CIt("foo", func() {}, defaultTimeout)
		}
		Expect(executable).Should(Panic())
	})

	It("should panic when invoking without sufficient arguments", func() {
		executable := func() {
			CIt("foo", func(ctx context.Context) {}, defaultTimeout, "crash")
		}
		Expect(executable).Should(Panic())
	})

	CIt("should succed when invoking with context and be closed", func(ctx context.Context) {
		done := ctx.Done()
		Expect(done).ShouldNot(BeClosed())
		Consistently(done).ShouldNot(Receive())
		Consistently(done).Should(BeClosed())
	}, time.Millisecond*5)

	CIt("should pass the correct values", func(ctx context.Context, inputOne, inputTwo string) {
		Expect(inputOne).To(Equal("foo"))
		Expect(inputTwo).To(Equal("bar"))
	}, defaultTimeout, "foo", "bar")

	CIt("should be able to share assertions", assertFoo("foo"), defaultTimeout)

})
