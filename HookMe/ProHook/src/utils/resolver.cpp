// Copyright 2026 HookMe Authors.
#include "ProHook/utils/resolver.h"
#include <stddef.h>
#include <memory>

namespace prohook {
    namespace utils {

        // --- Public Interface ---

        void* Resolver::GetSafeModuleHandle(HANDLE h_process, const std::wstring& module_name) {
            if (h_process == nullptr || h_process == GetCurrentProcess()) {
                return GetModuleHandleInternal<LocalReader>(GetCurrentProcess(), module_name);
            }
            return GetModuleHandleInternal<RemoteReader>(h_process, module_name);
        }

        void* Resolver::GetSafeProcAddress(HANDLE h_process, void* module_base, const std::string& func_name) {
            if (h_process == nullptr || h_process == GetCurrentProcess()) {
                return GetProcAddressInternal<LocalReader>(GetCurrentProcess(), module_base, func_name);
            }
            return GetProcAddressInternal<RemoteReader>(h_process, module_base, func_name);
        }

        void* Resolver::LocalLoadLibrary(const std::string& module_name) {
            return static_cast<void*>(LoadLibraryA(module_name.c_str()));
        }

        // --- Internal Template Logic ---

        template <typename Reader>
        void* Resolver::GetModuleHandleInternal(HANDLE h_process, const std::wstring& module_name) {
            PROCESS_BASIC_INFORMATION pbi;
            typedef NTSTATUS(NTAPI* pNtQueryInformationProcess)(HANDLE, PROCESSINFOCLASS, PVOID, ULONG, PULONG);

            // Dynamically resolve NtQueryInformationProcess for maximum reliability
            auto NtQueryInfo = (pNtQueryInformationProcess)GetProcAddress(GetModuleHandleW(L"ntdll.dll"), "NtQueryInformationProcess");
            if (!NtQueryInfo || NtQueryInfo(h_process, ProcessBasicInformation, &pbi, sizeof(pbi), nullptr) != 0)
                return nullptr;

            PEB peb_data;
            if (!Reader::Read(h_process, pbi.PebBaseAddress, &peb_data, sizeof(PEB))) return nullptr;

            PRO_PEB_LDR_DATA ldr_data;
            if (!Reader::Read(h_process, peb_data.Ldr, &ldr_data, sizeof(PRO_PEB_LDR_DATA))) return nullptr;

            // Walk the InLoadOrderModuleList (using the head of the list as the anchor)
            LIST_ENTRY* head = &ldr_data.InLoadOrderModuleList;
            LIST_ENTRY current_link;
            if (!Reader::Read(h_process, head->Flink, &current_link, sizeof(LIST_ENTRY))) return nullptr;

            // Iterate until we loop back to the head
            while (head->Flink != current_link.Flink) {
                PRO_LDR_DATA_TABLE_ENTRY entry;
                // InLoadOrderLinks is the FIRST member, so entry_addr == current_link address
                void* entry_addr = (BYTE*)current_link.Flink - offsetof(PRO_LDR_DATA_TABLE_ENTRY, InLoadOrderLinks);

                if (!Reader::Read(h_process, entry_addr, &entry, sizeof(PRO_LDR_DATA_TABLE_ENTRY))) break;

                if (entry.BaseDllName.Buffer) {
                    std::vector<wchar_t> name_buf(entry.BaseDllName.Length / sizeof(wchar_t) + 1, 0);
                    if (Reader::Read(h_process, entry.BaseDllName.Buffer, name_buf.data(), entry.BaseDllName.Length)) {
                        if (_wcsicmp(name_buf.data(), module_name.c_str()) == 0) {
                            return entry.DllBase;
                        }
                    }
                }

                // Move to the next link
                if (!Reader::Read(h_process, current_link.Flink, &current_link, sizeof(LIST_ENTRY))) break;
            }

            return nullptr;
        }

        template <typename Reader>
        void* Resolver::GetProcAddressInternal(HANDLE h_process, void* module_base, const std::string& func_name) {
            if (!module_base) return nullptr;

            IMAGE_DOS_HEADER dos_hdr;
            if (!Reader::Read(h_process, module_base, &dos_hdr, sizeof(dos_hdr))) return nullptr;

            // Check Signature and Bitness via Magic
            IMAGE_NT_HEADERS64 nt_hdrs;
            if (!Reader::Read(h_process, (BYTE*)module_base + dos_hdr.e_lfanew, &nt_hdrs, sizeof(nt_hdrs))) return nullptr;

            DWORD export_rva = 0;
            DWORD export_size = 0;

            if (nt_hdrs.OptionalHeader.Magic == IMAGE_NT_OPTIONAL_HDR64_MAGIC) {
                export_rva = nt_hdrs.OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT].VirtualAddress;
                export_size = nt_hdrs.OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT].Size;
            }
            else {
                IMAGE_NT_HEADERS32* nt32 = (IMAGE_NT_HEADERS32*)&nt_hdrs;
                export_rva = nt32->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT].VirtualAddress;
                export_size = nt32->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT].Size;
            }

            if (export_rva == 0) return nullptr;

            IMAGE_EXPORT_DIRECTORY exp_dir;
            if (!Reader::Read(h_process, (BYTE*)module_base + export_rva, &exp_dir, sizeof(exp_dir))) return nullptr;

            auto names = std::make_unique<DWORD[]>(exp_dir.NumberOfNames);
            auto ordinals = std::make_unique<WORD[]>(exp_dir.NumberOfNames);
            auto functions = std::make_unique<DWORD[]>(exp_dir.NumberOfFunctions);

            Reader::Read(h_process, (BYTE*)module_base + exp_dir.AddressOfNames, names.get(), sizeof(DWORD) * exp_dir.NumberOfNames);
            Reader::Read(h_process, (BYTE*)module_base + exp_dir.AddressOfNameOrdinals, ordinals.get(), sizeof(WORD) * exp_dir.NumberOfNames);
            Reader::Read(h_process, (BYTE*)module_base + exp_dir.AddressOfFunctions, functions.get(), sizeof(DWORD) * exp_dir.NumberOfFunctions);

            for (DWORD i = 0; i < exp_dir.NumberOfNames; ++i) {
                char name_buf[256] = { 0 };
                if (Reader::Read(h_process, (BYTE*)module_base + names[i], name_buf, sizeof(name_buf))) {
                    if (func_name == name_buf) {
                        DWORD func_rva = functions[ordinals[i]];

                        // Forwarder Check
                        if (func_rva >= export_rva && func_rva < (export_rva + export_size)) {
                            // Complexity: Remote forwarders require recursive module loading.
                            // For LoadLibraryW/Standard Injector, this is rarely hit.
                            return nullptr;
                        }
                        return (BYTE*)module_base + func_rva;
                    }
                }
            }

            return nullptr;
        }

    }  // namespace utils
}  // namespace prohook