package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/pkg/listen"
	"github.com/digitalocean/doctl/pkg/terminal"
	"github.com/digitalocean/godo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAppsCommand(t *testing.T) {
	cmd := Apps()
	require.NotNil(t, cmd)
	assertCommandNames(t, cmd,
		"console",
		"create",
		"get",
		"list",
		"update",
		"delete",
		"dev",
		"create-deployment",
		"get-deployment",
		"list-deployments",
		"list-regions",
		"logs",
		"propose",
		"restart",
		"spec",
		"tier",
		"list-alerts",
		"update-alert-destinations",
		"list-buildpacks",
		"upgrade-buildpack",
	)
}

var (
	testAppSpec = godo.AppSpec{
		Name: "test",
		Services: []*godo.AppServiceSpec{
			{
				Name: "service",
				GitHub: &godo.GitHubSourceSpec{
					Repo:   "digitalocean/doctl",
					Branch: "main",
				},
			},
		},
	}

	testAppTier = &godo.AppTier{
		Name:                 "Test",
		Slug:                 "test",
		EgressBandwidthBytes: "10240",
		BuildSeconds:         "3000",
	}

	testAppInstanceSize = &godo.AppInstanceSize{
		Name:            "Basic XXS",
		Slug:            "basic-xxs",
		CPUType:         godo.AppInstanceSizeCPUType_Dedicated,
		CPUs:            "1",
		MemoryBytes:     "536870912",
		USDPerMonth:     "5",
		USDPerSecond:    "0.0000018896447",
		TierSlug:        "basic",
		TierUpgradeTo:   "professional-xs",
		TierDowngradeTo: "basic-xxxs",
	}

	testAlerts = []*godo.AppAlert{
		{
			ID: "c586fc0d-e8e2-4c50-9bf6-6c0a6b2ed2a7",
			Spec: &godo.AppAlertSpec{
				Rule: godo.AppAlertSpecRule_DeploymentFailed,
			},
			Emails: []string{"test@example.com", "test2@example.com"},
			SlackWebhooks: []*godo.AppAlertSlackWebhook{
				{
					URL:     "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
					Channel: "channel name",
				},
			},
		},
	}

	testAlert = godo.AppAlert{
		ID: "c586fc0d-e8e2-4c50-9bf6-6c0a6b2ed2a7",
		Spec: &godo.AppAlertSpec{
			Rule: godo.AppAlertSpecRule_DeploymentFailed,
		},
		Emails: []string{"test@example.com", "test2@example.com"},
		SlackWebhooks: []*godo.AppAlertSlackWebhook{
			{
				URL:     "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				Channel: "channel name",
			},
		},
	}

	testAlertUpdate = godo.AlertDestinationUpdateRequest{
		Emails: []string{"test@example.com", "test2@example.com"},
		SlackWebhooks: []*godo.AppAlertSlackWebhook{
			{
				URL:     "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				Channel: "channel name",
			},
		},
	}

	testBuildpack = &godo.Buildpack{
		Name:         "Go",
		ID:           "digitalocean/go",
		Version:      "1.2.3",
		MajorVersion: 1,
		Latest:       true,
		DocsLink:     "ftp://docs/go",
		Description:  []string{"Install Go"},
	}
)

func TestRunAppsCreate(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		specFile, err := os.CreateTemp(t.TempDir(), "spec")
		require.NoError(t, err)
		defer specFile.Close()

		err = json.NewEncoder(specFile).Encode(&testAppSpec)
		require.NoError(t, err)

		app := &godo.App{
			ID:        uuid.New().String(),
			Spec:      &testAppSpec,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		createReq := &godo.AppCreateRequest{
			Spec: &testAppSpec,
		}

		tm.apps.EXPECT().Create(createReq).Times(1).Return(app, nil)

		config.Doit.Set(config.NS, doctl.ArgAppSpec, specFile.Name())

		err = RunAppsCreate(config)
		require.NoError(t, err)
	})
}

func TestRunAppsGet(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		app := &godo.App{
			ID:        uuid.New().String(),
			Spec:      &testAppSpec,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		tm.apps.EXPECT().Get(app.ID).Times(1).Return(app, nil)

		config.Args = append(config.Args, app.ID)

		err := RunAppsGet(config)
		require.NoError(t, err)
	})
}

