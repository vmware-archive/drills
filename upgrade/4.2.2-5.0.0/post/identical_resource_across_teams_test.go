package post_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const IdenticalResourcesIdentifier = "identical-resources-across-teams"

var _ = Describe("Identical resource configs across teams", func() {
	var (
		mainTeamContainers  []Container
		otherTeamContainers []Container
	)

	BeforeEach(func() {
		By("verifying that the teams and pipelines still exist")
		teamsAndPipelinesExist(map[string][]string{
			"team1-" + IdenticalResourcesIdentifier: []string{"pipeline1-" + IdenticalResourcesIdentifier},
			"team2-" + IdenticalResourcesIdentifier: []string{"pipeline2-" + IdenticalResourcesIdentifier},
		})
	})

	It("should have the same check container for both teams", func() {
		By("ensuring the team 1's container is a check container")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+IdenticalResourcesIdentifier)
		fly.Run("cr", "-r", "pipeline1-"+IdenticalResourcesIdentifier+"/time-resource")

		mainTeamContainers = fly.GetContainers()
		Expect(mainTeamContainers).To(HaveLen(1))
		Expect(mainTeamContainers[0].Type).To(Equal("check"))

		By("ensuring the other teams container is a check container")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team2-"+IdenticalResourcesIdentifier)
		fly.Run("cr", "-r", "pipeline2-"+IdenticalResourcesIdentifier+"/time-resource")
		otherTeamContainers = fly.GetContainers()
		Expect(otherTeamContainers).To(HaveLen(1))
		Expect(otherTeamContainers[0].Type).To(Equal("check"))

		By("ensuring the two teams check containers are the same")
		Expect(mainTeamContainers[0].ID).To(Equal(otherTeamContainers[0].ID))
	})
})
