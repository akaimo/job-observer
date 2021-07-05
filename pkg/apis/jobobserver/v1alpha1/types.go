package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cleaner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CleanerSpec `json:"spec"`
}

type CleanerSpec struct {
	TtlAfterFinished  string                `json:"ttlAfterFinished"`
	CleaningJobStatus string                `json:"cleaningJobStatus"`
	Selector          *metav1.LabelSelector `json:"selector"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CleanerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Cleaner `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Notificator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NotificatorSpec `json:"spec"`
}

type NotificatorSpec struct {
	Selector *metav1.LabelSelector `json:"selector"`
	Rule     Rule                  `json:"rule"`
	Receiver Receiver              `json:"receiver"`
}

type Rule struct {
	FinishingDeadline string `json:"finishingDeadline"`
}

type Receiver struct {
	SlackConfig SlackConfig `json:"slackConfig"`
}

type SlackConfig struct {
	ApiURL    *corev1.SecretKeySelector `json:"apiURL"`
	Channel   string `json:"channel"`
	Username  string `json:"username"`
	IconEmoji string `json:"iconEmoji"`
	IconURL   string `json:"iconURL"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NotificatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Notificator `json:"items"`
}
