// Copyright 2026 HookMe Authors.
#include "core/resolver.h"

#include <Windows.h>

namespace prohook {
    namespace core {

        HMODULE Resolver::SafeLoadLibrary(const std::string& module_name) {
            // Currently wraps standard API. 
            // Future: Replace with manual map or LdrLoadDll bypass.
            return LoadLibraryA(module_name.c_str());
        }

        void* Resolver::SafeGetProcAddress(HMODULE module_base, const std::string& func_name) {
            // We use our manual internal parser instead of GetProcAddress.
            return reinterpret_cast<void*>(GetProcAddress(module_base, func_name.c_str()));
        }

        void* Resolver::GetExportAddressInternal(void* module_base, const std::string& func_name) {
            if (!module_base) return nullptr;

            auto* dos_header = reinterpret_cast<PIMAGE_DOS_HEADER>(module_base);
            auto* nt_headers = reinterpret_cast<PIMAGE_NT_HEADERS>(
                reinterpret_cast<BYTE*>(module_base) + dos_header->e_lfanew);

            IMAGE_DATA_DIRECTORY& export_dir_info = nt_headers->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT];
            DWORD export_dir_rva = export_dir_info.VirtualAddress;
            DWORD export_dir_size = export_dir_info.Size;

            if (export_dir_rva == 0) return nullptr;

            auto* export_dir = reinterpret_cast<PIMAGE_EXPORT_DIRECTORY>(
                reinterpret_cast<BYTE*>(module_base) + export_dir_rva);

            auto* names = reinterpret_cast<DWORD*>(reinterpret_cast<BYTE*>(module_base) + export_dir->AddressOfNames);
            auto* ordinals = reinterpret_cast<WORD*>(reinterpret_cast<BYTE*>(module_base) + export_dir->AddressOfNameOrdinals);
            auto* functions = reinterpret_cast<DWORD*>(reinterpret_cast<BYTE*>(module_base) + export_dir->AddressOfFunctions);

            for (DWORD i = 0; i < export_dir->NumberOfNames; ++i) {
                const char* current_name = reinterpret_cast<const char*>(reinterpret_cast<BYTE*>(module_base) + names[i]);
                if (func_name == current_name) {
                    WORD ordinal = ordinals[i];
                    DWORD func_rva = functions[ordinal];

                    // --- THE FORWARDER CHECK ---
                    // If the RVA is inside the Export Directory's memory range, it's a forwarder string.
                    if (func_rva >= export_dir_rva && func_rva < (export_dir_rva + export_dir_size)) {
                        const char* forwarder_str = reinterpret_cast<const char*>(reinterpret_cast<BYTE*>(module_base) + func_rva);
                        return ResolveForwarder(forwarder_str);
                    }

                    return reinterpret_cast<void*>(reinterpret_cast<BYTE*>(module_base) + func_rva);
                }
            }
            return nullptr;
        }

        void* Resolver::ResolveForwarder(const char* forwarder_str) {
            // forwarder_str is usually "DLLNAME.FunctionName" (e.g. "NTDLL.RtlAllocateHeap")
            std::string s(forwarder_str);
            size_t dot = s.find('.');
            if (dot == std::string::npos) return nullptr;

            std::string dll = s.substr(0, dot) + ".dll";
            std::string func = s.substr(dot + 1);

            HMODULE h_mod = GetModuleHandleA(dll.c_str());
            if (!h_mod) h_mod = SafeLoadLibrary(dll);

            return SafeGetProcAddress(h_mod, func);
        }

    }  // namespace core
}  // namespace prohook