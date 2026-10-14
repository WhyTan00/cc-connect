package opencode

import (
	"os"
	"slices"
	"testing"

	"github.com/chenhg5/cc-connect/core"
)

func TestWorkspaceAgentOptions_PreservesCommandAcrossRebuild(t *testing.T) {
	// Use the test executable only for LookPath validation; never run a provider.
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		opts map[string]any
		args []string
	}{
		{"binary", map[string]any{"cmd": binary}, nil},
		{"arguments", map[string]any{"cmd": []string{binary, "--unit-test-flag", "two words"}}, []string{"--unit-test-flag", "two words"}},
		{"legacy_path", map[string]any{"cli_path": binary}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original, err := New(tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			snapshot := original.(*Agent).WorkspaceAgentOptions()
			command, args := core.ParseCmdOpts(snapshot, "opencode")
			if command != binary || !slices.Equal(args, tc.args) {
				t.Fatalf("workspace command = %q %v, want %q %v", command, args, binary, tc.args)
			}
			snapshot["work_dir"] = t.TempDir()
			rebuilt, err := New(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			agent := rebuilt.(*Agent)
			if agent.cmd != binary || !slices.Equal(agent.cliExtraArgs, tc.args) {
				t.Fatalf("rebuilt command = %q %v, want %q %v", agent.cmd, agent.cliExtraArgs, binary, tc.args)
			}
		})
	}
}

func TestWorkspaceAgentOptions_CommandSnapshotDoesNotAliasArguments(t *testing.T) {
	a := &Agent{cmd: "unit-test-cli", cliExtraArgs: []string{"--original"}, mode: "default"}
	snapshot := a.WorkspaceAgentOptions()
	command, ok := snapshot["cmd"].([]string)
	if !ok {
		t.Fatalf("command snapshot type = %T, want []string", snapshot["cmd"])
	}
	command[1] = "--changed"
	if a.cliExtraArgs[0] != "--original" {
		t.Fatal("mutating the workspace snapshot changed the project command")
	}
}
