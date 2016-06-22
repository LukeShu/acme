package acmetool

import "os"

// The state directory path to be used for a system-wide state storage
// directory.  It may vary by system and platform.  On most POSIX-like
// systems, it is "/var/lib/acme".  Specific builds might customise
// it.
var DefaultStateDir string

// The hooks directory is the path at which executable hooks are
// looked for.  On POSIX-like systems, this is usually
// "/usr/lib/acme/hooks" (or "/usr/libexec/acme/hooks" if /usr/libexec
// exists).
var DefaultHooksDir string

// The standard webroot path, into which the responder always tries to
// install challenges, not necessarily successfully.  This is intended
// to be a standard, system-wide path to look for challenges at.  On
// POSIX-like systems, it is usually "/var/run/acme/acme-challenge".
var DefaultWebRootDir string

func init() {
	if DefaultStateDir == "" {
		DefaultStateDir = "/var/lib/acme"
	}
	if DefaultHooksDir == "" {
		if _, err := os.Stat("/usr/libexec"); err == nil {
			DefaultHooksDir = "/usr/libexec/acme/hooks"
		} else {
			DefaultHooksDir = "/usr/lib/acme/hooks"
		}
	}
	if DefaultWebRootDir == "" {
		DefaultWebRootDir = "/var/run/acme/acme-challenge"
	}
}
