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

	job := newTestJobWithMachineMetadata(`X-ConditionMachineMetadata=region=us-west-1`)
	if !agent.AbleToRun(job) {
		t.Fatal("Expected to be able to run the job with same region metadata")
	}
}

func TestAbleToRunMatchingOneOfTheConditions(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-west-1"), state: NewState()}

	job := newTestJobWithMachineMetadata(`X-ConditionMachineMetadata= "region=us-east-1" "region=us-west-1"`)
	if !agent.AbleToRun(job) {
		t.Fatal("Expected to be able to run the job with one matching region in the metadata")
	}
}

func TestNotAbleToRunWithoutConditionMachineMetadata(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-east-1"), state: NewState()}

	job := newTestJobWithMachineMetadata(`X-ConditionMachineMetadata= "region=us-west-1"`)
	if agent.AbleToRun(job) {
		t.Fatal("Expected to not be able to run the job with different region metadata")
	}
}

func TestAbleToRunWithDeprecatedMachineMetadata(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-west-1"), state: NewState()}

	job := newTestJobWithMachineMetadata("X-MachineMetadataregion=us-west-1")
	if !agent.AbleToRun(job) {
		t.Fatal("Expected to be able to run the job with same region metadata")
	}
}

func TestAbleToRunWithBadFormattedMachineMetadata(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-west-1"), state: NewState()}

	job := newTestJobWithMachineMetadata(`X-ConditionMachineMetadata==us-west-1`)
	if !agent.AbleToRun(job) {
		t.Fatal("Expected to ignore bad formatted metadata")
	}

	job = newTestJobWithMachineMetadata(`X-ConditionMachineMetadata=us-west-1=`)
	if !agent.AbleToRun(job) {
		t.Fatal("Expected to ignore bad formatted metadata")
	}
}

func newTestMachine(region string) *machine.Machine {
	metadata := map[string]string{
		"region": region,
	}
	return machine.New("", "", metadata)
}

func newTestJobWithMachineMetadata(metadata string) *job.Job {
	contents := fmt.Sprintf(`
[X-Fleet]
%s
`, metadata)

	jp1 := job.NewJobPayload("us-west.service", *unit.NewSystemdUnitFile(contents))

	return job.NewJob("pong.service", *jp1)
}
