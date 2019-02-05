package post_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const PassedConstraintsBetweenJobsIdentifier = "passed-constraints-between-jobs"

var _ = Describe("Passed constraints within a pipeline run", func() {
	Context("when two jobs with passed constraints have already been run before the upgrade", func() {
		It("should not need to run again when the third job is triggered", func() {
			fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team-"+PassedConstraintsBetweenJobsIdentifier)

			By("manually triggering job 3")
			fly.Run("trigger-job", "-j", "pipeline-"+PassedConstraintsBetweenJobsIdentifier+"/job3")
			fly.Run("watch", "-j", "pipeline-"+PassedConstraintsBetweenJobsIdentifier+"/job3")

			jobBuilds1 := fly.flyTable("builds", "-j", "pipeline-"+PassedConstraintsBetweenJobsIdentifier+"/job1")
			Expect(jobBuilds1).To(HaveLen(1))

			jobBuilds2 := fly.flyTable("builds", "-j", "pipeline-"+PassedConstraintsBetweenJobsIdentifier+"/job2")
			Expect(jobBuilds2).To(HaveLen(1))
		})
	})
})
