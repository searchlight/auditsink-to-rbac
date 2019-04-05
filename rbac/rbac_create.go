package rbac

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/searchlight/auditsink-to-rbac/event"

	rbac "k8s.io/api/rbac/v1"
)

const (
	projectName = "auditsink-to-rbac"
)

func CreateRole(list event.EventList) error {
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
