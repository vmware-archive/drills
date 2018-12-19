package pre_test

import (
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Passed constraints within a pipeline run", func() {
	var (
		guid *uuid.UUID
	)

	BeforeEach(func() {
		var err error
		guid, err = uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		By("setting up a team")
		fly.Run("set-team", "-n", "team1-"+guid.String(), "--non-interactive", "--local-user="+parsedEnv.Username)

		By("setting the pipeline")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+guid.String())
		fly.Run("sp", "-p", "passed-constraints-"+guid.String(),
			"-c", "pipelines/passed-jobs.yml",
			"-v", "hash="+guid.String(),
			"-n",
		)
		fly.Run("up", "-p", "passed-constraints-"+guid.String())
	})

	It("starts job 1 and job 2", func() {
		By("manually triggering job 1")
		fly.Run("trigger-job", "-j", "passed-constraints-"+guid.String()+"/job1")
		fly.Run("watch", "-j", "passed-constraints-"+guid.String()+"/job1")

		By("watching job 2 get triggered through the pass constraints")
		watch := waitForBuildAndWatch("passed-constraints-"+guid.String(), "job2")
		Eventually(watch).Should(gbytes.Say("hello-world"))
	})
})