func TestRunAppsList(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		apps := []*godo.App{{
			ID:        uuid.New().String(),
			Spec:      &testAppSpec,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}}

		tm.apps.EXPECT().List(false).Times(1).Return(apps, nil)

		err := RunAppsList(config)
		require.NoError(t, err)
	})
}

func TestRunAppsUpdate(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		specFile, err := os.CreateTemp(t.TempDir(), "spec")
		require.NoError(t, err)
		defer specFile.Close()

		err = json.NewEncoder(specFile).Encode(&testAppSpec)
		require.NoError(t, err)

		app := &godo.App{
			ID:        uuid.New().String(),
			Spec:      &testAppSpec,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		updateReq := &godo.AppUpdateRequest{
			Spec: &testAppSpec,
		}

		tm.apps.EXPECT().Update(app.ID, updateReq).Times(1).Return(app, nil)

		config.Args = append(config.Args, app.ID)
		config.Doit.Set(config.NS, doctl.ArgAppSpec, specFile.Name())

		err = RunAppsUpdate(config)
		require.NoError(t, err)
	})
}

func TestRunAppsDelete(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		app := &godo.App{
			ID:        uuid.New().String(),
			Spec:      &testAppSpec,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		tm.apps.EXPECT().Delete(app.ID).Times(1).Return(nil)

		config.Args = append(config.Args, app.ID)
		config.Doit.Set(config.NS, doctl.ArgForce, true)

		err := RunAppsDelete(config)
		require.NoError(t, err)
	})
}

func TestRunAppsCreateDeployment(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		appID := uuid.New().String()
		deployment := &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
			Services: []*godo.DeploymentService{{
				Name:             "service",
				SourceCommitHash: "commit",
			}},
			Cause: "Manual",
			Phase: godo.DeploymentPhase_PendingDeploy,
			Progress: &godo.DeploymentProgress{
				PendingSteps: 1,
				RunningSteps: 0,
				SuccessSteps: 0,
				ErrorSteps:   0,
				TotalSteps:   1,

				Steps: []*godo.DeploymentProgressStep{{
					Name:      "name",
					Status:    "pending",
					StartedAt: time.Now(),
				}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		tm.apps.EXPECT().CreateDeployment(appID, true).Times(1).Return(deployment, nil)

		config.Args = append(config.Args, appID)
		config.Doit.Set(config.NS, doctl.ArgAppForceRebuild, true)

		err := RunAppsCreateDeployment(config)
		require.NoError(t, err)
	})
}

