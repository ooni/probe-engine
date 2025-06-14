package main

//
// Building C dependencies: tor
//
// Adapted from https://github.com/guardianproject/tor-android
// SPDX-License-Identifier: BSD-3-Clause
//

import (
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/pkg/cmd/buildtool/internal/buildtoolmodel"
	"github.com/ooni/probe-engine/pkg/must"
	"github.com/ooni/probe-engine/pkg/runtimex"
	"github.com/ooni/probe-engine/pkg/shellx"
)

// cdepsTorBuildMain is the script that builds tor.
func cdepsTorBuildMain(globalEnv *cBuildEnv, deps buildtoolmodel.Dependencies) {
	topdir := deps.AbsoluteCurDir() // must be mockable
	work := cdepsMustMkdirTemp()
	restore := cdepsMustChdir(work)
	defer restore()

	// See https://github.com/Homebrew/homebrew-core/blob/master/Formula/t/tor.rb
	cdepsMustFetch("https://www.torproject.org/dist/tor-0.4.8.16.tar.gz")
	deps.VerifySHA256( // must be mockable
		"6540dd377a120fb8e7d27530aa3b7ff72a0fa5b4f670fe1d64c987c1cfd390cb",
		"tor-0.4.8.16.tar.gz",
	)
	must.Run(log.Log, "tar", "-xf", "tor-0.4.8.16.tar.gz")
	_ = deps.MustChdir("tor-0.4.8.16") // must be mockable

	mydir := filepath.Join(topdir, "CDEPS", "tor")
	for _, patch := range cdepsMustListPatches(mydir) {
		must.Run(log.Log, "git", "apply", patch)
	}

	must.Run(log.Log, "autoreconf", "-vif")

	envp := cBuildExportAutotools(globalEnv)

	argv := runtimex.Try1(shellx.NewArgv("./configure"))
	if globalEnv.CONFIGURE_HOST != "" {
		argv.Append("--host=" + globalEnv.CONFIGURE_HOST)
	}
	argv.Append(
		"--enable-pic",
		"--enable-static-libevent", "--with-libevent-dir="+globalEnv.DESTDIR,
		"--enable-static-openssl", "--with-openssl-dir="+globalEnv.DESTDIR,
		"--enable-static-zlib", "--with-zlib-dir="+globalEnv.DESTDIR,
		"--disable-module-dirauth",
		"--disable-zstd", "--disable-lzma",
		"--disable-tool-name-check",
		"--disable-systemd",
		"--prefix=/",
		"--disable-unittests",
		"--disable-system-torrc",
		"--disable-seccomp",
	)
	runtimex.Try0(shellx.RunEx(defaultShellxConfig(), argv, envp))

	must.Run(log.Log, "make", "V=1", "-j", strconv.Itoa(runtime.NumCPU()))

	must.Run(log.Log, "install", "-m644", "src/feature/api/tor_api.h", globalEnv.DESTDIR+"/include")
	must.Run(log.Log, "install", "-m644", "libtor.a", globalEnv.DESTDIR+"/lib")
}
