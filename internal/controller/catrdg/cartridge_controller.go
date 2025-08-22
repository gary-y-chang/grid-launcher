package catrdg

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	catrdg "github.com/crc-platform-engineering/grid-launcher/api/catrdg/v1"
)

// CartridgeReconciler reconciles a Cartridge object
type CartridgeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catrdg.cnri.platform,resources=cartridges,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catrdg.cnri.platform,resources=cartridges/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=catrdg.cnri.platform,resources=cartridges/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cartridge object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *CartridgeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("CartridgeReconciler is running .....")

	// Fetch the Cartridge CR instance
	// catrdgv1.Cartridge
	cartridge := &catrdg.Cartridge{}
	err := r.Get(ctx, req.NamespacedName, cartridge)
	if err != nil {
		if errors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			log.Info("Cartridge resource not found")

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Cartridge CR and requeue the request")
		return ctrl.Result{}, err
	}
	log.Info("Found the Cartridge CR instance.", "Name", cartridge.Name, "Namespace", cartridge.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CartridgeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	log := ctrl.Log.WithName("CartridgeReconciler")
	return ctrl.NewControllerManagedBy(mgr).
		For(&catrdg.Cartridge{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				log.Info("Cartridge CR is updating ----", "name", e.ObjectNew.GetName())
				return false
			},
			CreateFunc: func(e event.CreateEvent) bool {
				log.Info("Cartridge CR is creating ----", "name", e.Object.GetName())
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				log.Info("Cartridge CR is deleted ----", "name", e.Object.GetName())
				return false
			},
		})).
		Complete(r)
}
