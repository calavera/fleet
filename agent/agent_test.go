package agent

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/coreos/fleet/job"
	"github.com/coreos/fleet/machine"
	"github.com/coreos/fleet/unit"
)

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

var parseMultivalueTest = map[string][]string{
	`foo=bar`:             []string{`foo=bar`},
	`"foo=bar"`:           []string{`foo=bar`},
	`"foo=bar" "baz=qux"`: []string{`foo=bar`, `baz=qux`},
	`"foo=bar baz"`:       []string{`foo=bar baz`},
	` "foo=bar" baz`:      []string{`foo=bar`, `baz`},
	`baz "foo=bar"`:       []string{`baz`, `foo=bar`},
}

func TestParseMultivalueLine(t *testing.T) {
	for q, w := range parseMultivalueTest {
		g := parseMultivalueLine(q)
		if !reflect.DeepEqual(g, w) {
			t.Errorf("Unexpected line parse for %q:\ngot %q\nwant %q", q, g, w)
		}
	}
}

var metadataAbleToRunTest = []struct {
	C string
	A bool
}{
	// valid metadata
	{`X-ConditionMachineMetadata=region=us-west-1`, true},
	{`X-ConditionMachineMetadata= "region=us-east-1" "region=us-west-1"`, true},
	{`X-ConditionMachineMetadata=region=us-east-1
X-ConditionMachineMetadata=region=us-west-1"`, true},
	{`X-ConditionMachineMetadata=region=us-east-1`, false},

	// ignored/invalid metadata
	{`X-ConditionMachineMetadata=us-west-1`, true},
	{`X-ConditionMachineMetadata==us-west-1`, true},
	{`X-ConditionMachineMetadata=region=`, true},
}

func TestAbleToRunWithConditionMachineMetadata(t *testing.T) {
	agent := &Agent{machine: newTestMachine("us-west-1"), state: NewState()}

	for i, e := range metadataAbleToRunTest {
		job := newTestJobWithMachineMetadata(e.C)
		g := agent.AbleToRun(job)
		if g != e.A {
			t.Errorf("Unexpected output %d, content: %q\n\tgot %q, want %q\n", i, e.C, g, e.A)
		}
	}
}
