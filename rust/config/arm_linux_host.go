// Copyright 2022 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"strings"

	"android/soong/android"
	cc_config "android/soong/cc/config"
)

var (
	linuxArmRustflags   = []string{}
	linuxArmLinkflags   = []string{}
	linuxArm64Rustflags = []string{}
	linuxArm64Linkflags = []string{}
)

func init() {
	registerToolchainFactory(android.LinuxMusl, android.Arm64, linuxMuslArm64ToolchainFactory)
	registerToolchainFactory(android.LinuxMusl, android.Arm, linuxMuslArmToolchainFactory)

	pctx.StaticVariable("LinuxToolchainArmRustFlags", strings.Join(linuxArmRustflags, " "))
	pctx.StaticVariable("LinuxToolchainArmLinkFlags", strings.Join(linuxArmLinkflags, " "))
	pctx.StaticVariable("LinuxToolchainArm64RustFlags", strings.Join(linuxArm64Rustflags, " "))
	pctx.StaticVariable("LinuxToolchainArm64LinkFlags", strings.Join(linuxArm64Linkflags, " "))
}

// Base 64-bit linux rust toolchain
type toolchainLinuxArm64 struct {
	toolchain64Bit
}

func (toolchainLinuxArm64) Supported() bool {
	return true
}

func (toolchainLinuxArm64) Bionic() bool {
	return false
}

func (t *toolchainLinuxArm64) Name() string {
	return "arm64"
}

func (t *toolchainLinuxArm64) ToolchainLinkFlags(ctx ToolchainFlagsContext) cc_config.FlagsWithDeps {
	// Prepend the lld flags from cc_config so we stay in sync with cc
	return cc_config.FlagsWithDeps{
		Flags: "${cc_config.LinuxLdflags} ${cc_config.LinuxArm64Ldflags} " +
			"${config.LinuxToolchainLinkFlags} ${config.LinuxToolchainArm64LinkFlags}",
	}
}

func (t *toolchainLinuxArm64) ToolchainRustFlags(ctx ToolchainFlagsContext) cc_config.FlagsWithDeps {
	return LinuxToolchainRustFlags(ctx).AppendNoDeps("${config.LinuxToolchainArm64RustFlags}")
}

// Specialization of the 64-bit linux rust toolchain for musl.  Adds the musl rust triple and
// linker flags to avoid using the host sysroot.
type toolchainLinuxMuslArm64 struct {
	toolchainLinuxArm64
}

func (t *toolchainLinuxMuslArm64) RustTriple() string {
	// Host rust on arm64 uses glibc: switch from the (unavailable for this
	// target) musl stdlib to a gnu stdlib so that rustc's host==target triple
	// check passes for proc-macro crates.
	return "aarch64-unknown-linux-gnu"
}

// Despite the "Musl" name this toolchain now targets the host glibc
// (RustTriple is aarch64-unknown-linux-gnu), so report Musl()=false /
// Glibc()=true. This makes rust/binary.go, rust/library.go and rust/bindgen.go
// skip the musl crt / libc_musl / clang_rt injection for host bins and
// dylibs (otherwise the rsp carries musl crt objects while the clang driver
// also adds the host glibc crt, producing duplicate _start/_init/_fini/
// __dso_handle symbols at link time).
func (t *toolchainLinuxMuslArm64) Musl() bool {
	return false
}

func (t *toolchainLinuxMuslArm64) Glibc() bool {
	return true
}


