package post_test

import (
	"fmt"

	"github.com/concourse/concourse/go-concourse/concourse"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const UniqueVersionHistoryIdentifier = "unique-version-history"

var _ = Describe("Unique Version History", func() {
	Context("when there are two resources with the same resource config", func() {
		BeforeEach(func() {
			By("setting up a team")
			fly.Run("set-team", "-n", "team-"+UniqueVersionHistoryIdentifier, "--non-interactive", "--local-user="+parsedEnv.Username)
			fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team-"+UniqueVersionHistoryIdentifier)

			By("setting the first pipeline")
			fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team-"+UniqueVersionHistoryIdentifier)
			fly.Run("sp", "-p", "pipeline1-"+UniqueVersionHistoryIdentifier,
				"-c", "../pipelines/unique-versions.yml",
				"-n",
			)
			fly.Run("up", "-p", "pipeline1-"+UniqueVersionHistoryIdentifier)
			fly.Run("cr", "-r", fmt.Sprintf("pipeline1-%s/some-resource", UniqueVersionHistoryIdentifier), "-f", "version:1")

			By("setting the second pipeline")
			fly.Run("sp", "-p", "pipeline2-"+UniqueVersionHistoryIdentifier,
				"-c", "../pipelines/unique-versions.yml",
				"-n",
			)
			fly.Run("up", "-p", "pipeline2-"+UniqueVersionHistoryIdentifier)
			fly.Run("cr", "-r", fmt.Sprintf("pipeline2-%s/some-resource", UniqueVersionHistoryIdentifier), "-f", "version:2")
		})

		It("should have different version history", func() {
			ccClient := login(parsedEnv.Endpoint, parsedEnv.Username, parsedEnv.Password)
			pipeline1Versions, _, found, err := ccClient.Team("team-"+UniqueVersionHistoryIdentifier).ResourceVersions("pipeline1-"+UniqueVersionHistoryIdentifier, "some-resource", concourse.Page{})
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(pipeline1Versions).To(HaveLen(1))

			pipeline2Versions, _, found, err := ccClient.Team("team-"+UniqueVersionHistoryIdentifier).ResourceVersions("pipeline2-"+UniqueVersionHistoryIdentifier, "some-resource", concourse.Page{})
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(pipeline2Versions).To(HaveLen(1))

			Expect(pipeline1Versions[0].Version).ToNot(Equal(pipeline2Versions[0].Version))
		})
	})
})
