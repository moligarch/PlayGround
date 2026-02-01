// Copyright 2026 HookMe Authors.
#ifndef PROHOOK_INCLUDE_INJECTORS_I_INJECTOR_H_
#define PROHOOK_INCLUDE_INJECTORS_I_INJECTOR_H_

#include <filesystem>
#include <vector>

namespace prohook::injectors {

    // Interface for delivery mechanisms to get our code into target processes.
    class IInjector {
    public:
        virtual ~IInjector() = default;

        // Injects a library into a remote process via PID.
        virtual bool Inject(unsigned long process_id,
            const std::filesystem::path& dll_path) = 0;

        // Injects raw shellcode or a reflected image into a remote process.
        virtual bool InjectRaw(unsigned long process_id,
            const std::vector<unsigned char>& payload) = 0;
    };

}  // namespace prohook::injectors

#endif  // PROHOOK_INCLUDE_INJECTORS_I_INJECTOR_H_