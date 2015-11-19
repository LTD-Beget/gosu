package main // import "github.com/LTD-Beget/gosu"

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"flag"
)

const (
	MIN_UID = 999
	MIN_GID = 599
)

func init() {
	// use seccomp if not root
	if syscall.Geteuid() > 0 {
		initSeccomp()
	}

	// make sure we only have one process and that it runs on the main thread (so that ideally, when we Exec, we keep our user switches and stuff)
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
}

func main() {
	log.SetFlags(0) // no timestamps on our logs

	var changeDir string

	flag.StringVar(&changeDir, "chdir", "", "Change dir after setuid")
	flag.Parse()

	args := flag.Args()

	if len(args) < 2 {
		log.Printf("Usage: %s user-spec command [args]", os.Args[0])
		log.Printf("   ie: %s tianon bash", os.Args[0])
		log.Printf("       %s nobody:root bash -c 'whoami && id'", os.Args[0])
		log.Printf("       %s 1000:1 id", os.Args[0])
		log.Println()
		log.Printf("%s version: %s (%s on %s/%s; %s)", os.Args[0], Version, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
		log.Println()
		os.Exit(1)
	}

	// clear HOME so that SetupUser will set it
	os.Setenv("HOME", "")

	// setup user
	if err := SetupUser(args[0]); err != nil {
		log.Fatalf("error: failed switching to %q: %v", args[0], err)
	}

	// search executable
	name, err := exec.LookPath(args[1])
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// change dir before exec
	if changeDir == "" {
		changeDir = os.Getenv("HOME")
	}

	if (changeDir != "") {
		os.Chdir(changeDir)
	}

	// call execve
	if err = syscall.Exec(name, args[1:], os.Environ()); err != nil {
		log.Fatalf("error: exec failed: %v", err)
	}
}
