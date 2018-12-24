package pre_test

import (
	"fmt"

	common "github.com/concourse/drills/upgrade/4.2.1-5.0.0/common"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Identical resource configs across teams", func() {
	var (
		mainTeamContainers  []Container
		otherTeamContainers []Container
		guid                *uuid.UUID
	)

	BeforeEach(func() {
		var err error
		guid, err = uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		By("setting two teams")
		fly.Run("set-team", "-n", "team1-"+guid.String(), "--non-interactive", "--local-user="+parsedEnv.Username)
		fly.Run("set-team", "-n", "team2-"+guid.String(), "--non-interactive", "--local-user="+parsedEnv.Username)

		By("setting team 1s pipeline and creating a check container for the resource that uses the shared resource config")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+guid.String())
		fly.Run("sp", "-p", "resource1-"+guid.String(),
			"-c", "pipelines/resource.yml",
			"-v", "hash="+guid.String(),
			"-y", "trigger=false",
			"-n",
		)
		fly.Run("up", "-p", "resource1-"+guid.String())
		fly.Run("cr", "-r", fmt.Sprintf("resource1-%s/time-resource", guid.String()))

		By("setting up team 2s pipeline and creating a check container for the resource that uses the shared resource config")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team2-"+guid.String())
		fly.Run("sp", "-p", "resource2-"+guid.String(),
			"-c", "pipelines/resource.yml",
			"-v", "hash="+guid.String(),
			"-y", "trigger=false",
			"-n",
		)
		fly.Run("up", "-p", "resource2-"+guid.String())
		fly.Run("cr", "-r", fmt.Sprintf("resource2-%s/time-resource", guid.String()))
	})

	It("should have a check container per team", func() {
		By("ensuring the team 1's container is a check container")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+guid.String())
		mainTeamContainers = fly.GetContainers()
		Expect(mainTeamContainers).To(HaveLen(1))
		Expect(mainTeamContainers[0].Type).To(Equal("check"))

		By("ensuring the other teams container is a check container")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team2-"+guid.String())
		otherTeamContainers = fly.GetContainers()
		Expect(otherTeamContainers).To(HaveLen(1))
		Expect(otherTeamContainers[0].Type).To(Equal("check"))

		By("ensuring the two teams check containers are not the same")
		Expect(mainTeamContainers[0].ID).ToNot(Equal(otherTeamContainers[0].ID))
	})
})
