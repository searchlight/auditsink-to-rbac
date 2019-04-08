package event

import (
	"log"
	"os"

	"github.com/searchlight/auditsink-to-rbac/rbac"

	"github.com/the-redback/go-oneliners"

	"encoding/json"

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

	//if systemGenerated(eventList.Items[0]) {
	//	return nil
	//}

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

		newEvent.Username = event.ImpersonatedUser.Username
		newEvent.UserGroup = event.ImpersonatedUser.Groups
		newEvent.UserAgent = event.UserAgent

		newEvent.ResponseCode = event.ResponseStatus.Code
		newEvent.RequestReceivedTimestamp = event.StageTimestamp
		newEvent.StageTimestamp = event.StageTimestamp
		newEvent.Annotations = event.Annotations

		newEventList.Items = append(newEventList.Items, newEvent)

		if newEvent.ResponseCode != 403 {
			return nil
		}
	}

	if len(newEventList.Items) == 0 {
		return nil
	}

	if err := rbac.CreateRole(newEventList); err != nil {
		return err
	}

	oneliners.PrettyJson(newEventList)

	data, err := json.Marshal(newEventList)
	if err != nil {
		log.Println(err)
	}

	file, err := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	if _, err = file.Write(data); err != nil {
		log.Println(err)
		return err
	}
	_, _ = file.WriteString("\n")
	return nil
}
