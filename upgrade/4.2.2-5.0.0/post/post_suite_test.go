package post_test

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
	"github.com/concourse/concourse/go-concourse/concourse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var (
	fly       Fly
	parsedEnv environment
	ccClient  concourse.Client
	Teams     string
	Pipelines []string
)
var totalTests = 0
var testsSucceeded = 0

const TARGET = "drill"

type environment struct {
	FlyPath  string `env:"FLY_5_0_0_PATH"`
	FlyHome  string `env:"FLY_HOME"`
	Username string `env:"ATC_USERNAME"`
	Password string `env:"ATC_PASSWORD"`
	Endpoint string `env:"ATC_URL"`
}

type Container struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	WorkerName string `json:"worker_name"`
}

type Worker struct {
	Name            string   `json:"name"`
	State           string   `json:"state"`
	GardenAddress   string   `json:"addr"`
	BaggageclaimUrl string   `json:"baggageclaim_url"`
	Team            string   `json:"team"`
	Tags            []string `json:"tags"`
}

type Team struct {
	ID   int                    `json:"id"`
	Name string                 `json:"name"`
	Auth map[string]interface{} `json:"auth"`
}

type ResourceVersion struct {
	ID      int               `json:"id"`
	Version map[string]string `json:"version"`
	Enabled bool              `json:"enabled"`
}

type Pipeline struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	TeamName string `json:"team_name"`
	Public   bool   `json:"public"`
	Paused   bool   `json:"paused"`
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
})

func TestPost(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "4.2.2 - 5.0.0 Post Suite")
}

type Fly struct {
	Bin    string
	Target string
	Home   string
}

func teamsAndPipelinesExist(teamPipelines map[string][]string) {
	teams := fly.GetTeams()

	existingTeams := []string{}
	for _, team := range teams {
		existingTeams = append(existingTeams, team.Name)
	}

	for team, expectedPipelines := range teamPipelines {
		Expect(existingTeams).To(ContainElement(team))

		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", team)

		pipelines := fly.GetPipelines()

		existingPipelines := []string{}
		for _, pipeline := range pipelines {
			existingPipelines = append(existingPipelines, pipeline.Name)
		}

		for _, p := range expectedPipelines {
			Expect(existingPipelines).To(ContainElement(p))
		}
	}
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

func (f *Fly) GetTeams() []Team {
	var teams = []Team{}

	sess := f.Start("teams", "--json")
	<-sess.Exited
	Expect(sess.ExitCode()).To(BeZero())

	err := json.Unmarshal(sess.Out.Contents(), &teams)
	Expect(err).ToNot(HaveOccurred())

	return teams
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

func (f *Fly) GetVersions(pipeline string, resource string) []ResourceVersion {
	var versions = []ResourceVersion{}

	sess := f.Start("resource-versions", "r", pipeline+"/"+resource, "--json")
	<-sess.Exited
	Expect(sess.ExitCode()).To(BeZero())

	err := json.Unmarshal(sess.Out.Contents(), &versions)
	Expect(err).ToNot(HaveOccurred())

	return versions
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

func (f *Fly) flyTable(argv ...string) []map[string]string {
	session := f.Start(append([]string{"--print-table-headers"}, argv...)...)
	<-session.Exited
	Expect(session.ExitCode()).To(Equal(0))

	result := []map[string]string{}

	var headers []string
	for i, cols := range parseTable(string(session.Out.Contents())) {
		if i == 0 {
			headers = cols
			continue
		}

		result = append(result, map[string]string{})

		for j, header := range headers {
			if header == "" || cols[j] == "" {
				continue
			}

			result[i-1][header] = cols[j]
		}
	}

	return result
}

func parseTable(content string) [][]string {
	result := [][]string{}

	var expectedColumns int
	rows := strings.Split(content, "\n")
	for i, row := range rows {
		if row == "" {
			continue
		}

		columns := splitTableColumns(row)
		if i == 0 {
			expectedColumns = len(columns)
		} else {
			Expect(columns).To(HaveLen(expectedColumns))
		}

		result = append(result, columns)
	}

	return result
}

func splitTableColumns(row string) []string {
	return regexp.MustCompile(`(\s{2,}|\t)`).Split(strings.TrimSpace(row), -1)
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

func waitForBuildAndWatch(pipelineName, jobName string, buildName ...string) *gexec.Session {
	args := []string{"watch", "-j", pipelineName + "/" + jobName}

	if len(buildName) > 0 {
		args = append(args, "-b", buildName[0])
	}

	keepPollingCheck := regexp.MustCompile("job has no builds|build not found|failed to get build")
	for {
		session := spawnFly(args...)
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

func spawnFly(argv ...string) *gexec.Session {
	return spawn(parsedEnv.FlyPath, append([]string{"-t", TARGET}, argv...)...)
}

func spawn(argc string, argv ...string) *gexec.Session {
	By("running: " + argc + " " + strings.Join(argv, " "))
	cmd := exec.Command(argc, argv...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	return session
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
