// Copyright 2026 HookMe Authors.
#ifndef PROHOOK_INCLUDE_INJECTORS_STANDARD_INJECTOR_H_
#define PROHOOK_INCLUDE_INJECTORS_STANDARD_INJECTOR_H_
#include "ProHook/injectors/i_injector.h"

#include <filesystem>

#include <windows.h>

namespace prohook::injectors {

    class StandardInjector : public IInjector {
    public:
        StandardInjector() = default;
        ~StandardInjector() override = default;

        bool Inject(unsigned long process_id,
            const std::filesystem::path& dll_path) override;

        // We will leave this for a future iteration when we do Shellcode/Reflective.
        bool InjectRaw(unsigned long process_id,
            const std::vector<unsigned char>& payload) override {
            return false;
        }
    };

}  // namespace prohook::injectors

#endif  // PROHOOK_INCLUDE_INJECTORS_STANDARD_INJECTOR_H_