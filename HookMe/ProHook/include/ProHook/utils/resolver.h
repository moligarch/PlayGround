// Copyright 2026 HookMe Authors.
#ifndef PROHOOK_INCLUDE_UTILS_RESOLVER_H_
#define PROHOOK_INCLUDE_UTILS_RESOLVER_H_

#include <windows.h>
#include <winternl.h>
#include <string>
#include <vector>

namespace prohook {
    namespace utils {

        // Precise definitions based on Geoff Chappell's research
        // This ensures we have the correct offsets regardless of winternl.h
        struct PRO_LDR_DATA_TABLE_ENTRY {
            LIST_ENTRY InLoadOrderLinks;
            LIST_ENTRY InMemoryOrderLinks;
            LIST_ENTRY InInitializationOrderLinks;
            PVOID DllBase;
            PVOID EntryPoint;
            ULONG SizeOfImage;
            UNICODE_STRING FullDllName;
            UNICODE_STRING BaseDllName;
            // ... remaining fields exist but are not needed for resolution
        };

        struct PRO_PEB_LDR_DATA {
            ULONG Length;
            BOOLEAN Initialized;
            HANDLE SsHandle;
            LIST_ENTRY InLoadOrderModuleList;
            LIST_ENTRY InMemoryOrderModuleList;
            LIST_ENTRY InInitializationOrderModuleList;
        };

        class Resolver {
        public:
            // Public Interface: Use NULL for local process, or a valid Handle for remote.
            static void* GetSafeModuleHandle(HANDLE h_process, const std::wstring& module_name);
            static void* GetSafeProcAddress(HANDLE h_process, void* module_base, const std::string& func_name);
            static void* LocalLoadLibrary(const std::string& module_name);
        private:
            // Internal logic using templates to support Local vs Remote memory access.
            template <typename Reader>
            static void* GetModuleHandleInternal(HANDLE h_process, const std::wstring& module_name);

            template <typename Reader>
            static void* GetProcAddressInternal(HANDLE h_process, void* module_base, const std::string& func_name);

            // Memory access policies
            struct LocalReader {
                static bool Read(HANDLE, void* addr, void* buffer, size_t size) {
                    if (!addr) return false;
                    memcpy(buffer, addr, size);
                    return true;
                }
            };

            struct RemoteReader {
                static bool Read(HANDLE h_process, void* addr, void* buffer, size_t size) {
                    SIZE_T bytes_read;
                    return ReadProcessMemory(h_process, addr, buffer, size, &bytes_read) && bytes_read == size;
                }
            };
        };

    }  // namespace utils
}  // namespace prohook

#endif  // PROHOOK_INCLUDE_UTILS_RESOLVER_H_