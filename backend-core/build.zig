//Zig build configuration
const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    // 1. Build Backend as a standalone Shared Library (.dll on Windows)
    const lib = b.addSharedLibrary(.{
        .name = "backend_core",
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });

    // Link C library dependencies directly
    lib.linkLibC();
    lib.addIncludePath(b.path("third_party"));
    // Example: If you have a custom C file you want to compile into your Zig core:
    // lib.addCSourceFile(.{ .file = b.path("third_party/custom_parser.c"), .flags = &.{"-std=c11"} });

    // Output directly to a shared bin directory at the project root
    const install_lib = b.addInstallArtifact(lib, .{
        .dest_dir = .{ .override = .{ .custom = "../bin" } }
    });
    b.getInstallStep().dependOn(&install_lib.step);

    // 2. Build Backend as a standalone Executable (Useful for IPC / Microservice model)
    const exe = b.addExecutable(.{
        .name = "backend_service",
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    exe.linkLibC();
    
    const install_exe = b.addInstallArtifact(exe, .{
        .dest_dir = .{ .override = .{ .custom = "../bin" } }
    });
    b.getInstallStep().dependOn(&install_exe.step);
}