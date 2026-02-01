// Copyright 2026 HookMe Authors.
#include "ProHook/injectors/standard_injector.h"
#include "ProHook/utils/win_utils.h"
#include "ProHook/utils/resolver.h"
#include <iostream>

namespace prohook {
    namespace injectors {

        bool StandardInjector::Inject(unsigned long process_id, const std::filesystem::path& dll_path) {
            HANDLE h_process = OpenProcess(
                PROCESS_QUERY_INFORMATION | PROCESS_CREATE_THREAD |
                PROCESS_VM_OPERATION | PROCESS_VM_WRITE | PROCESS_VM_READ,
                FALSE, process_id);

            if (!h_process) return false;

            // 1. Get Target Architecture
            utils::ProcessArch target_arch = utils::WinUtils::GetProcessArchitecture(h_process);

            // 2. Get DLL Architecture using our new Utility
            utils::ProcessArch dll_arch = utils::WinUtils::GetFileArchitecture(dll_path);

            // 3. Explicit Validation
            if (target_arch == utils::ProcessArch::Unknown || target_arch != dll_arch) {
                // Log mismatch: "Cannot inject [dll_arch] DLL into [target_arch] process."
                CloseHandle(h_process);
                return false;
            }

            auto full_path = std::filesystem::absolute(dll_path);
            if (!std::filesystem::exists(full_path)) {
                // std::wcerr << L"Required payload not found: " << full_path.wstring() << std::endl;
                CloseHandle(h_process);
                return false;
            }


            size_t path_size = (full_path.wstring().length() + 1) * sizeof(wchar_t);

            // 4. Allocate and Write memory
            void* remote_mem = VirtualAllocEx(h_process, nullptr, path_size, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
            if (!remote_mem || !WriteProcessMemory(h_process, remote_mem, full_path.c_str(), path_size, nullptr)) {
                if (remote_mem) VirtualFreeEx(h_process, remote_mem, 0, MEM_RELEASE);
                CloseHandle(h_process);
                return false;
            }

            // 5. Find LoadLibraryW address in the REMOTE process
            void* remote_k32 = utils::Resolver::GetSafeModuleHandle(h_process, L"kernel32.dll");
            void* load_lib_addr = utils::Resolver::GetSafeProcAddress(h_process, remote_k32, "LoadLibraryW");

            if (!load_lib_addr) {
                VirtualFreeEx(h_process, remote_mem, 0, MEM_RELEASE);
                CloseHandle(h_process);
                return false;
            }

            // 6. Execute remote thread
            HANDLE h_thread = CreateRemoteThread(h_process, nullptr, 0,
                (LPTHREAD_START_ROUTINE)load_lib_addr,
                remote_mem, 0, nullptr);

            bool success = false;
            if (h_thread) {
                WaitForSingleObject(h_thread, INFINITE);
                DWORD exit_code = 0;
                if (GetExitCodeThread(h_thread, &exit_code) && exit_code != 0) {
                    success = true;
                }
                CloseHandle(h_thread);
            }

            VirtualFreeEx(h_process, remote_mem, 0, MEM_RELEASE);
            CloseHandle(h_process);

            return success;
        }

    }  // namespace injectors
}  // namespace prohook