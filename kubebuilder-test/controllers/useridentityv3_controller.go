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
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	identityv2 "example.com/m/api/v2"
	"example.com/m/health"
	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	"github.com/operator-framework/operator-lib/status"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	identityv3 "example.com/m/api/v3"
)

// UserIdentityV3Reconciler reconciles a UserIdentityV3 object
type UserIdentityV3Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	health.HealthCheck
	PubsubClient *pubsub.Client
	Recorder     record.EventRecorder
}

type Param struct {
	User               string
	Project            string
	ServiceAccountName string
}

// +kubebuilder:rbac:groups=identity.company.org,resources=useridentityv3s,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=identity.company.org,resources=useridentityv3s/status,verbs=get;update;patch

func (r *UserIdentityV3Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	log := r.Log.WithValues("useridentity", req.NamespacedName)

	// your logic here

	// update execution time
	r.HealthCheck.Trigger()

	var userIdentity identityv3.UserIdentityV3
	if err := r.Get(ctx, req.NamespacedName, &userIdentity); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	user := "jenny"      // pretend we get the name
	project := "project" // pretend we get the project name

	log.V(10).Info(fmt.Sprintf("Create Resources for User:%s, Project:%s", user, project))

	t := template.New("")

	// Parse the template in the yamls
	for i, obj := range userIdentity.Spec.Template {
		jsonb, err := obj.MarshalJSON()
		if err != nil {
			log.Error(err, "unable to marshal object")
			return ctrl.Result{}, err
		}
		if _, err := t.New(fmt.Sprintf("%d", i)).
			Funcs(template.FuncMap{
				"serviceAccountName": func() string { return "default" },
				"user":               func(user string) string { return user },
				"project":            func(project string) string { return project },
			}).
			Parse(string(jsonb)); err != nil {
			log.Error(err, fmt.Sprintf("unable to parse template on index %d", i))
			return ctrl.Result{}, err
		}
	}

	// Go template
	renderTemplate := func(idx int, user string, project string) (*unstructured.Unstructured, error) {
		var buf bytes.Buffer
		if err := t.ExecuteTemplate(&buf, strconv.Itoa(idx), Param{
			User:               user,
			Project:            project,
			ServiceAccountName: "default",
		}); err != nil {
			return nil, err
		}
		decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		u := unstructured.Unstructured{}
		_, _, err := decoder.Decode(buf.Bytes(), nil, &u)
		return &u, err
	}

	// Apply resources
	for i := range userIdentity.Spec.Template {
		if err := func() error {
			rendered, err := renderTemplate(i, user, project)
			if err != nil {
				return err
			}

			var existing unstructured.Unstructured
			existing.SetGroupVersionKind(rendered.GroupVersionKind())
			existing.SetNamespace(rendered.GetNamespace())
			existing.SetName(rendered.GetName())

			log.V(10).Info(fmt.Sprintf("expanding %v into %v/%v", rendered.GroupVersionKind(), rendered.GetNamespace(), rendered.GetName()))

			_, err = ctrl.CreateOrUpdate(ctx, r.Client, &existing, func() error {
				if err := mergo.Merge(&existing, rendered, mergo.WithOverride); err != nil {
					return err
				}
				labels := existing.GetLabels()
				if labels == nil {
					labels = make(map[string]string)
				}

				return ctrl.SetControllerReference(&userIdentity, &existing, r.Scheme)
			})

			return err
		}(); err != nil {
			log.Error(err, fmt.Sprintf("Create resources for user:%s err", user))
			return ctrl.Result{}, nil
		}
	}

	log.V(10).Info(fmt.Sprintf("Create ClusterRoleBinding for User:%s, Project:%s finished", user, project))
	conditions := userIdentity.GetConditions()
	condition := status.Condition{
		Type:   Ready,
		Status: v1.ConditionTrue,
		Reason: UpToDate,
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

func (r *UserIdentityV3Reconciler) SetupWithManager(mgr ctrl.Manager) error {
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

func (r *UserIdentityV3Reconciler) SetConditionFail(ctx context.Context, err error, userIdentity identityv3.UserIdentityV3, log logr.Logger) error {
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
