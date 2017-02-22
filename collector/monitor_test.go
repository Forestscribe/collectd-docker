package collector

import (
	"errors"
	"reflect"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

type fakeMonitorDockerClient struct {
	labels map[string]string
	env    []string
}

func (f fakeMonitorDockerClient) InspectContainer(id string) (*docker.Container, error) {
	return &docker.Container{
		Config: &docker.Config{
			Labels: f.labels,
			Env:    f.env,
		},
	}, nil
}

func (f fakeMonitorDockerClient) Stats(opts docker.StatsOptions) error {
	return errors.New("Stats() is not implemented for fake docker client")
}

func TestLabelExtraction(t *testing.T) {
	tests := map[*fakeMonitorDockerClient]struct {
		app  string
		task string
		err  error
	}{
		&fakeMonitorDockerClient{
			labels: map[string]string{},
			env:    []string{},
		}: {
			err: ErrNoNeedToMonitor,
		},

		// labels
		&fakeMonitorDockerClient{
			labels: map[string]string{
				appLabel: "myapp",
			},
			env: []string{},
		}: {
			app:  "myapp",
			task: defaultTask,
		},
		&fakeMonitorDockerClient{
			labels: map[string]string{
				appLabel:  "myapp",
				taskLabel: "mytask",
			},
			env: []string{},
		}: {
			app:  "myapp",
			task: "mytask",
		},
		&fakeMonitorDockerClient{
			labels: map[string]string{
				appLabel:  "my.app",
				taskLabel: "my.ta.sk",
			},
			env: []string{},
		}: {
			app:  "my.app",
			task: "my.ta.sk",
		},
		&fakeMonitorDockerClient{
			labels: map[string]string{
				appLabel:          "my.app",
				taskLocationLabel: "some_label",
			},
			env: []string{},
		}: {
			app:  "my.app",
			task: defaultTask,
		},
		&fakeMonitorDockerClient{
			labels: map[string]string{
				appLabel:          "my.app",
				taskLocationLabel: "some_label",
				"some_label":      "ho.ho.ho",
			},
			env: []string{},
		}: {
			app:  "my.app",
			task: "ho.ho.ho",
		},

		// mesos style labels
		&fakeMonitorDockerClient{
			labels: map[string]string{
				appLabel:  "/my/app",
				taskLabel: "/my/ta/sk",
			},
			env: []string{},
		}: {
			app:  "/my/app",
			task: "/my/ta/sk",
		},

		// env
		&fakeMonitorDockerClient{
			labels: map[string]string{},
			env: []string{
				appEnvPrefix + "myapp",
			},
		}: {
			app:  "myapp",
			task: defaultTask,
		},
		&fakeMonitorDockerClient{
			labels: map[string]string{},
			env: []string{
				appEnvPrefix + "my.app",
				taskEnvPrefix + "my.ta.sk",
			},
		}: {
			app:  "my.app",
			task: "my.ta.sk",
		},
		&fakeMonitorDockerClient{
			labels: map[string]string{},
			env: []string{
				appEnvPrefix + "myapp",
			},
		}: {
			app:  "myapp",
			task: defaultTask,
		},
		&fakeMonitorDockerClient{
			labels: map[string]string{},
			env: []string{
				appEnvPrefix + "my.app",
				taskEnvLocationPrefix + "MESOS_TASK_ID",
				"MESOS_TASK_ID=topface_prod-test_app.c80a053f-f66f-11e4-a977-56847afe9799",
			},
		}: {
			app:  "my.app",
			task: "topface_prod-test_app.c80a053f-f66f-11e4-a977-56847afe9799",
		},
		&fakeMonitorDockerClient{
			labels: map[string]string{},
			env: []string{
				appEnvPrefix + "my.app",
				taskEnvLocationPrefix + "MESOS_TASK_ID",
				taskEnvLocationTrimPrefix + "topface_prod-test_app.",
				"MESOS_TASK_ID=topface_prod-test_app.c80a053f-f66f-11e4-a977-56847afe9799",
			},
		}: {
			app:  "my.app",
			task: "c80a053f-f66f-11e4-a977-56847afe9799",
		},
	}

	for c, e := range tests {
		m, err := NewMonitor(c, "", 1)
		if err != nil {
			if err != e.err {
				t.Errorf("expected error %q instead of %q for %#v", e.err, err, c)
			}

			continue
		} else {
			if e.err != nil {
				t.Errorf("expected error %q for %#v, got nothing", e.err, c)
				continue
			}
		}

		if m.app != e.app {
			t.Errorf("expected app %s got %s for %#v", e.app, m.app, c)
		}

		if m.task != e.task {
			t.Errorf("expected task %s got %s for %#v", e.task, m.task, c)
		}
	}

}

func oneTestExtractTagsFromApp(t *testing.T, app string, exp map[string]string) {
	tags := map[string]string{}
	extractTagsFromApp(tags, app)
	if !reflect.DeepEqual(tags, exp) {
		t.Errorf("expected tags got %s for %s", tags, exp)
	}
}

// test for private function extractTagsFromApp
func TestExtractTagsFromApp(t *testing.T) {
	app := "/group/app"
	exp := map[string]string{
		"app_id": "/group/app",
		"app":    "app",
		"group":  "/group",
		"group1": "/group",
	}
	oneTestExtractTagsFromApp(t, app, exp)

	app = "/group/group2/app"
	exp = map[string]string{
		"app_id": "/group/group2/app",
		"app":    "app",
		"group":  "/group/group2",
		"group1": "/group",
		"group2": "/group/group2",
	}
	oneTestExtractTagsFromApp(t, app, exp)

	app = "/app"
	exp = map[string]string{
		"app_id": "/app",
		"app":    "app",
		"group":  "/",
	}
	oneTestExtractTagsFromApp(t, app, exp)

}

func oneTestExtractTagsFromTask(t *testing.T, task string, exp map[string]string) {
	tags := map[string]string{}
	extractTagsFromTask(tags, task)
	if !reflect.DeepEqual(tags, exp) {
		t.Errorf("expected tags got %s for %s", tags, exp)
	}
}

// test for private function extractTagsFromTask
func TestExtractTagsFromTask(t *testing.T) {
	task := "ci_artifactory_artifactory.9dc3c24c-f7a7-11e6-bae2-8eadd2be2132"
	exp := map[string]string{
		"task":  "ci_artifactory_artifactory.9dc3c24c-f7a7-11e6-bae2-8eadd2be2132",
		"task1": "ci_artifactory_artifactory",
		"task2": "9dc3c24c",
		"task3": "f7a7",
		"task4": "11e6",
		"task5": "bae2",
		"task6": "8eadd2be2132",
	}
	oneTestExtractTagsFromTask(t, task, exp)

}
