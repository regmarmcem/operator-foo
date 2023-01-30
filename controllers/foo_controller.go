/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	samplecontrollerv1alpha1 "github.com/regmarmcem/operator-foo/api/v1alpha1"
)

var log = logf.Log.WithName("controller_foo")

// FooReconciler reconciles a Foo object
type FooReconciler struct {
	client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=samplecontroller.regmarmcem.github.io,resources=foos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=samplecontroller.regmarmcem.github.io,resources=foos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=samplecontroller.regmarmcem.github.io,resources=foos/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Foo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
// Reconcile Loopは以下の4段階
// 0. 監視するObjectのEventが発生し、RequestがWorkqueueに入ることで発火
// 1. Foo Objectを取得
// 2. Fooが過去に管理していた古いDeploymentが存在したら削除
// 3. Fooが管理するDeploymentが存在しなければ作成
// 4. Fooが管理するDeploymentのSpecとFooのSpecを比較し、臨んだ状態でなければ調整
// 5. Fooのステータスを更新
func (r *FooReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling Foo")

	foo := &samplecontrollerv1alpha1.Foo{}
	if err := r.client.Get(ctx, req.NamespacedName, foo); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Foo not found. Ignore not found")
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "failed to get Foo")
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("foo-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &samplecontrollerv1alpha1.Foo{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(
		&source.Kind{Type: &appsv1.Deployment{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &samplecontrollerv1alpha1.Foo{},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FooReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&samplecontrollerv1alpha1.Foo{}).
		Complete(r)
}