func (t *toolchainLinuxMuslArm64) ToolchainLinkFlags(ctx ToolchainFlagsContext) cc_config.FlagsWithDeps {
	// glibc host rust on an aarch64 Linux host links against the host system
	// glibc/gcc: target arch == host arch == aarch64, so clang++'s default
	// toolchain resolution finds the distro's crtbeginS.o/crtendS.o, libgcc_s
	// and glibc. The base toolchainLinkFlagsArrayArm64-LinuxLdflags carries
	// --gcc-toolchain=<x86_64 host gcc> (this tree is x86_64-host oriented),
	// which cannot supply aarch64 crt objects - that made ld.lld fail with
	// "cannot open crtbeginS.o". So drop LinuxLdflags/LinuxArm64Ldflags and the
	// x86 gcc -L search dirs from ToolchainRustFlags below.
	//
	// Keep the shared safety linker flags and add the glibc library set
	// (-lc -lrt -ldl -lpthread -lm -lgcc_s). LinuxHostGlobalLinkFlags supplies
	// the same set but only for Os()==Linux builds, not for this LinuxMusl
	// variant. rustc itself emits -nodefaultlibs (via -Z link-native-libraries=no),
	// so the explicit set is required.
	return cc_config.FlagsWithDeps{
		Flags: "-Wl,-z,noexecstack -Wl,-z,relro -Wl,-z,now -Wl,--no-undefined-version " +
			"-Wl,--compress-debug-sections=zstd -Wl,--build-id=md5 " +
			"${config.LinuxToolchainLinkFlags} ${config.LinuxToolchainArm64LinkFlags} " +
			"-lc -lrt -ldl -lpthread -lm -lgcc_s",
	}
}

func (t *toolchainLinuxMuslArm64) ToolchainRustFlags(ctx ToolchainFlagsContext) cc_config.FlagsWithDeps {
	// Do NOT add the x86_64 host-gcc -L search dirs (LinuxToolchainRustFlags):
	// libc/gcc for the aarch64 host resolve from the host system instead.
	return cc_config.FlagsWithDeps{}
}

func linuxMuslArm64ToolchainFactory(arch android.Arch) Toolchain {
	return toolchainLinuxMuslArm64Singleton
}

// Base 32-bit linux rust toolchain
type toolchainLinuxArm struct {
	toolchain32Bit
}

func (toolchainLinuxArm) Supported() bool {
	return true
}

func (toolchainLinuxArm) Bionic() bool {
	return false
}

func (t *toolchainLinuxArm) Name() string {
	return "arm"
}

func (toolchainLinuxArm) LibclangRuntimeLibraryArch() string {
	return "arm"
}

func (toolchainLinuxArm64) LibclangRuntimeLibraryArch() string {
	return "arm64"
}

func (t *toolchainLinuxArm) ToolchainLinkFlags(ctx ToolchainFlagsContext) cc_config.FlagsWithDeps {
	// Prepend the lld flags from cc_config so we stay in sync with cc
	return cc_config.FlagsWithDeps{
		Flags: "${cc_config.LinuxLdflags} ${cc_config.LinuxArmLdflags} " +
			"${config.LinuxToolchainLinkFlags} ${config.LinuxToolchainArmLinkFlags}",
	}
}

func (t *toolchainLinuxArm) ToolchainRustFlags(ctx ToolchainFlagsContext) cc_config.FlagsWithDeps {
	return LinuxToolchainRustFlags(ctx).AppendNoDeps("${config.LinuxToolchainArmRustFlags}")
}

// Specialization of the 32-bit linux rust toolchain for musl.  Adds the musl rust triple and
// linker flags to avoid using the host sysroot.
type toolchainLinuxMuslArm struct {
	toolchainLinuxArm
}

func (t *toolchainLinuxMuslArm) RustTriple() string {
	return "arm-unknown-linux-musleabihf"
}

func (t *toolchainLinuxMuslArm) ToolchainLinkFlags(ctx ToolchainFlagsContext) cc_config.FlagsWithDeps {
	linuxMusl := cc_config.FlagsWithDeps{
		Flags: "${config.LinuxMuslToolchainLinkFlags}",
	}
	return t.toolchainLinuxArm.ToolchainLinkFlags(ctx).Append(linuxMusl)
}

func (t *toolchainLinuxMuslArm) ToolchainRustFlags(ctx ToolchainFlagsContext) cc_config.FlagsWithDeps {
	return t.toolchainLinuxArm.ToolchainRustFlags(ctx).AppendNoDeps("${config.LinuxMuslToolchainRustFlags}")
}

func linuxMuslArmToolchainFactory(arch android.Arch) Toolchain {
	return toolchainLinuxMuslArmSingleton
}

var toolchainLinuxMuslArm64Singleton Toolchain = &toolchainLinuxMuslArm64{}
var toolchainLinuxMuslArmSingleton Toolchain = &toolchainLinuxMuslArm{}
