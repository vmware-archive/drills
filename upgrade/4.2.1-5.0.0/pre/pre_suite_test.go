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

const TARGET = "drill"

type environment struct {
	FlyPath  string `env:"FLY_4_2_1_PATH"`
	FlyHome  string `env:"FLY_HOME"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	Endpoint string `env:"ATC_URI"`
}

type Container struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ResourceVersion struct {
	ID      int               `json:"id"`
	Version map[string]string `json:"version"`
	Enabled bool              `json:"enabled"`
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

	// workers := fly.GetWorkers()
	// var mainTeamWorkerCount int
	// for _, worker := range workers {
	// 	if worker.Team == "main" {
	// 		mainTeamWorkerCount++
	// 	}
	// }

	// Expect(mainTeamWorkerCount).ToNot(BeZero())
})

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
