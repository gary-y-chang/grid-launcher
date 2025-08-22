package simreq

import (
	"context"
	"fmt"

	catrdg "github.com/crc-platform-engineering/grid-launcher/api/catrdg/v1"
	simreq "github.com/crc-platform-engineering/grid-launcher/api/simreq/v1"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func dipenseNetworks(subnets []catrdg.Subnet, values map[string]interface{}, simReq *simreq.SimRequest, ctx context.Context, r *SimRequestReconciler) error {
	/** 1. create one VPC original yaml:
	    apiVersion: kubeovn.io/v1
		kind: Vpc
	    metadata:
	      name: vpc-{{ .namespace }}
	    spec:
	      namespaces:
	        - {{ .namespace }}
	*/
	log := log.FromContext(ctx)
	log.Info("start create VPC")
	vpc := &kubeovnv1.Vpc{}
	vpc.SetName(fmt.Sprintf("vpc-%s", simReq.Spec.Grid))
	vpc.Spec.Namespaces = append(vpc.Spec.Namespaces, simReq.Spec.Grid)
	vpc.SetLabels(map[string]string{"simrequest": simReq.Name, "grid": simReq.Spec.Grid})

	existingVpc := &kubeovnv1.Vpc{}
	err := r.Get(ctx, types.NamespacedName{Name: vpc.Name}, existingVpc)
	if err != nil && k8serror.IsNotFound(err) {
		err = r.Create(ctx, vpc)
		if err != nil {
			log.Error(err, "Failed to create new Vpc", "vpc_name", vpc.Name)
			return err
		}
		fmt.Printf("VPC created successfully.  VPC name %s\n", vpc.Name)
	}

	/** 2. create Subnets with related NewworkAttachmentDefinition  */
	for i, subnet := range subnets {
		fmt.Printf("Subnet-%d to be created %s", i, subnet)
		/*****
				apiVersion: kubeovn.io/v1
				kind: Subnet
				metadata:
				name: servers-subnet-{{ .namespace }}
				spec:
				vpc: vpc-{{ .namespace }}
				protocol: IPv4
				provider: servers-segment.{{ .namespace }}.ovn
				cidrBlock: 172.16.0.0/25
				gateway: 172.16.0.1
				namespaces:
					- {{ .namespace }}
				---
				apiVersion: k8s.cni.cncf.io/v1
		    	kind: NetworkAttachmentDefinition
		    	metadata:
		      	  name: servers-segment
		          namespace: {{ .namespace }}
		        spec:
		          config: '{
					"cniVersion": "0.3.0",
					"type": "kube-ovn",
					"server_socket": "/run/openvswitch/kube-ovn-daemon.sock",
					"provider": "servers-segment.{{ .namespace }}.ovn"
		          }'
				*******/
	}

	return nil
}

func dispenseHosts() {
	fmt.Println(" To do ------ Dispense Hosts")
}
