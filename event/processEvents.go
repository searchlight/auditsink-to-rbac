package event

import (
	"log"

	"encoding/json"

	"github.com/the-redback/go-oneliners"

	"github.com/searchlight/auditsink-to-rbac/nats-streaming/publish"
	"github.com/searchlight/auditsink-to-rbac/system"

	"k8s.io/apiserver/pkg/apis/audit"
)

func systemGenerated(event audit.Event) bool {
	for _, systemUser := range system.UserList {
		if event.User.Username == systemUser {
			return true
		}
	}

	return event.RequestURI == system.SystemRequest
}

func ProcessEvents(eventBytes []byte) error {
	eventList := new(audit.EventList)

	if err := json.Unmarshal(eventBytes, eventList); err != nil {
		log.Println(err)
	}

	newEventList := system.EventList{}
	newEvent := system.Event{}
	newEventList.TypeMeta = eventList.TypeMeta

	for _, event := range eventList.Items {
		if event.ObjectRef == nil {
			continue
		} else if systemGenerated(event) {
			continue
		}

		newEvent.Level = string(event.Level)
		newEvent.AuditID = string(event.AuditID)
		newEvent.Stage = string(event.Stage)
		newEvent.RequestURI = event.RequestURI
		newEvent.SourceIPs = event.SourceIPs

		newEvent.ResourceUUID = string(event.ObjectRef.UID)
		newEvent.ResourceName = event.ObjectRef.Name
		newEvent.ResourceNamespace = event.ObjectRef.Namespace
		newEvent.ResourceGroup = event.ObjectRef.APIGroup
		newEvent.ResourceVersion = event.ObjectRef.APIVersion
		newEvent.ResourceKind = event.ObjectRef.Resource

		newEvent.Verb = event.Verb

		if event.ImpersonatedUser == nil {
			newEvent.Username = event.User.Username
			newEvent.UserGroup = event.User.Groups
		} else {
			newEvent.Username = event.ImpersonatedUser.Username
			newEvent.UserGroup = event.ImpersonatedUser.Groups
		}
		newEvent.UserAgent = event.UserAgent

		newEvent.ResponseCode = event.ResponseStatus.Code
		newEvent.RequestReceivedTimestamp = event.StageTimestamp
		newEvent.StageTimestamp = event.StageTimestamp
		newEvent.Annotations = event.Annotations

		newEventList.Items = append(newEventList.Items, newEvent)

	}

	if len(newEventList.Items) == 0 {
		return nil
	}

	oneliners.PrettyJson(newEventList)

	data, err := json.Marshal(newEventList)
	if err != nil {
		log.Println(err)
	}
	return publish.PublishToNatsServer(data)
}
