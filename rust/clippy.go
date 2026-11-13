// Copyright 2020 The Android Open Source Project
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

package rust

import (
	"android/soong/android"
	"android/soong/rust/config"
)

type ClippyProperties struct {
	// name of the lint set that should be used to validate this module.
	//
	// Possible values are "default" (for using a sensible set of lints
	// depending on the module's location), "android" (for the strictest
	// lint set that applies to all Android platform code), "vendor" (for a
	// relaxed set) and "none" (to disable the execution of clippy).  The
	// default value is "default". See also the `lints` property.
	Clippy_lints *string
}

type clippy struct {
	Properties ClippyProperties
}

func (c *clippy) props() []any {
	return []any{&c.Properties}
}

func (c *clippy) flags(ctx ModuleContext, flags Flags, deps PathDeps) (Flags, PathDeps) {
	dirEnabled, lints, err := config.ClippyLintsForDir(ctx.ModuleDir(), c.Properties.Clippy_lints)
	if err != nil {
		ctx.PropertyErrorf("clippy_lints", err.Error())
	}

	envDisable := ctx.Config().IsEnvTrue("SOONG_DISABLE_CLIPPY")

	// arm64 宿主（BuildOS==LinuxMusl）的预置 rust 工具链
	// prebuilts/rust-toolchain/linux-musl-arm64（linux-arm64 为其符号链接）只带
	// rustc/rustdoc，没有 clippy-driver；linux-musl-x86 的 clippy-driver 是 x86
	// 二进制，无法在 arm64 机器执行。clippy 仅是 lint，与 system.img 产物无关。
	// 若不在分析期禁用 clippy，ninja 会生成 <crate>.clippy 校验目标，调度时因
	// "clippy-driver missing and no known rule to make it" 直接失败
	// （实例：build/make/tools/aconfig/aconfig[linux_musl_arm64] aconfig.clippy）。
	// 因此在 musl 宿主上无条件禁用 clippy；glibc(x86_64) 宿主不受影响（clippy 照常）。
	envDisable = envDisable || ctx.Config().BuildOS == android.LinuxMusl

	flags.Clippy = dirEnabled && !envDisable
	flags.ClippyFlags = append(flags.ClippyFlags, lints)
	return flags, deps
}
