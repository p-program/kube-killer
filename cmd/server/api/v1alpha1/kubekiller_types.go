package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeKillerSpec defines the desired state of KubeKiller
type KubeKillerSpec struct {
	// Mode defines the operation mode: "demon" or "illidan"
	// +kubebuilder:validation:Enum=demon;illidan
	// +kubebuilder:default=illidan
	Mode string `json:"mode,omitempty"`

	// Interval defines how often the killer should run (e.g., "5m", "1h")
	// +kubebuilder:default="5m"
	Interval string `json:"interval,omitempty"`

	// ScheduleAt defines a specific time to execute the deletion task (RFC3339 format)
	// If set, interval will be ignored and task will run only once at the specified time
	// +optional
	ScheduleAt *metav1.Time `json:"scheduleAt,omitempty"`

	// Namespaces to operate on. Empty means all namespaces except kube-system
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// ExcludeNamespaces namespaces to exclude from operations
	// +optional
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`

	// DryRun if true, only log what would be deleted without actually deleting
	// +kubebuilder:default=false
	DryRun bool `json:"dryRun,omitempty"`

	// Resources defines which resource types to kill
	// +optional
	Resources []string `json:"resources,omitempty"`
}

// KubeKillerStatus defines the observed state of KubeKiller
type KubeKillerStatus struct {
	// LastRunTime is the last time the killer ran
	// +optional
	LastRunTime *metav1.Time `json:"lastRunTime,omitempty"`

	// LastRunResult shows the result of the last run
	// +optional
	LastRunResult string `json:"lastRunResult,omitempty"`

	// ResourcesKilled is the number of resources killed in the last run
	// +optional
	ResourcesKilled int `json:"resourcesKilled,omitempty"`

	// Phase indicates the current phase of the operator
	// +optional
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Mode",type="string",JSONPath=".spec.mode"
// +kubebuilder:printcolumn:name="Interval",type="string",JSONPath=".spec.interval"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="LastRun",type="date",JSONPath=".status.lastRunTime"
// +kubebuilder:printcolumn:name="Killed",type="integer",JSONPath=".status.resourcesKilled"

// KubeKiller is the Schema for the kubekillers API
type KubeKiller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeKillerSpec   `json:"spec,omitempty"`
	Status KubeKillerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubeKillerList contains a list of KubeKiller
type KubeKillerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeKiller `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeKiller{}, &KubeKillerList{})
}