func TestRunAppsCreateDeploymentWithWait(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		appID := uuid.New().String()
		deployment := &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
			Services: []*godo.DeploymentService{{
				Name:             "service",
				SourceCommitHash: "commit",
			}},
			Cause: "Manual",
			Phase: godo.DeploymentPhase_PendingDeploy,
			Progress: &godo.DeploymentProgress{
				PendingSteps: 1,
				RunningSteps: 0,
				SuccessSteps: 0,
				ErrorSteps:   0,
				TotalSteps:   1,

				Steps: []*godo.DeploymentProgressStep{{
					Name:      "name",
					Status:    "pending",
					StartedAt: time.Now(),
				}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		activeDeployment := &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
			Services: []*godo.DeploymentService{{
				Name:             "service",
				SourceCommitHash: "commit",
			}},
			Cause: "Manual",
			Phase: godo.DeploymentPhase_Active,
			Progress: &godo.DeploymentProgress{
				PendingSteps: 1,
				RunningSteps: 0,
				SuccessSteps: 1,
				ErrorSteps:   0,
				TotalSteps:   1,

				Steps: []*godo.DeploymentProgressStep{{
					Name:      "name",
					Status:    "pending",
					StartedAt: time.Now(),
				}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		tm.apps.EXPECT().CreateDeployment(appID, false).Times(1).Return(deployment, nil)
		tm.apps.EXPECT().GetDeployment(appID, deployment.ID).Times(2).Return(activeDeployment, nil)

		config.Args = append(config.Args, appID)
		config.Doit.Set(config.NS, doctl.ArgCommandWait, true)

		err := RunAppsCreateDeployment(config)
		require.NoError(t, err)
	})
}

func TestRunAppsRestart(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		appID := uuid.New().String()
		deployment := &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
			Services: []*godo.DeploymentService{{
				Name:             "service",
				SourceCommitHash: "commit",
			}},
			Cause: "Manual",
			Phase: godo.DeploymentPhase_PendingDeploy,
			Progress: &godo.DeploymentProgress{
				PendingSteps: 1,
				RunningSteps: 0,
				SuccessSteps: 0,
				ErrorSteps:   0,
				TotalSteps:   1,

				Steps: []*godo.DeploymentProgressStep{{
					Name:      "name",
					Status:    "pending",
					StartedAt: time.Now(),
				}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		tm.apps.EXPECT().Restart(appID, []string{"component1", "component2"}).Times(1).Return(deployment, nil)

		config.Args = append(config.Args, appID)
		config.Doit.Set(config.NS, doctl.ArgAppComponents, []string{"component1", "component2"})

		err := RunAppsRestart(config)
		require.NoError(t, err)
	})
}

func TestRunAppsRestartWithWait(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		appID := uuid.New().String()
		deployment := &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
			Services: []*godo.DeploymentService{{
				Name:             "service",
				SourceCommitHash: "commit",
			}},
			Cause: "Manual",
			Phase: godo.DeploymentPhase_PendingDeploy,
			Progress: &godo.DeploymentProgress{
				PendingSteps: 1,
				RunningSteps: 0,
				SuccessSteps: 0,
				ErrorSteps:   0,
				TotalSteps:   1,

				Steps: []*godo.DeploymentProgressStep{{
					Name:      "name",
					Status:    "pending",
					StartedAt: time.Now(),
				}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		activeDeployment := &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
			Services: []*godo.DeploymentService{{
				Name:             "service",
				SourceCommitHash: "commit",
			}},
			Cause: "Manual",
			Phase: godo.DeploymentPhase_Active,
			Progress: &godo.DeploymentProgress{
				PendingSteps: 1,
				RunningSteps: 0,
				SuccessSteps: 1,
				ErrorSteps:   0,
				TotalSteps:   1,

				Steps: []*godo.DeploymentProgressStep{{
					Name:      "name",
					Status:    "pending",
					StartedAt: time.Now(),
				}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		tm.apps.EXPECT().Restart(appID, nil).Times(1).Return(deployment, nil)
		tm.apps.EXPECT().GetDeployment(appID, deployment.ID).Times(2).Return(activeDeployment, nil)

		config.Args = append(config.Args, appID)
		config.Doit.Set(config.NS, doctl.ArgCommandWait, true)

		err := RunAppsRestart(config)
		require.NoError(t, err)
	})
}

func TestRunAppsGetDeployment(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		appID := uuid.New().String()
		deployment := &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
			Services: []*godo.DeploymentService{{
				Name:             "service",
				SourceCommitHash: "commit",
			}},
			Cause: "Manual",
			Phase: godo.DeploymentPhase_PendingDeploy,
			Progress: &godo.DeploymentProgress{
				PendingSteps: 1,
				RunningSteps: 0,
				SuccessSteps: 0,
				ErrorSteps:   0,
				TotalSteps:   1,

				Steps: []*godo.DeploymentProgressStep{{
					Name:      "name",
					Status:    "pending",
					StartedAt: time.Now(),
				}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		tm.apps.EXPECT().GetDeployment(appID, deployment.ID).Times(1).Return(deployment, nil)

		config.Args = append(config.Args, appID, deployment.ID)

		err := RunAppsGetDeployment(config)
		require.NoError(t, err)
	})
}

func TestRunAppsListDeployments(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		appID := uuid.New().String()
		deployments := []*godo.Deployment{{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
			Services: []*godo.DeploymentService{{
				Name:             "service",
				SourceCommitHash: "commit",
			}},
			Cause: "Manual",
			Phase: godo.DeploymentPhase_PendingDeploy,
			Progress: &godo.DeploymentProgress{
				PendingSteps: 1,
				RunningSteps: 0,
				SuccessSteps: 0,
				ErrorSteps:   0,
				TotalSteps:   1,

				Steps: []*godo.DeploymentProgressStep{{
					Name:      "name",
					Status:    "pending",
					StartedAt: time.Now(),
				}},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}}

		tm.apps.EXPECT().ListDeployments(appID).Times(1).Return(deployments, nil)

		config.Args = append(config.Args, appID)

		err := RunAppsListDeployments(config)
		require.NoError(t, err)
	})
}

func TestRunAppsGetLogs(t *testing.T) {
	appID := uuid.New().String()
	deploymentID := uuid.New().String()
	component := "service"

	types := map[string]godo.AppLogType{
		"build":  godo.AppLogTypeBuild,
		"deploy": godo.AppLogTypeDeploy,
		"run":    godo.AppLogTypeRun,
	}

	for typeStr, logType := range types {
		withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
			tm.apps.EXPECT().GetLogs(appID, deploymentID, component, logType, true, 1).Times(1).Return(&godo.AppLogs{LiveURL: "https://proxy-apps-prod-ams3-001.ondigitalocean.app/?token=aa-bb-11-cc-33"}, nil)
			tm.listen.EXPECT().Listen(gomock.Any()).Times(1).Return(nil)

			tc := config.Doit.(*doctl.TestConfig)
			tc.ListenFn = func(url *url.URL, token string, schemaFunc listen.SchemaFunc, out io.Writer, in <-chan []byte) listen.ListenerService {
				assert.Equal(t, "aa-bb-11-cc-33", token)
				assert.Equal(t, "wss://proxy-apps-prod-ams3-001.ondigitalocean.app/?token=aa-bb-11-cc-33", url.String())
				return tm.listen
			}

			config.Args = append(config.Args, appID, component)
			config.Doit.Set(config.NS, doctl.ArgAppDeployment, deploymentID)
			config.Doit.Set(config.NS, doctl.ArgAppLogType, typeStr)
			config.Doit.Set(config.NS, doctl.ArgAppLogFollow, true)
			config.Doit.Set(config.NS, doctl.ArgAppLogTail, 1)

			err := RunAppsGetLogs(config)
			require.NoError(t, err)
		})
	}
}

func TestRunAppsGetLogsWithAppName(t *testing.T) {
	appName := "test-app"
	component := "service"

	testApp := &godo.App{
		ID:   uuid.New().String(),
		Spec: &testAppSpec,
		ActiveDeployment: &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
		},
	}

	types := map[string]godo.AppLogType{
		"build":  godo.AppLogTypeBuild,
		"deploy": godo.AppLogTypeDeploy,
		"run":    godo.AppLogTypeRun,
	}

	for typeStr, logType := range types {
		withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
			tm.apps.EXPECT().Find(appName).Times(1).Return(testApp, nil)
			tm.apps.EXPECT().GetLogs(testApp.ID, testApp.ActiveDeployment.ID, component, logType, true, 1).Times(1).Return(&godo.AppLogs{LiveURL: "https://proxy-apps-prod-ams3-001.ondigitalocean.app/?token=aa-bb-11-cc-33"}, nil)
			tm.listen.EXPECT().Listen(gomock.Any()).Times(1).Return(nil)

			tc := config.Doit.(*doctl.TestConfig)
			tc.ListenFn = func(url *url.URL, token string, schemaFunc listen.SchemaFunc, out io.Writer, in <-chan []byte) listen.ListenerService {
				assert.Equal(t, "aa-bb-11-cc-33", token)
				assert.Equal(t, "wss://proxy-apps-prod-ams3-001.ondigitalocean.app/?token=aa-bb-11-cc-33", url.String())
				return tm.listen
			}

			config.Args = append(config.Args, appName, component)
			config.Doit.Set(config.NS, doctl.ArgAppDeployment, "")
			config.Doit.Set(config.NS, doctl.ArgAppLogType, typeStr)
			config.Doit.Set(config.NS, doctl.ArgAppLogFollow, true)
			config.Doit.Set(config.NS, doctl.ArgAppLogTail, 1)

			err := RunAppsGetLogs(config)
			require.NoError(t, err)
		})
	}
}

