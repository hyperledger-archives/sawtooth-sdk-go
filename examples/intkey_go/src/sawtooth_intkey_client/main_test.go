/**
 * Copyright 2018 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * ------------------------------------------------------------------------------
 */

package main

import (
	"testing"
)

func Test_Sha512Compute(t *testing.T) {
	sha512OfIntkey := "1cf126a27a1fd9ba83e819ebff8c9f98ec8976f39cce35e984242a2ee40368fd8704c8f6147d653fb02c17e666cfc89c0ddc54d5b48d1cabdfed4ec124f99bdb"
	computedSha512OfIntkey := Sha512HashValue(FAMILY_NAME)
	if sha512OfIntkey != computedSha512OfIntkey {
		t.Errorf("Sha512 computed doesn't match \nExpected: %s \nGot: %s \n",
			sha512OfIntkey, computedSha512OfIntkey)
	}
}
