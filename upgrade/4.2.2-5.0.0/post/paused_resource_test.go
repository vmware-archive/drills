package post_test

import (
	"time"

	"github.com/concourse/concourse/atc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const PausedToPinnedIdentifier = "paused-to-pinned-resource"

var _ = Describe("Paused resources becomes pinned", func() {
	Context("when there was a paused resource before the upgrade", func() {
		It("becomes pinned at the version it was paused at", func() {
			ccClient := login(parsedEnv.Endpoint, parsedEnv.Username, parsedEnv.Password)
			pausedResource, found, err := ccClient.Team("team-"+PausedToPinnedIdentifier).Resource("pipeline-"+PausedToPinnedIdentifier, "some-resource")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(pausedResource.PinnedVersion).To(Equal(atc.Version{"version": "pinned"}))

			fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team-"+PausedToPinnedIdentifier)
			Consistently(func() int {
				builds := fly.flyTable("builds", "-j", "pipeline-"+PausedToPinnedIdentifier+"/some-job")
				return len(builds)
			}, time.Minute, time.Second).Should(Equal(1))
		})
	})
})
