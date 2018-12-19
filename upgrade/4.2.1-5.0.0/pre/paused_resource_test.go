package pre_test

import (
	"fmt"

	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Paused resources", func() {
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
		fly.Run("sp", "-p", "resource1-"+guid.String(),
			"-c", "pipelines/release-resource.yml",
			"-y", "trigger=false",
			"-v", "hash="+guid.String(),
			"-n",
		)
		fly.Run("up", "-p", "resource1-"+guid.String())
		fly.Run("cr", "-r", fmt.Sprintf("resource1-%s/some-resource", guid.String()))

		By("pausing the resource")
		fly.Run("pr", "-r", fmt.Sprintf("resource1-%s/some-resource", guid.String()))
	})

	It("has a paused resource", func() {
		ccClient := login(parsedEnv.Endpoint, parsedEnv.Username, parsedEnv.Password)
		pausedResource, found, err := ccClient.Team("team1-"+guid.String()).Resource("resource1-"+guid.String(), "some-resource")
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(pausedResource.Paused).To(BeTrue())
	})
})
