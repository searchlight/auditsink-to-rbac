package system

import (
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
