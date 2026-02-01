// Copyright 2026 HookMe Authors.
#include "ProHook/injectors/standard_injector.h"

#include "core/resolver.h"

namespace prohook::injectors {

    bool StandardInjector::Inject(unsigned long process_id, const std::filesystem::path& dll_path) {
        std::wstring path_str = std::filesystem::absolute(dll_path).wstring();
        size_t path_size = (path_str.length() + 1) * sizeof(wchar_t);

        HANDLE h_process = OpenProcess(PROCESS_ALL_ACCESS, FALSE, process_id);
        if (!h_process) return false;

        void* remote_mem = VirtualAllocEx(h_process, nullptr, path_size,
            MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);

        if (!remote_mem || !WriteProcessMemory(h_process, remote_mem, path_str.c_str(), path_size, nullptr)) {
            if (remote_mem) VirtualFreeEx(h_process, remote_mem, 0, MEM_RELEASE);
            CloseHandle(h_process);
            return false;
        }

        // --- LOGIC REPAIRED: Resolving LoadLibraryW directly in the REMOTE process ---
        void* remote_k32 = core::Resolver::GetSafeModuleHandle(h_process, L"kernel32.dll");

        if (!remote_k32) {
            // Fallback or handle error: Kernel32 should always be there.
            VirtualFreeEx(h_process, remote_mem, 0, MEM_RELEASE);
            CloseHandle(h_process);
            return false;
        }

        void* load_library_addr = core::Resolver::GetSafeProcAddress(h_process, remote_k32, "LoadLibraryW");

        if (!load_library_addr) {
            VirtualFreeEx(h_process, remote_mem, 0, MEM_RELEASE);
            CloseHandle(h_process);
            return false;
        }

        HANDLE h_thread = CreateRemoteThread(h_process, nullptr, 0,
            (LPTHREAD_START_ROUTINE)load_library_addr, remote_mem, 0, nullptr);

        if (!h_thread) {
            VirtualFreeEx(h_process, remote_mem, 0, MEM_RELEASE);
            CloseHandle(h_process);
            return false;
        }

        WaitForSingleObject(h_thread, INFINITE);

        CloseHandle(h_thread);
        VirtualFreeEx(h_process, remote_mem, 0, MEM_RELEASE);
        CloseHandle(h_process);

        return true;
    }

}  // namespace prohook::injectors