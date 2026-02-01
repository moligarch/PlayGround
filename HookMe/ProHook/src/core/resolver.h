// Copyright 2026 HookMe Authors.
#ifndef PROHOOK_SRC_CORE_RESOLVER_H_
#define PROHOOK_SRC_CORE_RESOLVER_H_

#include <string>
#include <windows.h>

namespace prohook {
	namespace core {

		class Resolver {
		public:
			// High-level wrapper for loading a module.
			static HMODULE SafeLoadLibrary(const std::string& module_name);

			// High-level wrapper for getting an export address.
			static void* SafeGetProcAddress(HMODULE module_base, const std::string& func_name);

			// Internal: The manual PE parser.
			static void* GetExportAddressInternal(void* module_base, const std::string& func_name);

		private:
			// Internal: Handles the case where an export points to another DLL.
			static void* ResolveForwarder(const char* forwarder_str);
		};

	}  // namespace core
}  // namespace prohook

#endif  // PROHOOK_SRC_CORE_RESOLVER_H_