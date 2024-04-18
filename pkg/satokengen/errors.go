// This code is licensed under the AGPL-3.0 License.
// https://github.com/grafana/grafana/blob/main/LICENSE
// Original code is from Grafana: https://github.com/grafana/grafana/tree/main/pkg/components/satokengen
// All credit for the work is given back to Grafana

package satokengen

type ErrInvalidApiKey struct {
}

func (e *ErrInvalidApiKey) Error() string {
	return "invalid API key"
}
