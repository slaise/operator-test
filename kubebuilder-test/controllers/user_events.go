package controllers

import (
	"context"
	"encoding/json"

	"reflect"
	"sync"

	"cloud.google.com/go/pubsub"
	v2 "example.com/m/api/v2"
	"github.com/go-logr/logr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type UserEvents struct {
	ctx          context.Context
	log          logr.Logger
	client       client.Client
	subscription *pubsub.Subscription
	lock         sync.RWMutex
	users        chan<- event.GenericEvent
}

type Event struct {
	UserName string `json:"userName,omitempty"`
	Project  string `json:"project,omitempty"`
}

func CreateUserEvents(client client.Client, subscription *pubsub.Subscription, users chan<- event.GenericEvent) UserEvents {
	log := ctrl.Log.
		WithName("source").
		WithName(reflect.TypeOf(UserEvents{}).Name())
	return UserEvents{
		ctx:          context.Background(),
		log:          log,
		client:       client,
		subscription: subscription,
		lock:         sync.RWMutex{},
		users:        users,
	}
}

func (t *UserEvents) Run() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		err := t.subscribe()
		if err != nil {
			t.log.Error(err, "error subscribe event")
		}
	}
}

func (t *UserEvents) subscribe() error {
	return t.subscription.Receive(t.ctx, func(ctx context.Context, msg *pubsub.Message) {
		log := t.log.WithValues("messageId", msg.ID)
		userEvent := Event{}
		if err := json.Unmarshal(msg.Data, &userEvent); err != nil {
			log.Error(err, "unable to unmarshal event")
			msg.Nack()
			return
		}

		list := v2.UserIdentityV2List{}
		if err := t.client.List(context.Background(), &list); err != nil {
			log.Error(err, "unable to get UserIdentity")
			msg.Nack()
			return
		}

		for _, config := range list.Items {
			evt := event.GenericEvent{
				Meta: &config.ObjectMeta,
			}
			t.users <- evt
		}
		msg.Ack()
	})
}
