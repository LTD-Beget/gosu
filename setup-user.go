package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/LTD-Beget/libcontainer/seccomp"
	"github.com/LTD-Beget/libcontainer/system"
	"github.com/LTD-Beget/libcontainer/user"
)

func initSeccomp() {
	context := seccomp.New()

	// limit minimum uid
	args_uid := make([][]seccomp.Arg, 1)
	args_uid[0] = make([]seccomp.Arg, 1)
	args_uid[0][0] = seccomp.Arg{
		Index: 0,
		Op:    seccomp.LessThan,
		Value: MIN_UID,
	}

	setuid := seccomp.Syscall{
		Value:  syscall.SYS_SETUID,
		Action: seccomp.Errno,
		Args:   args_uid,
	}

	// limit minimum gid
	args_gid := make([][]seccomp.Arg, 1)
	args_gid[0] = make([]seccomp.Arg, 1)
	args_gid[0][0] = seccomp.Arg{
		Index: 0,
		Op:    seccomp.LessThan,
		Value: MIN_GID,
	}

	setgid := seccomp.Syscall{
		Value:  syscall.SYS_SETGID,
		Action: seccomp.Errno,
		Args:   args_gid,
	}

	// apply seccomp
	context.Add(&setuid)
	context.Add(&setgid)
	context.Load()
}

// this function comes from libcontainer/init_linux.go
// we don't use that directly because we don't want the whole namespaces package imported here

// SetupUser changes the groups, gid, and uid for the user inside the container
func SetupUser(u string) error {
	// Set up defaults.
	defaultExecUser := user.ExecUser{
		Uid:  syscall.Getuid(),
		Gid:  syscall.Getgid(),
		Home: "/",
	}
	passwdPath, err := user.GetPasswdPath()
	if err != nil {
		return err
	}
	groupPath, err := user.GetGroupPath()
	if err != nil {
		return err
	}

	execUser, err := user.GetExecUserPath(u, &defaultExecUser, passwdPath, groupPath)
	if err != nil {
		return fmt.Errorf("get supplementary groups %s", err)
	}

	// if not root - check uid/gid by hand if seccomp is not working
	if syscall.Geteuid() > 0 && (execUser.Uid <= MIN_UID || execUser.Gid <= MIN_GID) {
		return fmt.Errorf("Invalid UID or GID")
	}

	// set supplementary groups
	if err := syscall.Setgroups(execUser.Sgids); err != nil {
		return fmt.Errorf("setgroups %s", err)
	}

	// set gid
	if err := system.Setgid(execUser.Gid); err != nil {
		return fmt.Errorf("setgid %s", err)
	}

	// check if setgid is successfull
	if syscall.Getgid() != execUser.Gid {
		return fmt.Errorf("setgid failed")
	}

	// set uid
	if err := system.Setuid(execUser.Uid); err != nil {
		return fmt.Errorf("setuid %s", err)
	}

	// check if setuid is successful
	if syscall.Getuid() != execUser.Uid {
		return fmt.Errorf("setuid failed")
	}

	// if we didn't get HOME already, set it based on the user's HOME
	if envHome := os.Getenv("HOME"); envHome == "" {
		if err := os.Setenv("HOME", execUser.Home); err != nil {
			return fmt.Errorf("set HOME %s", err)
		}
	}
	return nil
}
