package pre_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/caarlos0/env"
	"github.com/concourse/go-concourse/concourse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var (
	fly       Fly
	parsedEnv environment
	ccClient  concourse.Client
)

var testsCompletedSuccessfully = 0
var totalTests = 0

const TARGET = "drill"

type environment struct {
	FlyPath  string `env:"FLY_4_2_2_PATH"`
	FlyHome  string `env:"FLY_HOME"`
	Username string `env:"ATC_USERNAME"`
	Password string `env:"ATC_PASSWORD"`
	Endpoint string `env:"ATC_URI"`
}

type Container struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	WorkerName string `json:"worker_name"`
}

type ResourceVersion struct {
	ID      int               `json:"id"`
	Version map[string]string `json:"version"`
	Enabled bool              `json:"enabled"`
}

type Team struct {
	ID   int                 `json:"id"`
	Name string              `json:"name"`
	Auth map[string][]string `json:"auth"`
}

type Worker struct {
	Name            string   `json:"name"`
	State           string   `json:"state"`
	GardenAddress   string   `json:"addr"`
	BaggageclaimUrl string   `json:"baggageclaim_url"`
	Team            string   `json:"team"`
	Tags            []string `json:"tags"`
}

type Pipeline struct {
	Name string `json:"name"`
}

var _ = BeforeSuite(func() {
	err := env.Parse(&parsedEnv)
	Expect(err).ToNot(HaveOccurred())

	fly = Fly{
		Target: TARGET,
		Bin:    parsedEnv.FlyPath,
		Home:   parsedEnv.FlyHome,
	}

	fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint)
	teams := fly.GetTeams()
	if len(teams) <= 1 {
		return
	}

	for _, team := range teams {
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", team.Name)

		pipelines := fly.GetPipelines()
		for _, pipeline := range pipelines {
			fly.Run("dp", "-p", pipeline.Name, "--non-interactive")
		}
	}

	for _, team := range teams {
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", team.Name)

		Eventually(func() interface{} {
			containers := fly.GetContainers()
			return len(containers)
		}, 5*time.Minute, 5*time.Second).Should(BeZero())
	}

	waitForWorkerToBeRunning("", "")
})

func (f *Fly) GetTeams() []Team {
	var teams = []Team{}

	sess := f.Start("teams", "--json")
	<-sess.Exited
	Expect(sess.ExitCode()).To(BeZero())

	err := json.Unmarshal(sess.Out.Contents(), &teams)
	Expect(err).ToNot(HaveOccurred())

	return teams
}

func (f *Fly) GetWorkers() []Worker {
	var workers = []Worker{}

	sess := f.Start("workers", "--json")
	<-sess.Exited
	Expect(sess.ExitCode()).To(BeZero())

	err := json.Unmarshal(sess.Out.Contents(), &workers)
	Expect(err).ToNot(HaveOccurred())

	return workers
}

func (f *Fly) GetPipelines() []Pipeline {
	var pipelines = []Pipeline{}

	sess := f.Start("pipelines", "--json")
	<-sess.Exited
	Expect(sess.ExitCode()).To(BeZero())

	err := json.Unmarshal(sess.Out.Contents(), &pipelines)
	Expect(err).ToNot(HaveOccurred())

	return pipelines
}
func TestPre(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "4.2.1 - 5.0.0 Pre Suite")
}

type Fly struct {
	Bin    string
	Target string
	Home   string
}

func (f *Fly) Login(user, password, endpoint string, argv ...string) {
	Eventually(func() *gexec.Session {

		args := append([]string{
			"login",
			"-c", endpoint,
			"-u", user,
			"-p", password,
		}, argv...)

		sess := f.Start(
			args...,
		)

		<-sess.Exited
		return sess
	}, 2*time.Minute, 10*time.Second).
		Should(gexec.Exit(0), "Fly should have been able to log in")
}

func (f *Fly) Run(argv ...string) {
	Wait(f.Start(argv...))
}

func (f *Fly) Start(argv ...string) *gexec.Session {
	return Start([]string{"HOME=" + f.Home}, f.Bin, append([]string{"-t", f.Target}, argv...)...)
}

func (f *Fly) GetContainers() []Container {
	var containers = []Container{}

	sess := f.Start("containers", "--json")
	<-sess.Exited
	Expect(sess.ExitCode()).To(BeZero())

	err := json.Unmarshal(sess.Out.Contents(), &containers)
	Expect(err).ToNot(HaveOccurred())

	return containers
}

func (f *Fly) GetVersions(pipeline string, resource string) []ResourceVersion {
	var versions = []ResourceVersion{}

	sess := f.Start("resource-versions", "r", pipeline+"/"+resource, "--json")
	<-sess.Exited
	Expect(sess.ExitCode()).To(BeZero())

	err := json.Unmarshal(sess.Out.Contents(), &versions)
	Expect(err).ToNot(HaveOccurred())

	return versions
}

func (f *Fly) WaitForBuildAndWatch(pipelineName, jobName string, buildName ...string) *gexec.Session {
	args := []string{"watch", "-j", pipelineName + "/" + jobName}

	if len(buildName) > 0 {
		args = append(args, "-b", buildName[0])
	}

	keepPollingCheck := regexp.MustCompile("job has no builds|build not found|failed to get build")
	for {
		session := f.Start(args...)
		<-session.Exited

		if session.ExitCode() == 1 {
			output := strings.TrimSpace(string(session.Err.Contents()))
			if keepPollingCheck.MatchString(output) {
				// build hasn't started yet; keep polling
				time.Sleep(time.Second)
				continue
			}
		}

		return session
	}
}

func Wait(session *gexec.Session) {
	<-session.Exited
	Expect(session.ExitCode()).To(Equal(0))
}

func Start(env []string, command string, argv ...string) *gexec.Session {
	TimestampedBy("running: " + command + " " + strings.Join(argv, " "))

	cmd := exec.Command(command, argv...)
	cmd.Env = env

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	return session
}

func TimestampedBy(msg string) {
	By(fmt.Sprintf("[%.9f] %s", float64(time.Now().UnixNano())/1e9, msg))
}

func login(atcURL, username, password string) concourse.Client {
	oauth2Config := oauth2.Config{
		ClientID:     "fly",
		ClientSecret: "Zmx5",
		Endpoint:     oauth2.Endpoint{TokenURL: parsedEnv.Endpoint + "/sky/token"},
		Scopes:       []string{"openid", "federated:id"},
	}

	token, err := oauth2Config.PasswordCredentialsToken(context.Background(), username, password)
	Expect(err).NotTo(HaveOccurred())

	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(token),
			Base: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}

	return concourse.NewClient(atcURL, httpClient, false)
}

func waitForWorkerToBeRunning(teamName string, tag string) {
	Eventually(func() interface{} {
		workers := fly.GetWorkers()

		for _, worker := range workers {
			if len(worker.Tags) > 0 || tag != "" {
				for _, t := range worker.Tags {
					if t == tag {
						if worker.Team == teamName && worker.State == "running" {
							return true
						}
					}
				}
			} else {
				if worker.Team == teamName && worker.State == "running" {
					return true
				}
			}
		}

		return false
	}, time.Minute, time.Second).Should(BeTrue())
}
