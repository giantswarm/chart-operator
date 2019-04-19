package helmclient

import (
	"fmt"

	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
)

// MergeValues merges config values so they can be used when installing or
// updating Helm releases. It takes in 2 maps with string keys and YAML values
// passed as a byte array.
//
// A deep merge is performed into a single map[string]interface{} output. If a
// value is present in both then the source map is preferred.
//
// Multiple keys with YAML values can be passed. If so the source and
// destination maps will be merged first and then merged together.
//
// The YAML values are parsed using yamlToStringMap. This is because the
// default behaviour of the YAML parser is to unmarshal into
// map[interface{}]interface{} which causes problems with the merge logic.
// See https://github.com/go-yaml/yaml/issues/139.
//
func MergeValues(destMap, srcMap map[string][]byte) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	mergedDestMap, err := mergeMapValues(destMap)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	mergedSrcMap, err := mergeMapValues(srcMap)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	result = mergeValues(mergedDestMap, mergedSrcMap)

	return result, nil
}

// mergeMapValues accepts a map with string keys and YAML values passed as a
// byte array. A deep merge is performed into a single map[string]interface{}
// output.
func mergeMapValues(inputMap map[string][]byte) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	for _, v := range inputMap {
		vals, err := yamlToStringMap(v)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		result = mergeValues(result, vals)
	}

	return result, nil
}

// mergeValues implements the merge logic. It performs a deep merge. If a value
// is present in both then the source map is preferred.
//
// Logic is based on the upstream logic implemented by Helm.
// https://github.com/helm/helm/blob/240e539cec44e2b746b3541529d41f4ba01e77df/cmd/helm/install.go#L358
func mergeValues(dest, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		if _, exists := dest[k]; !exists {
			// If the key doesn't exist already. Set the key to that value.
			dest[k] = v
			continue
		}

		nextMap, ok := v.(map[string]interface{})
		if !ok {
			// If it isn't another map. Overwrite the value.
			dest[k] = v
			continue
		}

		// Edge case: If the key exists in the destination but isn't a map.
		destMap, ok := dest[k].(map[string]interface{})
		if !ok {
			// If the source map has a map for this key. Prefer that value.
			dest[k] = v
			continue
		}

		// If we got to this point. It is a map in both so merge them.
		dest[k] = mergeValues(destMap, nextMap)
	}

	return dest
}

// yamlToStringMap unmarshals the YAML input into a map[string]interface{}
// with string keys. This is necessary because the default behaviour of the
// YAML parser is to return map[interface{}]interface{} types.
// See https://github.com/go-yaml/yaml/issues/139.
//
func yamlToStringMap(input []byte) (map[string]interface{}, error) {
	var raw interface{}
	var result map[string]interface{}

	err := yaml.Unmarshal(input, &raw)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	output := processMapValue(raw)
	result = output.(map[string]interface{})

	return result, nil
}

func processInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = processMapValue(v)
	}
	return res
}

func processInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = processMapValue(v)
	}
	return res
}

func processMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case bool:
		return v
	case float64:
		return v
	case int:
		return v
	case string:
		return v
	case []interface{}:
		return processInterfaceArray(v)
	case map[interface{}]interface{}:
		return processInterfaceMap(v)
	default:
		return microerror.Maskf(yamlConversionFailedError, "%#v with type %T not supported")
	}
}
