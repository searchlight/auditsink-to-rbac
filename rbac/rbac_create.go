package rbac

import (
	"os"
	"strings"

	"encoding/json"
	"io/ioutil"

	"github.com/appscode/go/encoding/yaml"
	"github.com/searchlight/auditsink-to-rbac/system"

	rbac "k8s.io/api/rbac/v1"
)

const (
	projectName     = "auditsink-to-rbac"
	RBACApiVersion  = "v1"
	RoleKind        = "Role"
	RoleBindingKind = "RoleBinding"
)

func getFileAndRoles(event system.Event) (*os.File, *rbac.Role, *rbac.RoleBinding, error) {
	role := new(rbac.Role)
	roleBinding := new(rbac.RoleBinding)

	roles := make([]string, 2)
	if _, err := os.Stat(event.Username + "-roles.yaml"); err == nil {
		data, err := ioutil.ReadFile(event.Username + "-roles.yaml")
		if err != nil {
			return nil, nil, nil, err
		}

		if len(data) != 0 {
			roles = strings.Split(string(data), "---")
			if err := yaml.Unmarshal([]byte(roles[0]), role); err != nil {
				return nil, nil, nil, err
			}
			if err := yaml.Unmarshal([]byte(roles[1]), roleBinding); err != nil {
				return nil, nil, nil, err
			}
		}
		if err := os.Remove(event.Username + "-roles.yaml"); err != nil {
			return nil, nil, nil, err
		}
	}

	file, err := os.OpenFile(event.Username+"-roles.yaml", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return file, nil, nil, err
	}

	role.APIVersion = rbac.GroupName + "/" + RBACApiVersion
	role.Kind = RoleKind

	roleBinding.APIVersion = rbac.GroupName + "/" + RBACApiVersion
	roleBinding.Kind = RoleBindingKind
	return file, role, roleBinding, nil
}

func ruleExists(rules []string, rule string) bool {
	for _, value := range rules {
		if value == rule {
			return true
		}
	}
	return false
}

func CreateRoleFromBytes(eventBytes []byte) error {
	eventList := system.EventList{}
	if err := json.Unmarshal(eventBytes, &eventList); err != nil {
		return err
	}
	return CreateRole(eventList)
}

func CreateRole(list system.EventList) error {

	for _, event := range list.Items {
		file, role, roleBinding, err := getFileAndRoles(event)
		defer file.Close()
		if err != nil {
			return err
		}

		role.Name = projectName + ":" + event.Username
		role.Namespace = event.ResourceNamespace
		role.Labels = map[string]string{
			projectName + "/user":   event.Username,
			projectName + "/source": "auditsink",
		}

		if role.Rules == nil {
			role.Rules = []rbac.PolicyRule{
				{
					Verbs:     []string{event.Verb},
					APIGroups: []string{event.ResourceGroup},
					Resources: []string{event.ResourceKind},
				},
			}
		} else {
			if !ruleExists(role.Rules[0].Verbs, event.Verb) {
				role.Rules[0].Verbs = append(role.Rules[0].Verbs, event.Verb)
			}
			if !ruleExists(role.Rules[0].APIGroups, event.ResourceGroup) {
				role.Rules[0].APIGroups = append(role.Rules[0].APIGroups, event.ResourceGroup)
			}
			if !ruleExists(role.Rules[0].Resources, event.ResourceKind) {
				role.Rules[0].Resources = append(role.Rules[0].Resources, event.ResourceKind)
			}
		}

		data, err := yaml.Marshal(role)
		if err != nil {
			return err
		}
		if _, err = file.Write(data); err != nil {
			return err
		}
		_, _ = file.WriteString("\n---\n")

		roleBinding.Name = projectName + ":" + event.Username
		roleBinding.Namespace = event.ResourceNamespace
		roleBinding.Labels = map[string]string{
			projectName + "/user":   event.Username,
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
				Name:      event.Username,
				Namespace: event.ResourceNamespace,
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
