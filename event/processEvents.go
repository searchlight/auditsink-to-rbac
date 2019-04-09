package event

import (
	"encoding/json"
	"log"

	"github.com/the-redback/go-oneliners"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/apis/audit"

	"github.com/searchlight/auditsink-to-rbac/nats-streaming/publish"
	"github.com/searchlight/auditsink-to-rbac/system"
)

func systemGenerated(event audit.Event) bool {
	for _, systemUser := range system.UserList {
		if event.User.Username == systemUser {
			return true
		}
	}

	return event.RequestURI == system.SystemRequest
}

func getResourceUID(unknown *runtime.Unknown, reference *audit.ObjectReference, verb string) string {
	if verb != system.VerbCreate && verb != system.VerbDelete {
		return string(reference.UID)
	}

	var responseObject map[string]interface{}
	if err := json.Unmarshal(unknown.Raw, &responseObject); err != nil {
		return ""
	}

	var uidSource interface{}
	var exist bool
	if verb == system.VerbCreate {
		uidSource, exist = responseObject["metadata"]
	} else if verb == system.VerbDelete {
		uidSource, exist = responseObject["details"]
	}
	if !exist {
		return ""
	}

	uid, exist := uidSource.(map[string]interface{})["uid"]
	if !exist {
		return ""
	}

	return uid.(string)
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

		newEvent.Verb = event.Verb

		newEvent.ResourceUUID = getResourceUID(event.ResponseObject, event.ObjectRef, newEvent.Verb)
		newEvent.ResourceName = event.ObjectRef.Name
		newEvent.ResourceNamespace = event.ObjectRef.Namespace
		newEvent.ResourceGroup = event.ObjectRef.APIGroup
		newEvent.ResourceVersion = event.ObjectRef.APIVersion
		newEvent.ResourceKind = event.ObjectRef.Resource

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
