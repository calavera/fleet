package agent

import (
	"fmt"
	"testing"

	"github.com/coreos/fleet/job"
	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/unit"
)

func TestAbleToRunConditionMachineBootIDMatch(t *testing.T) {
	uf := unit.NewSystemdUnitFile(`[X-Fleet]
X-ConditionMachineBootID=XYZ
`)
	payload := job.NewJobPayload("example.service", *uf)
	job := job.NewJob("example.service", *payload)

	mach := machine.New("XYZ", "", make(map[string]string, 0))
	agent := Agent{machine: mach, state: NewState()}
	if !agent.AbleToRun(job) {
		t.Fatalf("Agent should be able to run job")
	}
}

func TestAbleToRunConditionMachineBootIDMismatch(t *testing.T) {
	uf := unit.NewSystemdUnitFile(`[X-Fleet]
X-ConditionMachineBootID=XYZ
`)
	payload := job.NewJobPayload("example.service", *uf)
	job := job.NewJob("example.service", *payload)

	mach := machine.New("123", "", make(map[string]string, 0))
	agent := Agent{machine: mach, state: NewState()}
	if agent.AbleToRun(job) {
		t.Fatalf("Agent should not be able to run job")
	}
}

func TestAbleToRunWithConditionMachineMetadata(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-west-1"), state: NewState()}

	job := newTestJobWithMachineMetadata("X-ConditionMachineMetadataRegion=us-west-1")
	if !agent.AbleToRun(job) {
		t.Fatal("Expected to be able to run the job with same region metadata")
	}
}

func TestAbleToRunMatchingOneOfTheConditions(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-west-1"), state: NewState()}

	job := newTestJobWithMachineMetadata(`X-ConditionMachineMetadataRegion=us-east-1
X-ConditionMachineMetadataRegion=us-west-1`)

	if !agent.AbleToRun(job) {
		t.Fatal("Expected to be able to run the job with one matching region in the metadata")
	}
}

func TestNotAbleToRunWithoutConditionMachineMetadata(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-east-1"), state: NewState()}

	job := newTestJobWithMachineMetadata("X-ConditionMachineMetadataRegion=us-west-1")
	if agent.AbleToRun(job) {
		t.Fatal("Expected to not be able to run the job with different region metadata")
	}
}

func TestAbleToRunWithDeprecatedMachineMetadata(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-west-1"), state: NewState()}

	job := newTestJobWithMachineMetadata("X-MachineMetadataRegion=us-west-1")
	if !agent.AbleToRun(job) {
		t.Fatal("Expected to be able to run the job with same region metadata")
	}
}

func newTestMachine(region string) *machine.Machine {
	metadata := map[string]string{
		"Region": region,
	}
	return machine.New("", "", metadata)
}

func newTestJobWithMachineMetadata(metadata string) *job.Job {
	contents := fmt.Sprintf(`
[X-Fleet]
%s
`, metadata)

	ms := &machine.MachineState{"XXX", "", make(map[string]string, 0), "1"}
	js1 := job.NewJobState("loaded", "inactive", "running", []string{}, ms)
	jp1 := job.NewJobPayload("us-west.service", *unit.NewSystemdUnitFile(contents))

	return job.NewJob("pong.service", jp1, js1)
}
