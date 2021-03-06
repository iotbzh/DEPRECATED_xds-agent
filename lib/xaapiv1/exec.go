/*
 * Copyright (C) 2017 "IoT.bzh"
 * Author Sebastien Douheret <sebastien@iot.bzh>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xaapiv1

type (
	// ExecArgs JSON parameters of /exec command
	ExecArgs struct {
		ID              string   `json:"id" binding:"required"`
		SdkID           string   `json:"sdkID"` // sdk ID to use for setting env
		CmdID           string   `json:"cmdID"` // command unique ID
		Cmd             string   `json:"cmd" binding:"required"`
		Args            []string `json:"args"`
		Env             []string `json:"env"`
		RPath           string   `json:"rpath"`           // relative path into project
		TTY             bool     `json:"tty"`             // Use a tty, specific to gdb --tty option
		TTYGdbserverFix bool     `json:"ttyGdbserverFix"` // Set to true to activate gdbserver workaround about inferior output
		ExitImmediate   bool     `json:"exitImmediate"`   // when true, exit event sent immediately when command exited (IOW, don't wait file synchronization)
		CmdTimeout      int      `json:"timeout"`         // command completion timeout in Second
	}

	// ExecResult JSON result of /exec command
	ExecResult struct {
		Status string `json:"status"` // status OK
		CmdID  string `json:"cmdID"`  // command unique ID
	}

	// ExecSignalResult JSON result of /signal command
	ExecSignalResult struct {
		Status string `json:"status"` // status OK
		CmdID  string `json:"cmdID"`  // command unique ID
	}

	// ExecInMsg Message used to received input characters (stdin)
	ExecInMsg struct {
		CmdID     string `json:"cmdID"`
		Timestamp string `json:"timestamp"`
		Stdin     string `json:"stdin"`
	}

	// ExecOutMsg Message used to send output characters (stdout+stderr)
	ExecOutMsg struct {
		CmdID     string `json:"cmdID"`
		Timestamp string `json:"timestamp"`
		Stdout    string `json:"stdout"`
		Stderr    string `json:"stderr"`
	}

	// ExecExitMsg Message sent when executed command exited
	ExecExitMsg struct {
		CmdID     string `json:"cmdID"`
		Timestamp string `json:"timestamp"`
		Code      int    `json:"code"`
		Error     error  `json:"error"`
	}

	// ExecSignalArgs JSON parameters of /exec/signal command
	ExecSignalArgs struct {
		CmdID  string `json:"cmdID" binding:"required"`  // command id
		Signal string `json:"signal" binding:"required"` // signal number
	}
)

const (
	// ExecInEvent Event send in WS when characters are sent (stdin)
	ExecInEvent = "exec:input"

	// ExecOutEvent Event send in WS when characters are received (stdout or stderr)
	ExecOutEvent = "exec:output"

	// ExecExitEvent Event send in WS when program exited
	ExecExitEvent = "exec:exit"

	// ExecInferiorInEvent Event send in WS when characters are sent to an inferior (used by gdb inferior/tty)
	ExecInferiorInEvent = "exec:inferior-input"

	// ExecInferiorOutEvent Event send in WS when characters are received by an inferior
	ExecInferiorOutEvent = "exec:inferior-output"
)
