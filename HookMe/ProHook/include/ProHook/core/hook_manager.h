// Copyright 2026 HookMe Authors.
#ifndef PROHOOK_INCLUDE_CORE_HOOK_MANAGER_H_
#define PROHOOK_INCLUDE_CORE_HOOK_MANAGER_H_

#include <memory>
#include <string>
#include <vector>

#include "ProHook/hooks/i_hook.h"

namespace prohook::core {

	struct HookContext {
		std::string module_name;
		std::string function_name;
		void* proxy_func;
		void** original_func;
	};

	// Manages the orchestration of hooks and resolution of symbols.
	class HookManager {
	public:
		static HookManager& Instance();

		// Disallow copy and assign.
		HookManager(const HookManager&) = delete;
		HookManager& operator=(const HookManager&) = delete;

		// Sets the underlying hooking implementation (e.g., Detours).
		void SetHookProvider(std::unique_ptr<hooks::IHook> provider);

		// Registers a hook to be applied.
		// target_spec format: "module.dll!FunctionName"
		void AddHook(const std::string& target_spec, void* proxy, void** original);

		// Iterates through registered hooks, resolves addresses, and installs them.
		bool DeployAll();

		// Cleanly removes all active hooks.
		void Teardown();

	private:
		HookManager() = default;

		std::unique_ptr<hooks::IHook> provider_;
		std::vector<HookContext> pending_hooks_;
	};

}  // namespace prohook::core

#endif  // PROHOOK_INCLUDE_CORE_HOOK_MANAGER_H_