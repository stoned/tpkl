package tasks_test

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

//go:embed testdata/sighandler/tasks.pkl
var testSigHandlerModule string

//go:embed testdata/exitcode/tasks.pkl
var testExitCodeModule string

func buildTpkl(t *testing.T) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Minute)
	defer cancel()

	exe := filepath.Join(t.TempDir(), "_test_tpkl")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", exe, "..")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build ad hoc test tpkl executable: %s\n %s", err, output)
	}

	return exe
}

func setupTasksFile(t *testing.T, module *string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "tasks.pkl")

	err := os.WriteFile(path, []byte(*module), 0o644)
	if err != nil {
		t.Fatalf("cannot create temporary file: %q: %s", path, err)
	}

	return path
}

func TestSigHandler(t *testing.T) {
	t.Parallel()

	var err error

	tpklWitnessFile := filepath.Join(t.TempDir(), "witness")
	tpklExe := buildTpkl(t)
	tasksPath := setupTasksFile(t, &testSigHandlerModule)

	tpklArgs := []string{"run", "-m", tasksPath, "top"}
	tpklCmdStr := tpklExe + " " + strings.Join(tpklArgs, " ")
	tpklErrChan := make(chan error)
	tpklStderrChan := make(chan io.Reader)
	tpklStdoutChan := make(chan io.Reader)
	tpklProcessChan := make(chan *os.Process)

	go func() {
		var err error

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, tpklExe, tpklArgs...)
		cmd.Env = append(cmd.Environ(), "TEST_SIGHANDLER_WITNESS_FILE="+tpklWitnessFile)

		stderr, err := cmd.StderrPipe()
		if err != nil {
			tpklErrChan <- err

			return
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			tpklErrChan <- err

			return
		}

		err = cmd.Start()
		if err != nil {
			tpklErrChan <- err

			return
		}

		tpklProcessChan <- cmd.Process

		tpklStderrChan <- stderr

		tpklStdoutChan <- stdout

		err = cmd.Wait()

		tpklErrChan <- err

		close(tpklErrChan)
	}()

	tpklProcess := <-tpklProcessChan

	ticker := time.NewTicker(100 * time.Millisecond)

	var tpklStderr, tpklStdout io.Reader

	for events := 0; events < 3; {
		select {
		case err = <-tpklErrChan:
			if err != nil {
				t.Fatalf("error running tpkl command: %q: %s", tpklCmdStr, err)
			}
		case tpklStderr = <-tpklStderrChan:
			events++

		case tpklStdout = <-tpklStdoutChan:
			events++

		case <-ticker.C:
			_, err := os.Stat(tpklWitnessFile)
			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				t.Fatalf("error checking witness file %q: %s", tpklWitnessFile, err)
			}

			t.Logf("Witness file %q present", tpklWitnessFile)

			err = tpklProcess.Signal(syscall.SIGTERM)
			if err != nil {
				t.Fatalf("cannot send TERM signal to tpkl command: %s", err)
			}

			t.Logf("TERM signal sent to tpkl command")

			events++
		}
	}

	stdout, err := io.ReadAll(tpklStdout)
	if err != nil && !errors.Is(err, os.ErrClosed) {
		t.Fatalf("reading stdout from tpkl command %q: %s", tpklCmdStr, err)
	}

	stderr, err := io.ReadAll(tpklStderr)
	if err != nil && !errors.Is(err, os.ErrClosed) {
		t.Fatalf("reading stderr from tpkl command %q: %s", tpklCmdStr, err)
	}

	err = <-tpklErrChan

	outFormat := "\nStdout:\n%sStderr:\n%s"

	ee := (&exec.ExitError{})
	if errors.As(err, &ee) {
		expected := 128

		exitCode := ee.ExitCode()
		if exitCode != expected {
			t.Fatalf("unexpected exit code from tpkl command %q: expected %v, got %v"+outFormat,
				tpklCmdStr, expected, exitCode, stdout, stderr)
		}
	} else {
		t.Fatalf("unexpected error from tpkl command %q: %s"+outFormat,
			tpklCmdStr, err, stdout, stderr)
	}

	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "TASKFILE ") {
			taskfile := line[len("TASKFILE "):]

			_, err := os.Stat(taskfile)
			switch {
			case os.IsNotExist(err):
				continue
			case err == nil:
				t.Errorf("error task file still present %q", taskfile)
			default:
				t.Fatalf("unexpected error stating the taskfile %q: %s", taskfile, err)
			}
		}
	}
}

func TestExitCode(t *testing.T) {
	t.Parallel()

	var err error

	cases := []struct {
		args     []string
		exitCode int
	}{
		{[]string{"sh"}, 0},
		{[]string{"cmd"}, 0},
		{[]string{"sh", "-p", "code=2"}, 2},
		{[]string{"cmd", "-p", "code=2"}, 2},
		{[]string{"sh", "-p", "code=11"}, 11},
		{[]string{"cmd", "-p", "code=11"}, 11},
	}

	tpklExe := buildTpkl(t)
	tasksPath := setupTasksFile(t, &testExitCodeModule)

	for tcIdx, testCase := range cases {
		t.Run(fmt.Sprintf("%d:%d", tcIdx, testCase.exitCode), func(t *testing.T) {
			t.Parallel()

			tpklArgs := make([]string, 0, 3+len(testCase.args))

			tpklArgs = append(tpklArgs, "run", "-m", tasksPath)
			tpklArgs = append(tpklArgs, testCase.args...)
			tpklCmdStr := tpklExe + " " + strings.Join(tpklArgs, " ")

			ctx, cancel := context.WithTimeout(t.Context(), 1*time.Minute)
			defer cancel()

			cmd := exec.CommandContext(ctx, tpklExe, tpklArgs...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin

			err = cmd.Run()
			if testCase.exitCode == 0 {
				if err != nil {
					t.Errorf("unexpected error from tpkl command %q: %s", tpklCmdStr, err)
				}

				return
			}

			ee := (&exec.ExitError{})
			if errors.As(err, &ee) {
				got := ee.ExitCode()
				if got != testCase.exitCode {
					t.Errorf("wanted exit code %d from tpkl command %q, got %d", testCase.exitCode, tpklCmdStr, got)
				}
			} else {
				t.Errorf("unexpected error from tpkl command %q: %s", tpklCmdStr, err)
			}
		})
	}
}
