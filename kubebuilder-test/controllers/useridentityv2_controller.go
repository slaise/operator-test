/*
Copyright 2020 pc.

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
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	identityv2 "example.com/m/api/v2"
	"example.com/m/health"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/status"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// UserIdentityV2Reconciler reconciles a UserIdentityV2 object
type UserIdentityV2Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	health.HealthCheck
	PubsubClient *pubsub.Client
	Recorder     record.EventRecorder
}

const (
	Ready        status.ConditionType   = "Ready"
	UpToDate     status.ConditionReason = "UpToDate"
	UpdateFailed status.ConditionReason = "UpdateFailed"
)

// +kubebuilder:rbac:groups=identity.company.org,resources=useridentityv2s,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=identity.company.org,resources=useridentityv2s/status,verbs=get;update;patch

func (r *UserIdentityV2Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	log := r.Log.WithValues("useridentity", req.NamespacedName)

	// your logic here

	// update execution time
	r.HealthCheck.Trigger()

	var userIdentity identityv2.UserIdentityV2
	if err := r.Get(ctx, req.NamespacedName, &userIdentity); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	user := "jenny"      // pretend we get the name
	project := "project" // pretend we get the project name

	log.V(10).Info(fmt.Sprintf("Create Resources for User:%s, Project:%s", user, project))

	var serviceAccount corev1.ServiceAccount
	serviceAccount.Name = "default"
	annotations := make(map[string]string, 1)
	annotations["iam.gke.io/gcp-service-account"] = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", user, project)
	serviceAccount.Annotations = annotations
	_, err := ctrl.CreateOrUpdate(ctx, r.Client, &serviceAccount, func() error {
		return ctrl.SetControllerReference(&userIdentity, &serviceAccount, r.Scheme)
	})

	if err != nil {
		log.Error(err, fmt.Sprintf("Error create ServiceAccount for user: %s, project: %s", user, project))
		_ = r.SetConditionFail(ctx, err, userIdentity, log)
		return ctrl.Result{}, nil
	}

	log.V(10).Info(fmt.Sprintf("Create ServiceAccount for User:%s, Project:%s finished", user, project))

	var clusterRoleBinding rbacv1.ClusterRoleBinding
	clusterRoleBinding.Name = req.Name
	clusterRoleBinding.Namespace = req.Namespace
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &clusterRoleBinding, func() error {
		clusterRoleBinding.RoleRef = userIdentity.Spec.RoleRef

		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "default",
			},
		}
		return ctrl.SetControllerReference(&userIdentity, &clusterRoleBinding, r.Scheme)
	})

	if err != nil {
		log.Error(err, fmt.Sprintf("Error create ClusterRoleBinding for user: %s, project: %s", user, project))
		return ctrl.Result{}, nil
	}

	log.V(10).Info(fmt.Sprintf("Create ClusterRoleBinding for User:%s, Project:%s finished", user, project))
	conditions := userIdentity.GetConditions()
	condition := status.Condition{
		Type:    Ready,
		Status:  v1.ConditionTrue,
		Reason:  UpToDate,
		Message: err.Error(),
	}
	if conditions.SetCondition(condition) {
		if err := r.Status().Update(ctx, &userIdentity); err != nil {
			log.Error(err, "Set conditions failed")
			//r.Recorder.Event(userIdentity, corev1.EventTypeWarning, string(UpdateFailed), "Failed to update resource status")
			return ctrl.Result{}, err
		}
	}
	r.Recorder.Event(&userIdentity, corev1.EventTypeNormal, string(condition.Reason), condition.Message)
	return ctrl.Result{RequeueAfter: 10 * time.Minute}, nil
}

func (r *UserIdentityV2Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// define userevent and run
	ch := make(chan event.GenericEvent)
	subscription := r.PubsubClient.Subscription("userevent")
	userEvent := CreateUserEvents(mgr.GetClient(), subscription, ch)
	go userEvent.Run()

	return ctrl.NewControllerManagedBy(mgr).
		For(&identityv2.UserIdentityV2{}).
		Watches(&source.Channel{Source: ch, DestBufferSize: 1024}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func (r *UserIdentityV2Reconciler) SetConditionFail(ctx context.Context, err error, userIdentity identityv2.UserIdentityV2, log logr.Logger) error {
	conditions := userIdentity.GetConditions()
	condition := status.Condition{
		Type:    Ready,
		Status:  v1.ConditionFalse,
		Reason:  UpdateFailed,
		Message: err.Error(),
	}
	r.Recorder.Event(&userIdentity, corev1.EventTypeWarning, string(condition.Reason), condition.Message)
	if conditions.SetCondition(condition) {
		if err := r.Status().Update(ctx, &userIdentity); err != nil {
			log.Error(err, "Set conditions failed")
			r.Recorder.Event(&userIdentity, corev1.EventTypeWarning, string(UpdateFailed), "Failed to update resource status")
			return err
		}
	}
	return nil
}
