// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package vault provides helper functions to improve the go-metrics to stackdriver metric
// conversions specific to HashiCorp Vault.
package vault

import "github.com/armon/go-metrics"

// Extractor extracts known patterns from the key into metrics.Label for better metric grouping
// and to help avoid the limit of 500 custom metric descriptors per project
// (https://cloud.google.com/monitoring/quotas).
func Extractor(key []string) ([]string, []metrics.Label, error) {
	// Metrics documented at https://www.vaultproject.io/docs/internals/telemetry.html should be
	// extracted here into a base metric name with appropriate labels extracted from the 'key'.
	switch len(key) {
	case 2: // metrics of format: *.*
		// database.<method>
		if key[0] == "database" {
			return key[:1], []metrics.Label{
				{
					Name:  "method",
					Value: key[1],
				},
			}, nil
		}
	case 3: // metrics of format: *.*.*
		// vault.token.create_root
		if key[0] == "vault" && key[1] == "token" && key[2] == "create_root" {
			return key, nil, nil
		}

		// vault.barrier.<method>
		// vault.token.<method>
		// vault.policy.<method>
		if key[0] == "vault" && (key[1] == "barrier" || key[1] == "token" || key[1] == "policy") {
			return key[:2], []metrics.Label{
				{
					Name:  "method",
					Value: key[2],
				},
			}, nil
		}
		// vault.<backend>.<method>
		if key[0] == "vault" && (key[2] == "put" || key[2] == "get" || key[2] == "delete" || key[2] == "list") {
			return key[:2], []metrics.Label{
				{
					Name:  "method",
					Value: key[2],
				},
			}, nil
		}
		// database.<name>.<method>
		// note: there are database.<method>.error counters. Those are handled separately.
		if key[0] == "database" && key[2] != "error" {
			return key[:1], []metrics.Label{
				{
					Name:  "name",
					Value: key[1],
				},
				{
					Name:  "method",
					Value: key[2],
				},
			}, nil
		}
		// database.<method>.error
		if key[0] == "database" && key[2] == "error" {
			return []string{"database", "error"}, []metrics.Label{
				{
					Name:  "method",
					Value: key[1],
				},
			}, nil
		}
	case 4: // metrics of format: *.*.*.*
		// vault.route.<method>.<mount>
		if key[0] == "vault" && key[1] == "route" {
			return key[:2], []metrics.Label{
				{
					Name:  "method",
					Value: key[2],
				},
				{
					Name:  "mount",
					Value: key[3],
				},
			}, nil
		}
		// vault.audit.<type>.*
		if key[0] == "vault" && key[1] == "audit" {
			return []string{"vault", "audit", key[3]}, []metrics.Label{
				{
					Name:  "type",
					Value: key[2],
				},
			}, nil
		}
		// vault.rollback.attempt.<mount>
		if key[0] == "vault" && key[1] == "rollback" && key[2] == "attempt" {
			return key[:3], []metrics.Label{
				{
					Name:  "mount",
					Value: key[3],
				},
			}, nil
		}
		// vault.<backend>.lock.<method>
		if key[0] == "vault" && key[2] == "lock" {
			return key[:3], []metrics.Label{
				{
					Name:  "method",
					Value: key[3],
				},
			}, nil
		}
		// database.<name>.<method>.error
		if key[0] == "database" && key[3] == "error" {
			return []string{key[0], key[3]}, []metrics.Label{
				{
					Name:  "name",
					Value: key[1],
				},
				{
					Name:  "method",
					Value: key[2],
				},
			}, nil
		}
	default:
		// unknown key pattern, keep it as-is.
	}
	return key, nil, nil
}

// Bucketer specifies the bucket boundaries that should be used for the given metric key.
func Bucketer(key []string) []float64 {
	// These were chosen to give some reasonable boundaires for RPC times in the 10-100ms range and
	// then rough values for 1-5 seconds.
	// TODO: investigate better boundaires for different metrics.
	return []float64{10.0, 25.0, 50.0, 100.0, 150.0, 200.0, 250.0, 300.0, 500.0, 1000.0, 1500.0, 2000.0, 3000.0, 4000.0, 5000.0}
}
