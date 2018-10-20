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

const (
	// String literals
	FAMILY_NAME       string = "intkey"
	FAMILY_VERSION    string = "1.0"
	DISTRIBUTION_NAME string = "sawtooth-intkey"
	DEFAULT_URL       string = "http://127.0.0.1:8008"
	// Verbs
	VERB_SET string = "set"
	VERB_INC string = "inc"
	VERB_DEC string = "dec"
	// APIs
	BATCH_SUBMIT_API string = "batches"
	BATCH_STATUS_API string = "batch_statuses"
	STATE_API        string = "state"
	// Content types
	CONTENT_TYPE_OCTET_STREAM string = "application/octet-stream"
	// Integer literals
	FAMILY_NAMESPACE_ADDRESS_LENGTH uint = 6
	FAMILY_VERB_ADDRESS_LENGTH      uint = 64
)