func TestRunAppsGetLogsWithAppNameAndDeploymentID(t *testing.T) {
	appName := "test-app"
	component := "service"

	testApp := &godo.App{
		ID:   uuid.New().String(),
		Spec: &testAppSpec,
		ActiveDeployment: &godo.Deployment{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
		},
	}

	types := map[string]godo.AppLogType{
		"build":  godo.AppLogTypeBuild,
		"deploy": godo.AppLogTypeDeploy,
		"run":    godo.AppLogTypeRun,
	}

	for typeStr, logType := range types {
		withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
			tm.apps.EXPECT().Find(appName).Times(1).Return(testApp, nil)
			tm.apps.EXPECT().GetLogs(testApp.ID, testApp.ActiveDeployment.ID, component, logType, true, 1).Times(1).Return(&godo.AppLogs{LiveURL: "https://proxy-apps-prod-ams3-001.ondigitalocean.app/?token=aa-bb-11-cc-33"}, nil)
			tm.listen.EXPECT().Listen(gomock.Any()).Times(1).Return(nil)

			tc := config.Doit.(*doctl.TestConfig)
			tc.ListenFn = func(url *url.URL, token string, schemaFunc listen.SchemaFunc, out io.Writer, in <-chan []byte) listen.ListenerService {
				assert.Equal(t, "aa-bb-11-cc-33", token)
				assert.Equal(t, "wss://proxy-apps-prod-ams3-001.ondigitalocean.app/?token=aa-bb-11-cc-33", url.String())
				return tm.listen
			}

			config.Args = append(config.Args, appName, component)
			config.Doit.Set(config.NS, doctl.ArgAppDeployment, testApp.ActiveDeployment.ID)
			config.Doit.Set(config.NS, doctl.ArgAppLogType, typeStr)
			config.Doit.Set(config.NS, doctl.ArgAppLogFollow, true)
			config.Doit.Set(config.NS, doctl.ArgAppLogTail, 1)

			err := RunAppsGetLogs(config)
			require.NoError(t, err)
		})
	}
}

