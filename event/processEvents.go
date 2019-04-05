package event

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/the-redback/go-oneliners"

	"encoding/json"

	"github.com/searchlight/auditsink-to-rbac/system"
	"k8s.io/apiserver/pkg/apis/audit"

	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	projectName = "auditsink-to-rbac"
)

type Event struct {
	Level      string   `json:"level"`
	AuditID    string   `json:"auditID"`
	Stage      string   `json:"stage"`
	RequestURI string   `json:"requestURI"`
	SourceIPs  []string `json:"sourceIPs"`

	ResourceUUID      string `json:"resourceUUID"`
	ResourceName      string `json:"resourceName"`
	ResourceNamespace string `json:"resourceNamespace"`
	ResourceGroup     string `json:"resourceGroup"`
	ResourceVersion   string `json:"resourceVersion"`
	ResourceKind      string `json:"resourceKind"`

	Verb string `json:"verb"`

	Username  string   `json:"username"`
	UserGroup []string `json:"userGroup"`
	UserAgent string   `json:"userAgent"`

	ResponseCode int32 `json:"responseCode"`

	RequestReceivedTimestamp metav1.MicroTime `json:"requestReceivedTimestamp"`
	StageTimestamp           metav1.MicroTime `json:"stageTimestamp"`

	Annotations map[string]string `json:"annotations"`
}

type EventList struct {
	metav1.TypeMeta `json:",inline"`
	Items           []Event `json:"items"`
}

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

	newEventList := EventList{}
	newEvent := Event{}
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

	if err := CreateRole(newEventList); err != nil {
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

func CreateRole(list EventList) error {
	role := new(rbac.Role)
	roleBinding := new(rbac.RoleBinding)

	for _, Event := range list.Items {
		file, err := os.OpenFile(Event.Username+"-roles.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		role.Name = projectName + ":" + Event.Username
		role.Namespace = Event.ResourceNamespace
		role.Labels = map[string]string{
			projectName + "/user":   Event.Username,
			projectName + "/source": "auditsink",
		}

		role.Rules = []rbac.PolicyRule{
			{
				Verbs:     []string{Event.Verb},
				APIGroups: []string{Event.ResourceGroup},
				Resources: []string{Event.ResourceKind},
			},
		}
		data, err := yaml.Marshal(role)
		if err != nil {
			return err
		}
		if _, err = file.Write(data); err != nil {
			return err
		}
		_, _ = file.WriteString("\n---\n")

		roleBinding.Name = projectName + ":" + Event.Username
		roleBinding.Namespace = Event.ResourceNamespace
		roleBinding.Labels = map[string]string{
			projectName + "/user":   Event.Username,
			projectName + "/source": "auditsink",
		}
		roleBinding.RoleRef = rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     role.Kind,
			Name:     role.Name,
		}
		roleBinding.Subjects = []rbac.Subject{
			{
				Kind:      rbac.UserKind,
				APIGroup:  rbac.GroupName,
				Name:      Event.Username,
				Namespace: Event.ResourceNamespace,
			},
		}
		data, err = yaml.Marshal(roleBinding)
		if err != nil {
			return err
		}
		if _, err = file.Write(data); err != nil {
			return err
		}
	}

	return nil
}
