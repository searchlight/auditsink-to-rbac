package event

import (
	"encoding/json"
	"log"
	"os"
	"sort"

	"github.com/davecgh/go-spew/spew"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apiserver/pkg/apis/audit"
)

type Event struct {
	Level      string   `json:"level"`
	AuditID    string   `json:"auditID"`
	Stage      string   `json:"stage"`
	RequestURI string   `json:"requestURI"`
	SourceIPs  []string `json:"sourceIPs"`
	Verb       string   `json:"verb"`

	Username  string   `json:"username"`
	UserGroup []string `json:"userGroup"`
	UserAgent string   `json:"userAgent"`

	ResponseCode int32 `json:"responseCode"`

	RequestReceivedTimestamp metav1.MicroTime `json:"requestReceivedTimestamp"`
	StageTimestamp           metav1.MicroTime `json:"stageTimestamp"`

	Annotations map[string]string `json:"annotations"`
}

type EventList struct {
	metav1.TypeMeta
	Items []Event `json:"items"`
}

func systemAuthenticated(event audit.Event) bool {
	authenticatedGroups := []string{"system:authenticated", "system:masters"}
	sort.Strings(authenticatedGroups)
	for _, group := range event.User.Groups {
		for _, authGroup := range authenticatedGroups {
			if group == authGroup {
				return true
			}
		}
	}
	return false
}

func ProcessEvents(eventBytes []byte) error {
	eventList := new(audit.EventList)

	if err := json.Unmarshal(eventBytes, eventList); err != nil {
		log.Println(err)
	}

	if systemAuthenticated(eventList.Items[0]) {
		return nil
	}

	newEventList := new(EventList)
	newEvent := Event{}
	newEventList.TypeMeta = eventList.TypeMeta

	for _, event := range eventList.Items {
		newEvent.Level = string(event.Level)
		newEvent.AuditID = string(event.AuditID)
		newEvent.Stage = string(event.Stage)
		newEvent.RequestURI = string(event.RequestURI)
		newEvent.SourceIPs = event.SourceIPs
		newEvent.Verb = string(event.Verb)
		newEvent.Username = event.User.Username
		newEvent.UserGroup = event.User.Groups
		newEvent.UserAgent = event.UserAgent
		newEvent.ResponseCode = event.ResponseStatus.Code
		newEvent.RequestReceivedTimestamp = event.StageTimestamp
		newEvent.StageTimestamp = event.StageTimestamp
		newEvent.Annotations = event.Annotations

		newEventList.Items = append(newEventList.Items, newEvent)

		//if newEvent.ResponseCode == 403 {
		//	oneliners.PrettyJson(newEvent)
		//} else {
		//	return nil
		//}
		spew.Dump(newEvent)
	}

	data, err := json.Marshal(newEventList)
	if err != nil {
		log.Println(err)
	}

	file, err := os.OpenFile("audit.log", os.O_APPEND|os.O_WRONLY, 0600)
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
