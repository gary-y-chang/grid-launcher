package simreq

import (
	"context"
	"fmt"
	"time"

	catrdg "github.com/crc-platform-engineering/grid-launcher/api/catrdg/v1"
	simreq "github.com/crc-platform-engineering/grid-launcher/api/simreq/v1"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// SimRequestReconciler reconciles a SimRequest object
type SimRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=simreq.cnri.platform,resources=simrequests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=simreq.cnri.platform,resources=simrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=simreq.cnri.platform,resources=simrequests/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=k8s.cni.cncf.io,resources=network-attachment-definitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubeovn.io,resources=vpcs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeovn.io,resources=subnets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SimRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *SimRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("SimRequestReconciler is running .....")

	//  Fetch the SimRequest CR instance
	simReq := &simreq.SimRequest{}
	err := r.Get(ctx, req.NamespacedName, simReq)
	if err != nil {
		if errors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			log.Info("SimRequest resource not found. Ignoring since object must be deleted or not created")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get SimRequest CR and requeue the request")
		return ctrl.Result{}, err
	}

	log.Info("Found the SimRequest CR instance.", "CR Name", simReq.Name, "Namespace", simReq.Namespace)

	// 1. Check if the Simreq is marked for deletion. If yes, do the cleanup and return ctrl.Result{}, nil
	if !simReq.ObjectMeta.DeletionTimestamp.IsZero() { // The object is being deleted
		// Handle cleanup and finalizer logic
		if containString(simReq.ObjectMeta.Finalizers, "simulation.cnri.platform/simrequest") {
			log.Info("Preparing to cleanup resources. This may take a few minutes.", "Namespace", simReq.Spec.Grid)
			/** 1. update the Namespace's annotation, removing simrequest_name */
			err = updateNamespace(ctx, r, simReq.Name, simReq.Spec.Grid, "remove")
			if err != nil {
				log.Error(err, "Failed to remove Namespace annotation", "simulation.cnri.platform/simrequest", simReq.Name)
				return ctrl.Result{}, err
			}

			/** 2. VMIs and NetworkAttachmentDefinition woulde be deleted automatically by the controller
			  since they are created in the grid namespace by the SimRequest  */

			/** 3. Delete Subnets, VPC created by the SimRequest */
			listOpts := []client.ListOption{
				client.MatchingLabels{"simrequest": simReq.Name, "grid": simReq.Spec.Grid},
			}

			// vpcs, _ := clientset.KubeovnV1().Vpcs().List(ctx, metav1.ListOptions{})
			vpcs := &kubeovnv1.VpcList{}
			r.List(ctx, vpcs, listOpts...)
			for _, v := range vpcs.Items {
				vpc := &kubeovnv1.Vpc{}
				err = r.Get(ctx, types.NamespacedName{Name: v.Name}, vpc)
				if err != nil {
					fmt.Println(err, "Failed to get VPC", "vpc_name", vpc.Name)
				}
				vpc.SetFinalizers([]string{})
				err := r.Update(ctx, vpc)
				// _, err := clientset.KubeovnV1().Vpcs().Update(ctx, &vpc, metav1.UpdateOptions{})
				if err != nil {
					fmt.Println(err, "Failed to update VPC", "vpc_name", vpc.Name)
				}
				time.Sleep(time.Second * 3) // Wait for being updated

				// Re-fetch updated version and delete
				u_vpc := &kubeovnv1.Vpc{}
				err = r.Get(ctx, types.NamespacedName{Name: v.Name}, u_vpc)
				if err != nil {
					fmt.Println(err, "Failed to re-fetch VPC", "vpc_name", u_vpc.Name)
				}

				err = r.Delete(ctx, u_vpc)
				if err != nil {
					fmt.Println(err, "Failed to delete Vpc", "vpc_name", u_vpc.Name)
				}
				// clientset.KubeovnV1().Vpcs().Delete(ctx, vpc.Name, metav1.DeleteOptions{})
				fmt.Println("Deleted VPC:", vpc.Name)
			}

			subnets := &kubeovnv1.SubnetList{}
			r.List(ctx, subnets, listOpts...)

			for _, s := range subnets.Items {
				f := len(s.Finalizers)
				for f != 0 {
					subnet := &kubeovnv1.Subnet{}
					err := r.Get(ctx, types.NamespacedName{Name: s.Name}, subnet)
					if err != nil {
						fmt.Println(err, "Failed to get Subnet", "subnet name", subnet.Name)
					}
					subnet.SetFinalizers([]string{})
					err = r.Update(ctx, subnet)
					if err != nil {
						fmt.Println(err, "Failed to update Subnet", "subnet_name", subnet.Name)
					}
					fmt.Println("Clear up Subnet finalizers:", s.Name, subnet.Finalizers)

					time.Sleep(time.Second * 3) // Wait for the subnet to be updated

					// Re-fetch updated version and delete
					u_subnet := &kubeovnv1.Subnet{}
					err = r.Get(ctx, types.NamespacedName{Name: s.Name}, u_subnet)
					if err != nil {
						fmt.Println(err, "Failed to re-fetch Subnet", "subnet name", subnet.Name)
					}
					fmt.Println("Re-fetch Subnet with finalizers:", u_subnet.Name, u_subnet.Finalizers)
					err = r.Delete(ctx, u_subnet)
					if err != nil {
						fmt.Println(err, "Failed to delete Subnet", "subnet name", subnet.Name)
					}
					f = len(u_subnet.Finalizers)
				}

				fmt.Println("Deleted Subnet:", s.Name)
			}

			// Remove finalizer after cleanup
			simReq.ObjectMeta.Finalizers = removeString(simReq.ObjectMeta.Finalizers, "simulation.cnri.platform/simrequest")
			if err := r.Update(ctx, simReq); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	}

	// 2. Find and fetch the Cartridge CR instance requested in the SimRequest
	cart := &catrdg.Cartridge{}
	if err := r.Get(ctx, types.NamespacedName{Name: simReq.Spec.CartridgeName, Namespace: simReq.Spec.Grid}, cart); err != nil {
		log.Info("Cartridge resource not found. Ignoring since object must be deleted or not created")
		// 2.1 the Cartridge CR is not found, mark the SimRequest Status to failed
		simReq.Status.Result = "Failed"
	} else {
		// 2.2. the Cartridge CR is found, proceed to launch the resources defined in the Cartridge CR
		// update the Namespace's annotation, adding simrequest_name
		err = updateNamespace(ctx, r, simReq.Name, simReq.Spec.Grid, "update")
		if err != nil {
			log.Error(err, "Failed to update Namespace annotation", "grid.cnri.platform/simrequest", simReq.Name)
			return ctrl.Result{}, err
		}

		// Launch requested resources (e.g., vpc, subnet, vmi ...)
		fmt.Printf("Cartridge resource name: %s found in namespace: %s.", cart.Name, cart.Namespace)
		valuesMap := valuesMerged(cart.Spec.Values, simReq.Spec.Values)
		values := convertMapToJsonString(valuesMap)

		dipenseNetworks(cart.Spec.Networks, valuesMap, simReq, ctx, r)

		dispenseHosts()
		// values, err := launchResource(ctx, r, simReq, cart.Spec)
		// if err != nil {
		// 	log.Error(err, "Failed to launch reuested resources")
		// 	return ctrl.Result{}, err
		// }

		// Update CR Status to indicate successful initiation.
		simReq.Status.Result = "Success"
		simReq.Status.StartTime = time.Now().String()
		simReq.Status.AppliedValues = values
	}

	// 3. Update the SimRequest status with the result of the simulation, and return
	if err := r.Status().Update(ctx, simReq); err != nil {
		log.Error(err, "Failed to update SimRequest status", "SimRequest", simReq.Name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("SimRequestReconciler")
	return ctrl.NewControllerManagedBy(mgr).
		For(&simreq.SimRequest{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				if e.ObjectNew.(*simreq.SimRequest).DeletionTimestamp.IsZero() {
					log.Info("---- SimRequest CR status updated, do nothing ----")
					return false
				}
				log.Info("---- SimRequest CR is being deleted, cleanup resources ----")
				return true
			},
			CreateFunc: func(e event.CreateEvent) bool {
				log.Info("SimRequest CR is creating ----")
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				log.Info("SimRequest CR deleted, do nothing ----")
				return true
			},
		})).
		Complete(r)
}
