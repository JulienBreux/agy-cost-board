package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

type CloudRunServiceManifest struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string            `yaml:"name"`
		Labels      map[string]string `yaml:"labels"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"metadata"`
	Spec struct {
		Template struct {
			Metadata struct {
				Annotations map[string]string `yaml:"annotations"`
			} `yaml:"metadata"`
			Spec struct {
				ContainerConcurrency int `yaml:"containerConcurrency"`
				TimeoutSeconds       int `yaml:"timeoutSeconds"`
				Containers           []struct {
					Name  string `yaml:"name"`
					Image string `yaml:"image"`
					Ports []struct {
						Name          string `yaml:"name"`
						ContainerPort int    `yaml:"containerPort"`
					} `yaml:"ports"`
					Resources struct {
						Limits struct {
							CPU    string `yaml:"cpu"`
							Memory string `yaml:"memory"`
						} `yaml:"limits"`
					} `yaml:"resources"`
					Env []struct {
						Name  string `yaml:"name"`
						Value string `yaml:"value"`
					} `yaml:"env"`
					StartupProbe struct {
						HTTPGet struct {
							Path string `yaml:"path"`
							Port int    `yaml:"port"`
						} `yaml:"httpGet"`
					} `yaml:"startupProbe"`
					LivenessProbe struct {
						HTTPGet struct {
							Path string `yaml:"path"`
							Port int    `yaml:"port"`
						} `yaml:"httpGet"`
					} `yaml:"livenessProbe"`
				} `yaml:"containers"`
			} `yaml:"spec"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

func TestCloudRunDeployYaml(t *testing.T) {
	rootDir := findRepoRoot(t)
	deployYAMLPath := filepath.Join(rootDir, "deploy.yaml")

	data, err := os.ReadFile(deployYAMLPath)
	if err != nil {
		t.Fatalf("deploy.yaml does not exist at repo root: %v", err)
	}

	var manifest CloudRunServiceManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("failed to parse deploy.yaml as valid YAML: %v", err)
	}

	if manifest.APIVersion != "serving.knative.dev/v1" {
		t.Errorf("expected apiVersion 'serving.knative.dev/v1', got '%s'", manifest.APIVersion)
	}
	if manifest.Kind != "Service" {
		t.Errorf("expected kind 'Service', got '%s'", manifest.Kind)
	}
	if manifest.Metadata.Name != "agy-cost-board" {
		t.Errorf("expected metadata.name 'agy-cost-board', got '%s'", manifest.Metadata.Name)
	}

	if len(manifest.Spec.Template.Spec.Containers) == 0 {
		t.Fatalf("expected at least one container defined in deploy.yaml")
	}

	container := manifest.Spec.Template.Spec.Containers[0]
	if container.Name != "agy-cost-board" {
		t.Errorf("expected container name 'agy-cost-board', got '%s'", container.Name)
	}

	if len(container.Ports) == 0 || container.Ports[0].ContainerPort != 8080 {
		t.Errorf("expected container port 8080")
	}

	if container.Resources.Limits.CPU != "1" {
		t.Errorf("expected cpu limit '1', got '%s'", container.Resources.Limits.CPU)
	}
	if container.Resources.Limits.Memory != "512Mi" {
		t.Errorf("expected memory limit '512Mi', got '%s'", container.Resources.Limits.Memory)
	}

	// Check probes point to /healthz
	if container.StartupProbe.HTTPGet.Path != "/healthz" {
		t.Errorf("expected startupProbe path '/healthz', got '%s'", container.StartupProbe.HTTPGet.Path)
	}
	if container.LivenessProbe.HTTPGet.Path != "/healthz" {
		t.Errorf("expected livenessProbe path '/healthz', got '%s'", container.LivenessProbe.HTTPGet.Path)
	}

	// Check environment variables match app.json
	expectedEnvs := map[string]string{
		"PROJECT_ID":                     "PROJECT_ID",
		"AGY_COST_BOARD_DEMO":            "false",
		"AGY_COST_BOARD_TELEMETRY_TABLE": "",
		"AGY_COST_BOARD_BILLING_TABLE":   "",
	}

	envFound := make(map[string]bool)
	for _, env := range container.Env {
		envFound[env.Name] = true
		if expectedVal, ok := expectedEnvs[env.Name]; ok {
			if env.Value != expectedVal {
				t.Errorf("expected env %s to have value '%s', got '%s'", env.Name, expectedVal, env.Value)
			}
		}
	}

	for key := range expectedEnvs {
		if !envFound[key] {
			t.Errorf("expected env '%s' to be present in deploy.yaml", key)
		}
	}
}