func TestRunAppsConsole(t *testing.T) {
	appID := uuid.New().String()
	deploymentID := uuid.New().String()
	component := "service"

	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		tm.apps.EXPECT().GetExec(appID, deploymentID, component).Times(1).Return(&godo.AppExec{URL: "wss://proxy-apps-prod-ams3-001.ondigitalocean.app/?token=aa-bb-11-cc-33"}, nil)
		tm.listen.EXPECT().Listen(gomock.Any()).Times(1).Return(nil)
		tm.terminal.EXPECT().ReadRawStdin(gomock.Any(), gomock.Any()).Times(1).Return(func() {}, nil)
		tm.terminal.EXPECT().MonitorResizeEvents(gomock.Any(), gomock.Any()).Times(1).Return(nil)

		tc := config.Doit.(*doctl.TestConfig)
		tc.ListenFn = func(url *url.URL, token string, schemaFunc listen.SchemaFunc, out io.Writer, in <-chan []byte) listen.ListenerService {
			assert.Equal(t, "aa-bb-11-cc-33", token)
			assert.Equal(t, "wss://proxy-apps-prod-ams3-001.ondigitalocean.app/?token=aa-bb-11-cc-33", url.String())
			return tm.listen
		}
		tc.TerminalFn = func() terminal.Terminal {
			return tm.terminal
		}

		config.Args = append(config.Args, appID, component)
		config.Doit.Set(config.NS, doctl.ArgAppDeployment, deploymentID)

		err := RunAppsConsole(config)
		require.NoError(t, err)
	})
}

const (
	validJSONSpec = `{
	"name": "test",
	"services": [
		{
			"name": "web",
			"github": {
				"repo": "digitalocean/sample-golang",
				"branch": "main"
			}
		}
	],
	"static_sites": [
		{
			"name": "static",
			"git": {
				"repo_clone_url": "git@github.com:digitalocean/sample-gatsby.git",
				"branch": "main"
			},
			"routes": [
				{
				"path": "/static"
				}
			]
		}
	]
}`
	validYAMLSpec = `name: test
services:
- github:
    branch: main
    repo: digitalocean/sample-golang
  name: web
static_sites:
- git:
    branch: main
    repo_clone_url: git@github.com:digitalocean/sample-gatsby.git
  name: static
  routes:
  - path: /static
`
	unknownFieldSpec = `
name: test
bugField: bad
services:
- name: web
  github:
    repo: digitalocean/sample-golang
    branch: main
static_sites:
- name: static
  git:
    repo_clone_url: git@github.com:digitalocean/sample-gatsby.git
    branch: main
  routes:
  - path: /static
`
)

