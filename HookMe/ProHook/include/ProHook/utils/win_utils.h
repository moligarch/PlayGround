// Copyright 2026 HookMe Authors.
#ifndef PROHOOK_INCLUDE_UTILS_WIN_UTILS_H_
#define PROHOOK_INCLUDE_UTILS_WIN_UTILS_H_

#include <filesystem>

#include <windows.h>

namespace prohook {
    namespace utils {

        enum class ProcessArch {
            x86,
            x64,
            Unknown
        };

        class WinUtils {
        public:
            // Returns the architecture of a remote process
            static ProcessArch GetProcessArchitecture(HANDLE h_process);

            // Returns the architecture of a PE file (DLL/EXE) on disk
            static ProcessArch GetFileArchitecture(const std::filesystem::path& file_path);

            // Returns the architecture of the current (calling) process
            static ProcessArch GetCurrentProcessArchitecture();

            // Helper to check if the system itself is 64-bit
            static bool Is64BitOperatingSystem();
        };

    }  // namespace utils
}  // namespace prohook

#endif  // PROHOOK_INCLUDE_UTILS_WIN_UTILS_H_