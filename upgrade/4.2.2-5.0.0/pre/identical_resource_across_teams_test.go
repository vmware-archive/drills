package pre_test

import (
	"fmt"

	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const IdenticalResourcesIdentifier = "identical-resources-across-teams"

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
		fly.Run("set-team", "-n", "team1-"+IdenticalResourcesIdentifier, "--non-interactive", "--local-user="+parsedEnv.Username)
		fly.Run("set-team", "-n", "team2-"+IdenticalResourcesIdentifier, "--non-interactive", "--local-user="+parsedEnv.Username)

		By("setting team 1s pipeline and creating a check container for the resource that uses the shared resource config")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+IdenticalResourcesIdentifier)
		fly.Run("sp", "-p", "pipeline1-"+IdenticalResourcesIdentifier,
			"-c", "../pipelines/resource.yml",
			"-v", "hash="+guid.String(),
			"-y", "trigger=false",
			"-n",
		)
		fly.Run("up", "-p", "pipeline1-"+IdenticalResourcesIdentifier)
		fly.Run("cr", "-r", fmt.Sprintf("pipeline1-%s/time-resource", IdenticalResourcesIdentifier))

		By("setting up team 2s pipeline and creating a check container for the resource that uses the shared resource config")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team2-"+IdenticalResourcesIdentifier)
		fly.Run("sp", "-p", "pipeline2-"+IdenticalResourcesIdentifier,
			"-c", "../pipelines/resource.yml",
			"-v", "hash="+guid.String(),
			"-y", "trigger=false",
			"-n",
		)
		fly.Run("up", "-p", "pipeline2-"+IdenticalResourcesIdentifier)
		fly.Run("cr", "-r", fmt.Sprintf("pipeline2-%s/time-resource", IdenticalResourcesIdentifier))
	})

	It("should have a check container per team", func() {
		By("ensuring the team 1's container is a check container")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+IdenticalResourcesIdentifier)
		mainTeamContainers = fly.GetContainers()
		Expect(mainTeamContainers).To(HaveLen(1))
		Expect(mainTeamContainers[0].Type).To(Equal("check"))

		By("ensuring the other teams container is a check container")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team2-"+IdenticalResourcesIdentifier)
		otherTeamContainers = fly.GetContainers()
		Expect(otherTeamContainers).To(HaveLen(1))
		Expect(otherTeamContainers[0].Type).To(Equal("check"))

		By("ensuring the two teams check containers are not the same")
		Expect(mainTeamContainers[0].ID).ToNot(Equal(otherTeamContainers[0].ID))
	})
})