var validAppSpec = &godo.AppSpec{
	Name: "test",
	Services: []*godo.AppServiceSpec{
		{
			Name: "web",
			GitHub: &godo.GitHubSourceSpec{
				Repo:   "digitalocean/sample-golang",
				Branch: "main",
			},
		},
	},
	StaticSites: []*godo.AppStaticSiteSpec{
		{
			Name: "static",
			Git: &godo.GitSourceSpec{
				RepoCloneURL: "git@github.com:digitalocean/sample-gatsby.git",
				Branch:       "main",
			},
			Routes: []*godo.AppRouteSpec{
				{Path: "/static"},
			},
		},
	},
}

func testTempFile(t *testing.T, data []byte) string {
	t.Helper()
	file := t.TempDir() + "/file"
	err := os.WriteFile(file, data, 0644)
	require.NoError(t, err, "writing temp file")
	return file
}

func TestRunAppSpecValidate(t *testing.T) {
	tcs := []struct {
		name       string
		spec       string
		schemaOnly bool
		mock       func(tm *tcMocks)

		wantError string
		wantOut   string
	}{
		{
			name:       "valid yaml",
			spec:       validYAMLSpec,
			schemaOnly: true,
			wantOut:    validYAMLSpec,
		},
		{
			name:       "valid json",
			spec:       validJSONSpec,
			schemaOnly: true,
			wantOut:    validYAMLSpec,
		},
		{
			name: "valid json with ProposeApp req",
			spec: validJSONSpec,
			mock: func(tm *tcMocks) {
				tm.apps.EXPECT().Propose(&godo.AppProposeRequest{
					Spec: validAppSpec,
				}).Return(&godo.AppProposeResponse{
					Spec: &godo.AppSpec{
						Name: "validated-spec",
					},
				}, nil)
			},
			wantOut: "name: validated-spec\n",
		},
		{
			name:       "invalid",
			spec:       "hello",
			schemaOnly: true,
			wantError:  "parsing app spec: json: cannot unmarshal string into Go value of type godo.AppSpec",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
				config.Args = append(config.Args, testTempFile(t, []byte(tc.spec)))
				config.Doit.Set(config.NS, doctl.ArgSchemaOnly, tc.schemaOnly)
				var buf bytes.Buffer
				config.Out = &buf

				if tc.mock != nil {
					tc.mock(tm)
				}

				err := RunAppsSpecValidate(config)
				if tc.wantError != "" {
					require.Equal(t, tc.wantError, err.Error())
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tc.wantOut, buf.String())
			})
		})
	}
}

func TestRunAppSpecGet(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		app := &godo.App{
			ID:   uuid.New().String(),
			Spec: &testAppSpec,
		}

		tm.apps.EXPECT().Get(app.ID).Times(2).Return(app, nil)

		t.Run("yaml", func(t *testing.T) {
			var buf bytes.Buffer
			config.Doit.Set(config.NS, doctl.ArgFormat, "yaml")
			config.Args = append(config.Args, app.ID)
			config.Out = &buf

			err := RunAppsSpecGet(config)
			require.NoError(t, err)
			require.Equal(t, `name: test
services:
- github:
    branch: main
    repo: digitalocean/doctl
  name: service
`, buf.String())
		})

		t.Run("json", func(t *testing.T) {
			var buf bytes.Buffer
			config.Doit.Set(config.NS, doctl.ArgFormat, "json")
			config.Args = append(config.Args, app.ID)
			config.Out = &buf

			err := RunAppsSpecGet(config)
			require.NoError(t, err)
			require.Equal(t, `{
  "name": "test",
  "services": [
    {
      "name": "service",
      "github": {
        "repo": "digitalocean/doctl",
        "branch": "main"
      }
    }
  ]
}
`, buf.String())
		})
	})

	t.Run("with-deployment", func(t *testing.T) {
		withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
			appID := uuid.New().String()
			deployment := &godo.Deployment{
				ID:   uuid.New().String(),
				Spec: &testAppSpec,
			}

			tm.apps.EXPECT().GetDeployment(appID, deployment.ID).Times(1).Return(deployment, nil)

			var buf bytes.Buffer
			config.Doit.Set(config.NS, doctl.ArgFormat, "yaml")
			config.Doit.Set(config.NS, doctl.ArgAppDeployment, deployment.ID)
			config.Args = append(config.Args, appID)
			config.Out = &buf

			err := RunAppsSpecGet(config)
			require.NoError(t, err)
			require.Equal(t, `name: test
services:
- github:
    branch: main
    repo: digitalocean/doctl
  name: service
`, buf.String())
		})
	})
}

