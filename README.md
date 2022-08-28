# Bazel Tips

A collection of useful Bazel commands and tips, with some explanation what is happening, demonstrated via a simple example Go project.

To get started, download [Bazelisk](https://github.com/bazelbuild/bazelisk) and put it in your path as `bazel`. Bazelisk will automatically download and manage the right version of `bazel`.

Useful links:
* [Intro to Bazel](https://bazel.build/about/intro)
* [Bazel Concepts](https://bazel.build/concepts/build-ref)
* [Bazel Cheatsheet](https://skia.googlesource.com/buildbot/+/main/BAZEL_CHEATSHEET.md)

## Basics

Build everything:

    bazel build //...

Run all tests:

    bazel test //...

## Build

Bazel builds targets via sandboxing, with only the declared source files and dependencies of the build rule available inside the sandbox. Using tmpfs on Linux can speed up creating sandboxes significantly:

    bazel build --sandbox_base=/dev/shm //...

To debug build errors, showing what happened in the sandbox might be helpful:

    bazel build --sandbox_debug //...

By default, bazel will only show whether a build target was up to date and the outputs build when individual targets are built. To show it even when multiple targets are build:

    bazel build --show_result=999999 //...

## Run

You can run executable targets via:

    bazel run //cmd/bazel-tips
    INFO: Analyzed target //cmd/bazel-tips:bazel-tips (0 packages loaded, 0 targets configured).
    INFO: Found 1 target...
    Target //cmd/bazel-tips:bazel-tips up-to-date:
      bazel-bin/cmd/bazel-tips/bazel-tips_/bazel-tips
    INFO: Elapsed time: 0.122s, Critical Path: 0.00s
    INFO: 1 process: 1 internal.
    INFO: Build completed successfully, 1 total action
    INFO: Build completed successfully, 1 total action
    Doing something
    Hello, Bazel!

To provide command line flags to the executable:

    bazel run //cmd/bazel-tips -- --help

## Tests

Show streamed output from tests:

    bazel test --test_output=streamed //...

Use `--test_output=all` to see the full output:

    bazel test --test_output=all //...

Discard cached result:

    bazel test --nocache_test_results //...

To debug flaky tests, you can ask Bazel to re-run the test multiple times, and show how many times it failed:

    bazel test --runs_per_test=10 //...

Provide additional arguments to the test runner underneath:

    bazel test --test_output=streamed --test_arg=-test.v //...

`//` is the root of the Bazel repository, and `...` means every target recursively. You can build or test only specific packages via:

    bazel test //pkg/mypackage:mypackage_test

Bazel runs tests without environment variables in the host system. You need to explicitly specify environment variables that will be available for a test:

    bazel test --test_env=FOO=bar //pkg/mypackage:mypackage_test

To set an environment variable to the value from the host system:

    bazel test --test_env=FOO //pkg/mypackage:mypackage_test

## Packages, Rules and Targets

Using `//pkg/mypackage` tells Bazel which package we want to test, and `:mypackage_test` is a label of a test target in that package. A package is a directory with source files and a BUILD or BUILD.bazel file, specifying how to build those source files. Any subdirectory of a package also belongs to that package, unless there's a BUILD file in it (thus making it a package itself). You need to use the full path for a subpackage, and can't use `..` to reference a package up a level.

If you look into our hypothetical BUILD (or BUILD.bazel) file, you will find a test target with the name `mypackage_test`:

    go_test(
        name = "mypackage_test",
        srcs = ["mypackage_test.go"],
        embed = [":mypackage"],
    )

Here, `go_test` is a Bazel rule. It is loaded from [rules_go](https://github.com/bazelbuild/rules_go) via a load statement at the beginning of the BUILD file:

    load("@io_bazel_rules_go//go:def.bzl", "go_test")

Via rules, we can define Bazel targets (like `mypackage_test` above). Bazel targets must have a unique name in their package. Source files are targets too. So, for example `//pkg/mypackage:mypackage_test.go` is a target, just like the test generated from it, `//pkg/mypackage:mypackage_test`.

You can reference targets in the same package via the abbreviated form `:<target_name>`. For example, the `go_test` target above references a `go_library` target from the same package:

    go_library(
        name = "mypackage",
        srcs = ["mypackage.go"],
        importpath = "github.com/ldx/bazel_tips/pkg/mypackage",
        visibility = ["//visibility:public"],
    )

Targets have visibility: this defines whether rules from other packages can also use the target. Packages can also declare a default visibility for their targets. Setting `visibility = ["//visibility:public"]` tells Bazel that any other package can use the target `//pkg/mypackage:mypackage`.

## Workspace

The rules from `rules_go` are installed via the WORKSPACE file. A directory with a WORKSPACE file is considered the root of a Bazel repository (which is different from a Git repository), also called @. Other, external repositories are defined in the WORKSPACE file using workspace rules.

When referencing packages and targets, we started from the root of the Bazel repository, `//`. A WORKSPACE file usually starts via declaring the name of it:

    workspace(name = "bazel_tips")

You can use that when referencing packages and targets:

    bazel test @bazel_tips//pkg/mypackage:mypackage_test

This is not necessary for our own packages, but repository rules in WORKSPACE files create external repositories, and sometimes we want to reference packages and targets in those external repositories. For example, `rules_go` creates an external repository called `@go_sdk` when installing it via WORKSPACE. You can use this SDK from your repository, for example, to use `gofmt` from the SDK downloaded by `rules_go`:

    bazel run @go_sdk//:bin/gofmt -- --help
    INFO: Analyzed target @go_sdk//:bin/gofmt (0 packages loaded, 0 targets configured).
    INFO: Found 1 target...
    INFO: Elapsed time: 0.144s, Critical Path: 0.00s
    INFO: 1 process: 1 internal.
    INFO: Build completed successfully, 1 total action
    INFO: Running command line: /home/vilmos/.cache/bazel/_bazel_vilmos/aa9d8e47d6f7444801INFO: Build completed successfully, 1 total action
    usage: gofmt [flags] [path ...]
      -cpuprofile string
            write cpu profile to this file
      -d    display diffs instead of rewriting files
      -e    report all errors (not just the first 10 on different lines)
      -l    list files whose formatting differs from gofmt's
      -r string
            rewrite rule (e.g., 'a[b:len(a)] -> a[b:]')
      -s    simplify code
      -w    write result to (source) file instead of stdout

You can use:

    bazel fetch //...

to download all dependencies required to build the repository.

## Queries

Bazel comes with a powerful query language.

To list the dependencies of a target:

    bazel query 'deps(//pkg/mypackage:mypackage)'
    //pkg/mypackage:mypackage
    //pkg/mypackage:mypackage.go
    @bazel_tools//tools/allowlists/function_transition_allowlist:function_transition_allowlist
    @bazel_tools//tools/build_defs/cc/whitelists/parse_headers_and_layering_check:disabling_parse_headers_and_layering_check_allowed
    @bazel_tools//tools/build_defs/cc/whitelists/starlark_hdrs_check:loose_header_check_allowed_in_toolchain
    @bazel_tools//tools/cpp:build_interface_so
    @bazel_tools//tools/cpp:current_cc_toolchain
    @bazel_tools//tools/cpp:interface_library_builder
    @bazel_tools//tools/cpp:link_dynamic_library
    @bazel_tools//tools/cpp:link_dynamic_library.sh
    @bazel_tools//tools/cpp:toolchain
    @bazel_tools//tools/objc:host_xcodes
    @bazel_tools//tools/osx:current_xcode_config
    @bazel_tools//tools/whitelists/function_transition_whitelist:function_transition_whitelist
    @io_bazel_rules_go//:cgo_context_data
    @io_bazel_rules_go//:cgo_context_data_proxy
    @io_bazel_rules_go//:default_nogo
    @io_bazel_rules_go//:go_config
    @io_bazel_rules_go//:go_context_data
    @io_bazel_rules_go//:stdlib
    @io_bazel_rules_go//go/config:cover_format
    @io_bazel_rules_go//go/config:debug
    @io_bazel_rules_go//go/config:linkmode
    @io_bazel_rules_go//go/config:msan
    @io_bazel_rules_go//go/config:pure
    @io_bazel_rules_go//go/config:race
    @io_bazel_rules_go//go/config:static
    @io_bazel_rules_go//go/config:strip
    @io_bazel_rules_go//go/config:tags
    @io_bazel_rules_go//go/platform:internal_cgo_off
    @io_bazel_rules_go//go/private:always_true
    @io_bazel_rules_go//go/private:is_compilation_mode_dbg
    @io_bazel_rules_go//go/private:stamp
    @io_bazel_rules_go//go/toolchain:cgo_constraint
    @io_bazel_rules_go//go/toolchain:cgo_off
    @io_bazel_rules_go//go/tools/coverdata:coverdata
    @io_bazel_rules_go//go/tools/coverdata:coverdata.go
    @io_bazel_rules_nogo//:nogo
    @local_config_cc//:builtin_include_directory_paths
    @local_config_cc//:cc-compiler-armeabi-v7a
    @local_config_cc//:cc-compiler-k8
    @local_config_cc//:compiler_deps
    @local_config_cc//:empty
    @local_config_cc//:local
    @local_config_cc//:stub_armeabi-v7a
    @local_config_cc//:toolchain

List all packages under `pkg`:

    bazel query --output package //pkg/...
    pkg/mypackage

Query the outputs generated by a rule:

    bazel cquery --output=files //pkg/mypackage:mypackage
    INFO: Analyzed target //pkg/mypackage:mypackage (0 packages loaded, 0 targets configured).
    INFO: Found 1 target...
    bazel-out/k8-fastbuild/bin/pkg/mypackage/mypackage.a

Sometime a dependency can not be resolved; the `--keep_going` flag tells bazel to keep going and ignore errors.

To show reverse dependencies:

    bazel query 'rdeps(..., //pkg/mypackage:mypackage)'
    //cmd/bazel-tips:bazel-tips
    //cmd/bazel-tips:bazel-tips_lib
    //pkg/mypackage:mypackage
    //pkg/mypackage:mypackage_test

See the [Bazel Query How-To](https://bazel.build/query/quickstart) for a lot more useful examples.

## Bazel Info

Bazel runtime information and configuration can be obtained via `bazel info`.

To show the output base:

    bazel info output_base
    /home/vilmos/.cache/bazel/_bazel_vilmos/aa9d8e47d6f744480136e37f4c1ec205

This is the directory under which is the md5 hash of the path of the workspace root directory. A few interesting directories inside the output base:
* `external/` contains downloaded remote repositories.
* `execroot/` working directory for actions.
* `execroot/<workspace name>/bazel-out/` output from the build.
* `execroot/<workspace name>/bazel-out/bin` built binaries.

## Profiling

Bazel generates profiling data for each run. By default, this is saved in `$(bazel info output_base)/command.profile.gz`.

To show profiling data:


    bazel analyze-profile $(bazel info output_base)/command.profile.gz
    WARNING: This information is intended for consumption by Bazel developers only, and may change at any time. Script against it at your own risk
    INFO: Profile created on Sun Aug 28 11:47:14 PDT 2022, build ID: ded80479-54a4-4ae1-b866-f714e3e13639, output base: /home/vilmos/.cache/bazel/_bazel_vilmos/aa9d8e47d6f744480136e37f4c1ec205
    
    === PHASE SUMMARY INFORMATION ===
    
    Total launch phase time         0.009 s    0.50%
    Total init phase time           0.021 s    1.21%
    Total target pattern evaluation phase time    0.125 s    6.98%
    Total interleaved loading-and-analysis phase time    0.710 s   39.52%
    Total preparation phase time    0.002 s    0.12%
    Total execution phase time      0.923 s   51.38%
    Total finish phase time         0.005 s    0.28%
    ------------------------------------------------
    Total run time                  1.797 s  100.00%
    
    Critical path (896 ms):
           Time Percentage   Description
         487 ms   54.31%   action 'GoToolchainBinaryCompile external/go_sdk/builder.a'
         168 ms   18.71%   action 'GoToolchainBinary external/go_sdk/builder'
         113 ms   12.66%   action 'GoCompilePkg external/io_bazel_rules_go/go/tools/bzltestutil/bzltestutil.a'
        30.4 ms    3.39%   action 'GoCompilePkg pkg/mypackage/mypackage_test~testmain.a'
        97.9 ms   10.92%   action 'GoLink pkg/mypackage/mypackage_test_/mypackage_test'
        0.07 ms    0.01%   runfiles for //pkg/mypackage mypackage_test
