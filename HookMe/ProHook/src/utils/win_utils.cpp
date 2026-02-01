// Copyright 2026 HookMe Authors.
#include "ProHook/utils/win_utils.h"
#include <fstream>

namespace prohook {
    namespace utils {

        ProcessArch WinUtils::GetProcessArchitecture(HANDLE h_process) {
            if (!h_process) return ProcessArch::Unknown;

            BOOL is_wow64 = FALSE;
            if (!IsWow64Process(h_process, &is_wow64)) {
                return ProcessArch::Unknown;
            }

            if (is_wow64) {
                // Target is 32-bit running on 64-bit OS
                return ProcessArch::x86;
            }

            // If not WoW64, it's either native x64 or native x86 on a 32-bit OS
            SYSTEM_INFO sys_info;
            GetNativeSystemInfo(&sys_info);

            if (sys_info.wProcessorArchitecture == PROCESSOR_ARCHITECTURE_AMD64) {
                return ProcessArch::x64;
            }
            else if (sys_info.wProcessorArchitecture == PROCESSOR_ARCHITECTURE_INTEL) {
                return ProcessArch::x86;
            }

            return ProcessArch::Unknown;
        }

        ProcessArch WinUtils::GetFileArchitecture(const std::filesystem::path& file_path) {
            std::ifstream file(file_path, std::ios::binary);
            if (!file) return ProcessArch::Unknown;

            // Read DOS Header
            IMAGE_DOS_HEADER dos_hdr;
            if (!file.read(reinterpret_cast<char*>(&dos_hdr), sizeof(dos_hdr)))
                return ProcessArch::Unknown;

            if (dos_hdr.e_magic != IMAGE_DOS_SIGNATURE)
                return ProcessArch::Unknown;

            // Seek to NT Headers (Signature + FileHeader)
            // We skip the 4-byte Signature (PE\0\0) to get straight to the FileHeader
            file.seekg(dos_hdr.e_lfanew + sizeof(DWORD));

            IMAGE_FILE_HEADER file_hdr;
            if (!file.read(reinterpret_cast<char*>(&file_hdr), sizeof(file_hdr)))
                return ProcessArch::Unknown;

            if (file_hdr.Machine == IMAGE_FILE_MACHINE_AMD64) {
                return ProcessArch::x64;
            }
            else if (file_hdr.Machine == IMAGE_FILE_MACHINE_I386) {
                return ProcessArch::x86;
            }

            return ProcessArch::Unknown;
        }

        ProcessArch WinUtils::GetCurrentProcessArchitecture() {
#ifdef _WIN64
            return ProcessArch::x64;
#else
            return ProcessArch::x86;
#endif
        }

        bool WinUtils::Is64BitOperatingSystem() {
            SYSTEM_INFO sys_info;
            GetNativeSystemInfo(&sys_info);
            return sys_info.wProcessorArchitecture == PROCESSOR_ARCHITECTURE_AMD64;
        }
    }  // namespace utils
}  // namespace prohook