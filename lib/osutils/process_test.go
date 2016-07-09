// SILVER - Service Wrapper
//
// Copyright (c) 2016 PaperCut Software http://www.papercut.com/
// Use of this source code is governed by an MIT or GPL Version 2 license.
// See the project's LICENSE file for more information.
//

package osutils_test

import (
	"os/exec"
	"runtime"
	"testing"
	"time"
	"syscall"

	"github.com/papercutsoftware/silver/lib/osutils"
)

func Test_ProcessKillGracefully_ConsoleProgram(t *testing.T) {
	var testCmd string
	var testArgs []string
	if runtime.GOOS == "windows" {
		testCmd = `c:\Windows\System32\msg.exe`
		testArgs = []string{"*"}
	} else {
		testCmd = "ping"
		testArgs = []string{"localhost"}
	}
	testProcessKillGracefully(testCmd, testArgs, t)
}

func Test_ProcessKillGracefully_GUIProgram(t *testing.T) {
	var testCmd string
	var testArgs []string
	if runtime.GOOS == "windows" {
		testCmd = `c:\Windows\notepad.exe`
		testArgs = []string{}
	} else {
		testCmd = "/sbin/ping"
		testArgs = []string{"localhost"}
	}
	testProcessKillGracefully(testCmd, testArgs, t)
}

func testProcessKillGracefully(command string, args []string, t *testing.T) {
	t.Logf("Starting %v %v", command, args)
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	err := cmd.Start()
	// Give time to open
	time.Sleep(1 * time.Second)
	if err != nil {
		t.Fatalf("Error starting test cmd: %v", cmd)
	}
	start := time.Now()
	go func() {
		err := cmd.Wait()
		t.Logf("Cmd complete in %v : %v", time.Now().Sub(start), err)
	}()

	// Act
	err = osutils.ProcessKillGracefully(cmd.Process.Pid, 5*time.Second)

	// Assert no error
	if err != nil {
		t.Errorf("Cmd did not exit cleanly: %v", err)
	}

	duration := time.Now().Sub(start)
	t.Logf("Exit happened in %v", duration)

	// Assert time killed within 1 second
	if duration > 1*time.Second {
		t.Errorf("Expected kill to return quicker!")
	}
	// Assert time killed larger than 500 ms
	if duration < 500*time.Millisecond {
		t.Errorf("Expected it to take > 500 ms due to check!")
	}
}

func TestProcessIsRunning_DetectsRunning(t *testing.T) {
	// Arrange
	var testCmd string
	var testArgs []string
	if runtime.GOOS == "windows" {
		testCmd = `c:\Windows\System32\ping.exe`
		testArgs = []string{"-n", "1000", "localhost"}
	} else {
		testCmd = "ping"
		testArgs = []string{"localhost"}
	}
	cmd := exec.Command(testCmd, testArgs...)
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Error starting test cmd: %v", cmd)
	}
	time.Sleep(500 * time.Millisecond)

	pid := cmd.Process.Pid

	// Act
	running, err := osutils.ProcessIsRunning(pid)

	// Assert
	if err != nil {
		t.Errorf("ProcessIsRunning returned error: %v", err)
	}
	if !running {
		t.Errorf("Expected process at pid %d to be running", pid)
	}

	cmd.Process.Kill()
}

func TestProcessIsRunning_DetectsNotRunning(t *testing.T) {
	// Arrange
	var testCmd string
	var testArgs []string
	if runtime.GOOS == "windows" {
		testCmd = `c:\Windows\System32\ping.exe`
		testArgs = []string{"-n", "1000", "localhost"}
	} else {
		testCmd = "ping"
		testArgs = []string{"localhost"}
	}
	cmd := exec.Command(testCmd, testArgs...)
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Error starting test cmd: %v", cmd)
	}
	pid := cmd.Process.Pid
	go func() {
		time.Sleep(500 * time.Millisecond)
		cmd.Process.Kill()
	}()
	cmd.Wait()

	// Act
	running, err := osutils.ProcessIsRunning(pid)

	// Assert
	if err != nil {
		t.Errorf("ProcessIsRunning returned error: %v", err)
	}
	if running {
		t.Errorf("Expected process at pid %d NOT to be running", pid)
	}
}
