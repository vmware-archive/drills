package pre_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

const PassedConstraintsBetweenJobsIdentifier = "passed-constraints-between-jobs"

var _ = Describe("Passed constraints within a pipeline run", func() {
	BeforeEach(func() {
		By("setting up a team")
		fly.Run("set-team", "-n", "team-"+PassedConstraintsBetweenJobsIdentifier, "--non-interactive", "--local-user="+parsedEnv.Username)

		By("setting the pipeline")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team-"+PassedConstraintsBetweenJobsIdentifier)
		fly.Run("sp", "-p", "pipeline-"+PassedConstraintsBetweenJobsIdentifier,
			"-c", "../pipelines/passed-jobs.yml",
			"-n",
		)
		fly.Run("up", "-p", "pipeline-"+PassedConstraintsBetweenJobsIdentifier)
	})

	It("starts job 1 and job 2", func() {
		By("manually triggering job 1")
		fly.Run("trigger-job", "-j", "pipeline-"+PassedConstraintsBetweenJobsIdentifier+"/job1")
		fly.Run("watch", "-j", "pipeline-"+PassedConstraintsBetweenJobsIdentifier+"/job1")

		By("watching job 2 get triggered through the pass constraints")
		watch := fly.WaitForBuildAndWatch("pipeline-"+PassedConstraintsBetweenJobsIdentifier, "job2")
		Eventually(watch).Should(gbytes.Say("hello-world"))
	})
})