func TestRunAppsListRegions(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		regions := []*godo.AppRegion{{
			Slug:        "ams",
			Label:       "Amsterdam",
			Flag:        "netherlands",
			Continent:   "Europe",
			DataCenters: []string{"ams3"},
			Default:     true,
		}}

		tm.apps.EXPECT().ListRegions().Times(1).Return(regions, nil)

		err := RunAppsListRegions(config)
		require.NoError(t, err)
	})
}

func TestRunAppsTierList(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		tiers := []*godo.AppTier{testAppTier}

		tm.apps.EXPECT().ListTiers().Times(1).Return(tiers, nil)

		err := RunAppsTierList(config)
		require.NoError(t, err)
	})
}

func TestRunAppsTierGet(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		tm.apps.EXPECT().GetTier(testAppTier.Slug).Times(1).Return(testAppTier, nil)

		config.Args = append(config.Args, testAppTier.Slug)
		err := RunAppsTierGet(config)
		require.NoError(t, err)
	})
}

func TestRunAppsTierInstanceSizeList(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		instanceSizes := []*godo.AppInstanceSize{testAppInstanceSize}

		tm.apps.EXPECT().ListInstanceSizes().Times(1).Return(instanceSizes, nil)

		err := RunAppsTierInstanceSizeList(config)
		require.NoError(t, err)
	})
}

func TestRunAppsTierInstanceSizeGet(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		tm.apps.EXPECT().GetInstanceSize(testAppInstanceSize.Slug).Times(1).Return(testAppInstanceSize, nil)

		config.Args = append(config.Args, testAppInstanceSize.Slug)
		err := RunAppsTierInstanceSizeGet(config)
		require.NoError(t, err)
	})
}

func TestRunAppsListAlerts(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		appID := uuid.New().String()
		tm.apps.EXPECT().ListAlerts(appID).Times(1).Return(testAlerts, nil)

		config.Args = append(config.Args, appID)
		err := RunAppListAlerts(config)
		require.NoError(t, err)
	})
}

func TestRunAppsUpdateAlertDestinations(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		destinationsFile, err := os.CreateTemp(t.TempDir(), "dest")
		require.NoError(t, err)
		defer destinationsFile.Close()

		err = json.NewEncoder(destinationsFile).Encode(&testAlertUpdate)
		require.NoError(t, err)
		appID := uuid.New().String()
		tm.apps.EXPECT().UpdateAlertDestinations(appID, testAlert.ID, &testAlertUpdate).Times(1).Return(&testAlert, nil)

		config.Args = append(config.Args, appID, testAlert.ID)
		config.Doit.Set(config.NS, doctl.ArgAppAlertDestinations, destinationsFile.Name())
		err = RunAppUpdateAlertDestinations(config)
		require.NoError(t, err)
	})
}

func TestRunAppsListBuildpacks(t *testing.T) {
	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		tm.apps.EXPECT().ListBuildpacks().Times(1).Return([]*godo.Buildpack{testBuildpack}, nil)

		err := RunAppListBuildpacks(config)
		require.NoError(t, err)
	})
}

func TestRunAppsUpgradeBuildpack(t *testing.T) {
	deployment := &godo.Deployment{
		ID:        uuid.New().String(),
		Spec:      &testAppSpec,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	withTestClient(t, func(config *CmdConfig, tm *tcMocks) {
		appID := uuid.New().String()
		tm.apps.EXPECT().UpgradeBuildpack(appID, godo.UpgradeBuildpackOptions{
			BuildpackID:       "buildpack-id",
			MajorVersion:      3,
			TriggerDeployment: true,
		}).Times(1).Return([]string{"www", "api"}, deployment, nil)

		config.Args = append(config.Args, appID)
		config.Doit.Set(config.NS, doctl.ArgBuildpack, "buildpack-id")
		config.Doit.Set(config.NS, doctl.ArgMajorVersion, 3)
		config.Doit.Set(config.NS, doctl.ArgTriggerDeployment, true)
		err := RunAppUpgradeBuildpack(config)
		require.NoError(t, err)
	})
}
