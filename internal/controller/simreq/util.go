package simreq

import (
	"context"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func containString(sl []string, name string) bool {
	for _, value := range sl {
		if value == name {
			return true
		}
	}
	return false
}

func removeString(sl []string, s string) []string {
	for i, item := range sl {
		if item == s {
			return append(sl[:i], sl[i+1:]...)
		}
	}
	return sl
}

func updateNamespace(ctx context.Context, r *SimRequestReconciler, simReqName string, targetNS string, action string) error {
	log := log.FromContext(ctx)
	ns := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: targetNS}, ns)
	if err == nil {
		log.Info("Namespace found", "namespace", ns.Name)
		if action == "update" {
			// update the Namespace's label, adding simrequest_name
			ns.ObjectMeta.Labels["simrequest"] = simReqName
		} else if action == "remove" {
			// update the Namespace's label, removing simrequest_name
			delete(ns.Labels, "simrequest")
		}

		err = r.Update(ctx, ns)
		if err != nil {
			log.Error(err, "Failed to update Namespace annotation", "namespace", ns.Name)
			return err
		}
	}
	return nil
}

func valuesFromString(valueString string) (map[string]interface{}, error) {
	var values map[string]interface{}

	err := yaml.Unmarshal([]byte(valueString), &values)
	if err != nil {
		fmt.Printf("Error unmarshalling values from string: %v", err)
		return nil, err
	}
	return values, nil
}

func mergeMaps(deft, overwriting map[string]interface{}) map[string]interface{} {
	for k, v := range overwriting {
		// Check if the value from overwriting is a map, and merge if necessary.
		if vAsMap, ok := v.(map[string]interface{}); ok {
			if deft[k] == nil {
				deft[k] = make(map[string]interface{})
			}
			if aAsMap, ok := deft[k].(map[string]interface{}); ok {
				deft[k] = mergeMaps(aAsMap, vAsMap)
			} else {
				deft[k] = vAsMap
			}
		} else {
			// Overwrite the value in a with the value from b
			deft[k] = v
		}
	}
	return deft
}

func valuesMerged(valuesCartSpec string, valuesSimReqSpec string) map[string]interface{} {
	valuesFromCart, _ := valuesFromString(valuesCartSpec)
	valuesFromSimReq, _ := valuesFromString(valuesSimReqSpec)

	values := mergeMaps(valuesFromCart, valuesFromSimReq)
	return values
}

func convertMapToJsonString(data map[string]interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return ""
	}

	jsonString := string(jsonBytes)
	return jsonString
}
