package pre_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

const PausedToPinnedIdentifier = "paused-to-pinned-resource"

var _ = Describe("Paused resources", func() {
	BeforeEach(func() {
		By("setting up a team")
		fly.Run("set-team", "-n", "team-"+PausedToPinnedIdentifier, "--non-interactive", "--local-user="+parsedEnv.Username)

		By("setting the pipeline")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team-"+PausedToPinnedIdentifier)
		fly.Run("sp", "-p", "pipeline-"+PausedToPinnedIdentifier,
			"-c", "../pipelines/mock-resource.yml",
			"-y", "trigger=true",
			"-n",
		)
		fly.Run("up", "-p", "pipeline-"+PausedToPinnedIdentifier)
		fly.Run("cr", "-r", fmt.Sprintf("pipeline-%s/some-resource", PausedToPinnedIdentifier))

		watch := fly.WaitForBuildAndWatch("pipeline-"+PausedToPinnedIdentifier, "some-job")
		Eventually(watch).Should(gbytes.Say("succeeded"))

		By("pausing the resource")
		fly.Run("pr", "-r", fmt.Sprintf("pipeline-%s/some-resource", PausedToPinnedIdentifier))
	})

	It("has a paused resource", func() {
		ccClient := login(parsedEnv.Endpoint, parsedEnv.Username, parsedEnv.Password)
		pausedResource, found, err := ccClient.Team("team-"+PausedToPinnedIdentifier).Resource("pipeline-"+PausedToPinnedIdentifier, "some-resource")
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(pausedResource.Paused).To(BeTrue())
	})
})
